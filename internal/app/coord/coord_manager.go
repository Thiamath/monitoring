/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Coord implementation for CoordManager

package coord

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-app-cluster-api-go"
	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-infrastructure-monitor-go"

	"github.com/rs/zerolog/log"

	"google.golang.org/grpc"
)

type CoordManager struct {
	clustersClient grpc_infrastructure_go.ClustersClient
	params *AppClusterConnectParams
}

type AppClusterConnectParams struct {
	AppClusterPrefix string
	AppClusterPort int
	UseTLS bool
	CACert string
	Insecure bool
}

type clusterClient struct {
	grpc_app_cluster_api_go.InfrastructureMonitorClient
	conn *grpc.ClientConn
}

func NewClusterClient(address string, params *AppClusterConnectParams) (*clusterClient, derrors.Error) {
	return nil, nil
}

func (c *clusterClient) Close() error {
	err := c.conn.Close()
	if err != nil {
		log.Warn().Msg("error closing client connection")
	}

	return err
}

// Create a new query manager.
func NewCoordManager(clustersClient grpc_infrastructure_go.ClustersClient, params *AppClusterConnectParams) (*CoordManager, derrors.Error) {
	manager := &CoordManager{
		clustersClient: clustersClient,
		params: params,
	}

	return manager, nil
}

// TODO:
// For all:
// - find cluster - check org id
//     rpc GetCluster(ClusterId) returns (Cluster) {}

// - make client
// - connect
// - forward request
// - close connection

func (m *CoordManager) getClusterClient(organizationId, clusterId string) (*clusterClient, derrors.Error) {
	return nil, nil
}

// Retrieve a summary of high level cluster resource availability
func (m *CoordManager) GetClusterSummary(ctx context.Context, request *grpc_infrastructure_monitor_go.ClusterSummaryRequest) (*grpc_infrastructure_monitor_go.ClusterSummary, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}

	res, err := client.GetClusterSummary(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing GetClusterSummary on cluster", err)
	}

	client.Close()
	return res, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (m *CoordManager) GetClusterStats(ctx context.Context, request *grpc_infrastructure_monitor_go.ClusterStatsRequest) (*grpc_infrastructure_monitor_go.ClusterStats, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}

	res, err := client.GetClusterStats(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing GetClusterStats on cluster", err)
	}

	client.Close()
	return res, nil
}

// Execute a query directly on the monitoring storage backend
func (m *CoordManager) Query(ctx context.Context, request *grpc_infrastructure_monitor_go.QueryRequest) (*grpc_infrastructure_monitor_go.QueryResponse, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}

	res, err := client.Query(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing Query on cluster", err)
	}

	client.Close()
	return res, nil
}
