/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package monitoring_manager

import (
	"fmt"
	"net"

	"github.com/nalej/derrors"

        "github.com/nalej/grpc-infrastructure-go"
        "github.com/nalej/grpc-monitoring-go"

	"github.com/nalej/monitoring/internal/pkg/retrieve"

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
	// Create system model connection
	smConn, err := grpc.Dial(s.Configuration.SystemModelAddress, grpc.WithInsecure())
	if err != nil {
		return derrors.NewUnavailableError("cannot create connection with the system model", err)
	}

	// Create clients
	clustersClient := grpc_infrastructure_go.NewClustersClient(smConn)

	// Start listening
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	// Create managers and handler
	params := &AppClusterConnectParams{
		AppClusterPrefix: s.Configuration.AppClusterPrefix,
		AppClusterPort: s.Configuration.AppClusterPort,
		UseTLS: s.Configuration.UseTLS,
		CACert: s.Configuration.CACert,
		Insecure: s.Configuration.Insecure,
	}

	manager, derr := NewManager(clustersClient, params)
	if derr != nil {
		return derr
	}
	handler, derr := retrieve.NewHandler(manager)
	if derr != nil {
		return derr
	}

	// Create server and register handler
	server := grpc.NewServer()
	grpc_monitoring_go.RegisterMonitoringManagerServer(server, handler)

	reflection.Register(server)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := server.Serve(lis); err != nil {
		return derrors.NewUnavailableError("failed to serve", err)
	}

	return nil
}
