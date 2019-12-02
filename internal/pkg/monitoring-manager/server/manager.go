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

package server

import (
	"context"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/monitoring/internal/pkg/monitoring-manager/clients"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-monitoring-go"
)

type Manager struct {
	clustersClient grpc_infrastructure_go.ClustersClient
	params         *clients.AppClusterConnectParams
}

// Create a new query manager.
func NewManager(clustersClient grpc_infrastructure_go.ClustersClient, params *clients.AppClusterConnectParams) (Manager, derrors.Error) {
	manager := Manager{
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
func (m *Manager) GetClusterSummary(ctx context.Context, request *grpc_monitoring_go.ClusterSummaryRequest) (*grpc_monitoring_go.ClusterSummary, error) {
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
func (m *Manager) GetClusterStats(ctx context.Context, request *grpc_monitoring_go.ClusterStatsRequest) (*grpc_monitoring_go.ClusterStats, error) {
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
func (m *Manager) Query(ctx context.Context, request *grpc_monitoring_go.QueryRequest) (*grpc_monitoring_go.QueryResponse, error) {
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

func (m *Manager) GetOrganizationApplicationStats(ctx context.Context, request *grpc_monitoring_go.OrganizationApplicationStatsRequest) (*grpc_monitoring_go.OrganizationApplicationStatsResponse, error) {
	clusterList, err := m.clustersClient.ListClusters(ctx, &grpc_organization_go.OrganizationId{OrganizationId: request.OrganizationId})
	if err != nil {
		return nil, derrors.NewFailedPreconditionError("could not find cluster list", err)
	}

	serviceInstanceStats := make([]*grpc_monitoring_go.OrganizationApplicationStats, 0)
	orgAppStats := &grpc_monitoring_go.OrganizationApplicationStatsResponse{
		ServiceInstanceStats: serviceInstanceStats,
		Timestamp:            0,
	}
	for _, cluster := range clusterList.Clusters {
		client, derr := m.getClusterClient(request.OrganizationId, cluster.ClusterId)
		if derr != nil {
			// TODO what to do if a cluster client cannot be created
			continue
		}
		containerStats, err := client.GetContainerStats(ctx, nil)
		if err != nil {
			// TODO what to do if cluster do not respond with stats
			continue
		}
		for _, stats := range containerStats.ContainerStats {
			serviceInstanceStats = append(serviceInstanceStats, &grpc_monitoring_go.OrganizationApplicationStats{
				OrganizationId: request.OrganizationId,
				//OrganizationName:         organizationName, TODO
				//AppInstanceId:            aooInstanceId, TODO
				//AppInstanceName:          appInstanceNAme, TODO
				//ServiceGroupInstanceId:   serviceGroupInstanceId, TODO
				//ServiceGroupInstanceName: serviceGroupInstanceName, TODO
				//ServiceInstanceId:        serviceInstanceId, TODO
				//ServiceInstanceName:      serviceInstanceName, TODO
				CpuMillicore: stats.CpuMillicore * cluster.MillicoresConversionFactor,
				MemoryByte:   stats.MemoryByte,
				StorageByte:  stats.StorageByte,
			})
		}
	}

	return orgAppStats, nil
}
