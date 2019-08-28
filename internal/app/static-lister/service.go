/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package static_lister

// Static lister app provides a Prometheus metrics endpoint that returns
// a static metric series with specified name and a label with the values
// read from label-file. We monitor label-file and change the values
// accordingly.
// This is needed to be able to detect that a certain component
// does not exist anymore at all (e.g., deployment deleted); no metrics
// would be available from kube-state-metrics, so we need a "master list"
// of components we expect. With this static-lister, we can change the
// master list easily by changing the label file (which is likely to be
// a config map), and we can have different expected components per cluster.
// Local Prometheus will scrape an instance of static-lister for each
// resource type (deployment, daemonset, statefulset), and this information
// will get exported to the central availability monitoring system where
// we can compare it with component information originating from
// kube-state-metrics.

// Example:
// $ cat /tmp/foo
// abc
// def
// ghi
// $ ./static-lister run --name deployments --label-name deployment --label-file /tmp/foo
// $ curl http://localhost:9001/metrics
// # HELP nalej_components_deployments nalej components deployments that should be available
// # TYPE nalej_components_deployments gauge
// nalej_components_deployments{deployment="abc"} 1
// nalej_components_deployments{deployment="def"} 1
// nalej_components_deployments{deployment="ghi"} 1
//
// With this information in Prometheus, we can use it in the following query:
//
// kube_deployment_status_replicas_available{namespace="nalej"} / kube_deployment_spec_replicas or on(deployment) (nalej_components_deployments == bool 0)
//
// This will calculate the ratio between available and requested replicas for a
// deployment (anything smaller than 1 means there is some degradation). We then
// union this with any metrics from the component list that isn't present
// in the ratio list (that is what "or" does), adding the "== bool 0" to flip
// the 1 into a 0. This way, non-existing deployments that we expected to exist
// will have a ratio of availability of 0 (or, complete degradation and no
// pods running).

// We use the standard Prometheus client code and handler to create the metric
// that results in the above output.

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nalej/derrors"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog/log"
)

// Service with configuration and gRPC server
type Service struct {
	Configuration *Config
}

func NewService(conf *Config) (*Service, derrors.Error) {
	err := conf.Validate()
	if err != nil {
		log.Error().Msg("Invalid configuration")
		return nil, err
	}
	conf.Print()

	return &Service{
		Configuration: conf,
	}, nil
}

// Run the service, launch the REST service handler.
func (s *Service) Run() derrors.Error {
	// Channel to signal errors from starting the servers
	errChan := make(chan error, 1)

	// Listen on metrics port
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	httpServer, derr := s.startMetrics(httpListener, errChan)
	if derr != nil {
		return derr
	}
	defer httpServer.Shutdown(context.TODO()) // Add timeout in context

	// Wait for termination signal
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM)
	signal.Notify(sigterm, syscall.SIGINT)

	select {
	case sig := <-sigterm:
		log.Info().Str("signal", sig.String()).Msg("Gracefully shutting down")
	case err := <-errChan:
		// We've already logged the error
		return derrors.NewInternalError("failed starting server", err)
	}

	return nil
}

// Initialize and start the metrics endpoint
// This starts the HTTP server providing the "/metrics" endpoint.
func (s *Service) startMetrics(httpListener net.Listener, errChan chan<- error) (*http.Server, derrors.Error) {
	// Create empty prometheus registry to not expose all default
	// internal measurements
	registry := prometheus.NewRegistry()
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})

	// Register series
	gauge := prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: s.Configuration.Namespace,
		Subsystem: s.Configuration.Subsystem,
		Name: s.Configuration.Name,
		Help: fmt.Sprintf("%s %s %s that should be available", s.Configuration.Namespace, s.Configuration.Subsystem, s.Configuration.Name),
	}, []string{s.Configuration.LabelName})
	err := registry.Register(gauge)
	if err != nil {
		return nil, derrors.NewInternalError("unable to register gauge with prometheus", err)
	}

	// Create value file watcher - fills the gauge vector and watches for
	// changes in the label file
	watcher, err := NewWatcher(s.Configuration.LabelFile, gauge)
	if err != nil {
		return nil, derrors.NewInternalError("unable to create watcher", err)
	}
	go watcher.Run(errChan)

	// Create server with metrics handler
	http.Handle("/metrics", handler)
	httpServer := &http.Server{}

	// Start HTTP server
	log.Info().Int("port", s.Configuration.Port).Msg("Launching HTTP server")
	go func() {
		err := httpServer.Serve(httpListener)
		if err == http.ErrServerClosed {
			log.Info().Err(err).Msg("closed http server")
		} else if err != nil {
			log.Error().Err(err).Msg("failed to serve http")
			errChan <- err
		}
	}()

	return httpServer, nil
}
