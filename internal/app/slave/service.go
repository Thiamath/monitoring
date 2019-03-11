/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package slave

import (
	"fmt"
	"net"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/internal/pkg/collect"
	"github.com/nalej/infrastructure-monitor/pkg/metrics"
	"github.com/nalej/infrastructure-monitor/pkg/provider/events/kubernetes"
	"github.com/nalej/infrastructure-monitor/pkg/provider/metrics/prometheus"

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

	// Create collector
	collector, derr := metrics.NewSimpleCollector()
	if derr != nil {
		return derr
	}

	// Create Kubernetes event collector provider
	kubeEvents, derr := kubernetes.NewEventsProvider(s.Configuration.Kubeconfig, s.Configuration.InCluster, collector)
	if derr != nil {
		return derr
	}

	// Create metrics endpoint provider
	promMetrics, derr := prometheus.NewMetricsProvider()
	if derr != nil {
		return derr
	}

	// Create managers and handler
	// Events collector and Metrics HTTP endpoint
	collectManager, derr := collect.NewManager(kubeEvents, promMetrics)
	if derr != nil {
		return derr
	}
	collectHandler, derr := collect.NewHandler(collectManager)
	if derr != nil {
		return derr
	}

	// Query gRPC endpoints
	// TBD

	// Create server and register handler
	server := grpc.NewServer()
	// TBD: register handler

	// Start managers
	derr = collectManager.Start()
	if derr != nil {
		return derr
	}

	// Listen on gRPC port
	reflection.Register(server)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := server.Serve(lis); err != nil {
		return derrors.NewUnavailableError("failed to serve", err)
	}

	// Listen on HTTP port
	// TBD
	_ = collectHandler

	return nil
}
