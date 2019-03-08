/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package slave

import (
	"fmt"
	"net"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/internal/pkg/metrics"
	"github.com/nalej/infrastructure-monitor/pkg/provider/collector/kubernetes"

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
	// Start listening
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	// Create Kubernetes Collector Provider
	kubeCollector, derr := kubernetes.NewCollectorProvider(s.Configuration.Kubeconfig, s.Configuration.InCluster)
	if derr != nil {
		return derr
	}

	// Create managers and handler
	metricsManager, derr := metrics.NewManager(kubeCollector)
	if derr != nil {
		return derr
	}
	// TBD create handler

	// Create server and register handler
	server := grpc.NewServer()
	// TBD: register handler

	// Start managers
	derr = metricsManager.Start()
	if derr != nil {
		return derr
	}

	reflection.Register(server)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := server.Serve(lis); err != nil {
		return derrors.NewUnavailableError("failed to serve", err)
	}

	return nil
}
