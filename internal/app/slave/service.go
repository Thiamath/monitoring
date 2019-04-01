/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package slave

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-infrastructure-monitor-go"

	"github.com/nalej/deployment-manager/pkg/utils"
	"github.com/nalej/infrastructure-monitor/internal/pkg/collect"
	"github.com/nalej/infrastructure-monitor/internal/pkg/retrieve"
	"github.com/nalej/infrastructure-monitor/pkg/provider/events/kubernetes"
	metrics_prometheus "github.com/nalej/infrastructure-monitor/pkg/provider/metrics/prometheus"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	httpServer, err := s.startCollect(errChan)
	if err != nil {
		return err
	}
	defer httpServer.Shutdown(context.TODO()) // Add timeout in context

	grpcServer, err := s.startRetrieve(errChan)
	if err != nil {
		return err
	}
	defer grpcServer.GracefulStop()

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

// Initialize and start the collecting of metrics through events
// This starts the HTTP server providing the "/metrics" endpoint.
func (s *Service) startCollect(errChan chan<- error) (*http.Server, derrors.Error) {
	// Listen on metrics port
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.MetricsPort))
	if err != nil {
		return nil, derrors.NewUnavailableError("failed to listen", err)
	}

	// Create metrics endpoint provider
	promMetrics, derr := metrics_prometheus.NewMetricsProvider()
	if derr != nil {
		return nil, derr
	}

	// Create Kubernetes event collector provider
	labelSelector := utils.NALEJ_ANNOTATION_ORGANIZATION // only get events relevant for user applications
	kubeEvents, derr := kubernetes.NewEventsProvider(s.Configuration.Kubeconfig, s.Configuration.InCluster,
		labelSelector, promMetrics.GetCollector())
	if derr != nil {
		return nil, derr
	}

	// Create managers and handler
	// Events collector and Metrics HTTP endpoint
	collectManager, derr := collect.NewManager(kubeEvents, promMetrics)
	if derr != nil {
		return nil, derr
	}
	collectHandler, derr := collect.NewHandler(collectManager)
	if derr != nil {
		return nil, derr
	}

	// Create server with metrics handler
	httpServer := &http.Server{
		Handler: collectHandler,
	}

	// Start manager
	derr = collectManager.Start()
	if derr != nil {
		return nil, derr
	}

	// Start HTTP server
	log.Info().Int("port", s.Configuration.MetricsPort).Msg("Launching HTTP server")
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

// Initialize and start the retrieval/query API.
// This starts the gRPC server.
func (s *Service) startRetrieve(errChan chan<- error) (*grpc.Server, derrors.Error) {
	// Start listening on API port
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return nil, derrors.NewUnavailableError("failed to listen", err)
	}

	// Create query providers
	queryProviders := query.QueryProviders{}
	for queryProviderType, queryProviderConfig := range(s.Configuration.QueryProviders) {
		if queryProviderConfig.Enabled() {
			queryProvider, derr := queryProviderConfig.NewProvider()
			if derr != nil {
				return nil, derr
			}
			queryProviders[queryProviderType] = queryProvider
		}
	}

	// Create manager and handler for gRPC endpoints
	retrieveManager, derr := NewRetrieveManager(queryProviders)
	if derr != nil {
		return nil, derr
	}
	retrieveHandler, derr := retrieve.NewHandler(retrieveManager)
	if derr != nil {
		return nil, derr
	}

	// Create server and register handler
	grpcServer := grpc.NewServer()
	grpc_infrastructure_monitor_go.RegisterSlaveServer(grpcServer, retrieveHandler)

	// Start gRPC server
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	go func() {
		if err := grpcServer.Serve(grpcListener); err != nil {
			log.Error().Err(err).Msg("failed to serve grpc")
			errChan <- err
		}
		log.Info().Msg("closed grpc server")
	}()

	return grpcServer, nil
}
