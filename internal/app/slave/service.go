/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package slave

import (
	"fmt"
	"net"
	"net/http"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/internal/pkg/collect"
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
	grpcListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}
	httpListener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.MetricsPort))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	// Create metrics endpoint provider
	promMetrics, derr := prometheus.NewMetricsProvider()
	if derr != nil {
		return derr
	}

	// Create Kubernetes event collector provider
	kubeEvents, derr := kubernetes.NewEventsProvider(s.Configuration.Kubeconfig, s.Configuration.InCluster, promMetrics.GetCollector())
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
	grpcServer := grpc.NewServer()
	httpServer := http.Server{}

	http.HandleFunc("/metrics", collectHandler.Metrics)

	// Start managers
	derr = collectManager.Start()
	if derr != nil {
		return derr
	}

	// Listen on HTTP port
	log.Info().Int("port", s.Configuration.MetricsPort).Msg("Launching HTTP server")
	go httpServer.Serve(httpListener)

	// Listen on gRPC port
	reflection.Register(grpcServer)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := grpcServer.Serve(grpcListener); err != nil {
		return derrors.NewUnavailableError("failed to serve", err)
	}


	return nil
}
