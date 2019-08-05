/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Manager implementation for cluster monitoring

package cluster

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-monitoring-go"
)

type Manager struct {
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

// Create a new query manager.
func NewManager(clustersClient grpc_infrastructure_go.ClustersClient, params *AppClusterConnectParams) (*Manager, derrors.Error) {
	manager := &Manager{
		clustersClient: clustersClient,
		params: params,
	}

	return manager, nil
}

func (m *Manager) getClusterClient(organizationId, clusterId string) (*clusterClient, derrors.Error) {
	getClusterRequest := &grpc_infrastructure_go.ClusterId{
		OrganizationId: organizationId,
		ClusterId: clusterId,
	}

	cluster, err := m.clustersClient.GetCluster(context.Background(), getClusterRequest)
	if err != nil || cluster == nil {
		return nil, derrors.NewUnavailableError("unable to retrieve cluster", err)
	}

	return NewClusterClient(cluster.GetHostname(), m.params)
}

// Retrieve a summary of high level cluster resource availability
func (m *Manager) GetClusterSummary(ctx context.Context, request *grpc_monitoring_go.ClusterSummaryRequest) (*grpc_monitoring_go.ClusterSummary, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}
	defer client.Close()

	res, err := client.GetClusterSummary(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing GetClusterSummary on cluster", err)
	}

	return res, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (m *Manager) GetClusterStats(ctx context.Context, request *grpc_monitoring_go.ClusterStatsRequest) (*grpc_monitoring_go.ClusterStats, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}
	defer client.Close()

	res, err := client.GetClusterStats(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing GetClusterStats on cluster", err)
	}

	return res, nil
}

// Execute a query directly on the monitoring storage backend
func (m *Manager) Query(ctx context.Context, request *grpc_monitoring_go.QueryRequest) (*grpc_monitoring_go.QueryResponse, derrors.Error) {
	client, derr := m.getClusterClient(request.GetOrganizationId(), request.GetClusterId())
	if derr != nil {
		return nil, derr
	}
	defer client.Close()

	res, err := client.Query(ctx, request)
	if err != nil {
		return nil, derrors.NewUnavailableError("error executing Query on cluster", err)
	}

	return res, nil
}
