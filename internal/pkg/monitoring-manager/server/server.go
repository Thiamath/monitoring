/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package server

import (
	"fmt"
	grpc_organization_go "github.com/nalej/grpc-organization-go"
	"github.com/nalej/monitoring/internal/pkg/monitoring-manager/clients"
	"github.com/nalej/monitoring/internal/pkg/monitoring-manager/server/asset"
	"net"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-inventory-proxy-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/grpc-monitoring-go"

	"github.com/rs/zerolog/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
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

	// Create Edge Inventory Proxy connection
	eipConn, err := grpc.Dial(s.Configuration.EdgeInventoryProxyAddress, grpc.WithInsecure())
	if err != nil {
		return derrors.NewUnavailableError("cannot create connection with the edge inventory proxy", err)
	}

	// Create clients
	clustersClient := grpc_infrastructure_go.NewClustersClient(smConn)
	organizationsClient := grpc_organization_go.NewOrganizationsClient(smConn)
	assetsClient := grpc_inventory_go.NewAssetsClient(smConn)
	controllersClient := grpc_inventory_go.NewControllersClient(smConn)
	eipClient := grpc_edge_inventory_proxy_go.NewEdgeControllerProxyClient(eipConn)

	// Start listening
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Configuration.Port))
	if err != nil {
		return derrors.NewUnavailableError("failed to listen", err)
	}

	// Create managers and handler
	params := &clients.AppClusterConnectParams{
		AppClusterPrefix:         s.Configuration.AppClusterPrefix,
		AppClusterPort:           s.Configuration.AppClusterPort,
		UseTLS:                   s.Configuration.UseTLS,
		CACertPath:               s.Configuration.CACertPath,
		ClientCertPath:           s.Configuration.ClientCertPath,
		SkipServerCertValidation: s.Configuration.SkipServerCertValidation,
	}

	// Cluster monitoring
	clusterManager, derr := NewManager(&clustersClient, &organizationsClient, params)
	if derr != nil {
		return derr
	}
	clusterHandler, derr := NewHandler(clusterManager, s.Configuration.CacheTTL)
	if derr != nil {
		return derr
	}

	// Asset monitoring
	assetManager, derr := asset.NewManager(eipClient, assetsClient, controllersClient)
	if derr != nil {
		return derr
	}
	assetHandler, derr := asset.NewHandler(assetManager)

	// Create server and register handler
	server := grpc.NewServer()
	grpc_monitoring_go.RegisterMonitoringManagerServer(server, clusterHandler)
	grpc_monitoring_go.RegisterAssetMonitoringServer(server, assetHandler)

	reflection.Register(server)
	log.Info().Int("port", s.Configuration.Port).Msg("Launching gRPC server")
	if err := server.Serve(lis); err != nil {
		return derrors.NewUnavailableError("failed to serve", err)
	}

	return nil
}
