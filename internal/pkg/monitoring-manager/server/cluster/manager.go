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

// Manager implementation for cluster monitoring

package cluster

import (
	"context"
	"github.com/nalej/monitoring/internal/pkg/monitoring-manager/clients"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-monitoring-go"
)

type Manager struct {
	clustersClient grpc_infrastructure_go.ClustersClient
	params         *AppClusterConnectParams
}

type AppClusterConnectParams struct {
	AppClusterPrefix         string
	AppClusterPort           int
	UseTLS                   bool
	CACertPath               string
	ClientCertPath           string
	SkipServerCertValidation bool
}

// Create a new query manager.
func NewManager(clustersClient grpc_infrastructure_go.ClustersClient, params *AppClusterConnectParams) (*Manager, derrors.Error) {
	manager := &Manager{
		clustersClient: clustersClient,
		params:         params,
	}

	return manager, nil
}

func (m *Manager) getClusterClient(organizationId, clusterId string) (*clients.ClusterClient, derrors.Error) {
	getClusterRequest := &grpc_infrastructure_go.ClusterId{
		OrganizationId: organizationId,
		ClusterId:      clusterId,
	}

	cluster, err := m.clustersClient.GetCluster(context.Background(), getClusterRequest)
	if err != nil || cluster == nil {
		return nil, derrors.NewUnavailableError("unable to retrieve cluster", err)
	}

	return clients.NewClusterClient(cluster.GetHostname(), m.params)
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
