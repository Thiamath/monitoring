/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package static_lister

// Static lister app provides a Prometheus metrics endpoint that returns
// a static metric series with specified name and a label with the values
// read from label-file. We monitor label-file and change the values
// accordingly.

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
