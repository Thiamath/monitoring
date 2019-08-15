/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics_collector

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-monitoring-go"

	"github.com/nalej/monitoring/internal/pkg/retrieve"
	"github.com/nalej/monitoring/pkg/provider/query"

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

	// Start listening on API port
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	grpcServer, derr := s.startRetrieve(grpcListener, errChan)
	if derr != nil {
		return derr
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

// Initialize and start the retrieval/query API.
// This starts the gRPC server.
func (s *Service) startRetrieve(grpcListener net.Listener, errChan chan<- error) (*grpc.Server, derrors.Error) {
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
	grpc_monitoring_go.RegisterMetricsCollectorServer(grpcServer, retrieveHandler)

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
