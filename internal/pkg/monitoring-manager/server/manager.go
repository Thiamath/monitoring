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
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/grpc-organization-go"
	"github.com/nalej/monitoring/internal/pkg/monitoring-manager/clients"
	"github.com/rs/zerolog/log"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-infrastructure-go"
	"github.com/nalej/grpc-monitoring-go"
)

const (
	defaultTimeout = 10 * time.Second
)

type Manager struct {
	clustersClient      *grpc_infrastructure_go.ClustersClient
	organizationsClient *grpc_organization_go.OrganizationsClient

	params *clients.AppClusterConnectParams
}

func (m *Manager) getOrganizationsClient() grpc_organization_go.OrganizationsClient {
	return *m.organizationsClient
}

func (m *Manager) getClustersClient() grpc_infrastructure_go.ClustersClient {
	return *m.clustersClient
}

// Create a new query manager.
func NewManager(clustersClient *grpc_infrastructure_go.ClustersClient, organizationsClient *grpc_organization_go.OrganizationsClient, params *clients.AppClusterConnectParams) (Manager, derrors.Error) {
	manager := Manager{
		clustersClient:      clustersClient,
		organizationsClient: organizationsClient,
		params:              params,
	}

	return manager, nil
}

func (m *Manager) getMetricsCollectorClient(organizationId, clusterId string) (*clients.MetricsCollectorClient, derrors.Error) {
	getClusterRequest := &grpc_infrastructure_go.ClusterId{
		OrganizationId: organizationId,
		ClusterId:      clusterId,
	}

	cluster, err := m.getClustersClient().GetCluster(context.Background(), getClusterRequest)
	if err != nil || cluster == nil {
		return nil, derrors.NewUnavailableError("unable to retrieve cluster", err)
	}

	return clients.NewMetricsCollectorClient(cluster.GetHostname(), m.params)
}

// Retrieve a summary of high level cluster resource availability
func (m *Manager) GetClusterSummary(ctx context.Context, request *grpc_monitoring_go.ClusterSummaryRequest) (*grpc_monitoring_go.ClusterSummary, error) {
	client, derr := m.getMetricsCollectorClient(request.GetOrganizationId(), request.GetClusterId())
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
	client, derr := m.getMetricsCollectorClient(request.GetOrganizationId(), request.GetClusterId())
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
	client, derr := m.getMetricsCollectorClient(request.GetOrganizationId(), request.GetClusterId())
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
	getOrganizationCtx, getOrganizationCancel := context.WithTimeout(ctx, defaultTimeout)
	defer getOrganizationCancel()
	organization, err := m.getOrganizationsClient().GetOrganization(getOrganizationCtx, &grpc_organization_go.OrganizationId{OrganizationId: request.OrganizationId})
	if err != nil {
		return nil, derrors.NewFailedPreconditionError("could not get organization", err)
	}

	listClustersCtx, listclustersCancel := context.WithTimeout(ctx, defaultTimeout)
	defer listclustersCancel()
	clusterList, err := m.getClustersClient().ListClusters(listClustersCtx, &grpc_organization_go.OrganizationId{OrganizationId: request.OrganizationId})
	if err != nil {
		return nil, derrors.NewFailedPreconditionError("could not geet cluster list", err)
	}

	orgContainerStats := m.requestContainerStatsToClusters(clusterList, organization, ctx)

	serviceInstanceStats := m.aggregateStatsByServiceInstanceId(orgContainerStats, request, organization.Name)

	orgAppStats := &grpc_monitoring_go.OrganizationApplicationStatsResponse{
		ServiceInstanceStats: serviceInstanceStats,
		Timestamp:            time.Now().Unix(),
	}

	return orgAppStats, nil
}

func (m *Manager) requestContainerStatsToClusters(clusterList *grpc_infrastructure_go.ClusterList, organization *grpc_organization_go.Organization, ctx context.Context) []*grpc_monitoring_go.ContainerStats {
	containerStatsFutures := make([]chan *grpc_monitoring_go.ContainerStatsResponse, 0, len(clusterList.Clusters))
	for _, cluster := range clusterList.Clusters {
		metricsCollector, derr := m.getMetricsCollectorClient(organization.OrganizationId, cluster.ClusterId)
		if derr != nil {
			log.Error().
				Str("organizationId", organization.OrganizationId).
				Str("clusterId", cluster.ClusterId).
				Err(derr).
				Msg("could not create metrics-collector client. The aggregation will not include this cluster stats.")
			continue
		}
		statsFuture := make(chan *grpc_monitoring_go.ContainerStatsResponse)
		containerStatsFutures = append(containerStatsFutures, statsFuture)
		go getClusterContainerStats(cluster, metricsCollector, ctx, statsFuture)
	}
	orgContainerStats := make([]*grpc_monitoring_go.ContainerStats, 0)
	for _, statsFuture := range containerStatsFutures {
		containerStatsResponse := <-statsFuture
		orgContainerStats = append(orgContainerStats, containerStatsResponse.ContainerStats...)
	}
	return orgContainerStats
}

func getClusterContainerStats(cluster *grpc_infrastructure_go.Cluster, metricsCollector *clients.MetricsCollectorClient, ctx context.Context, statsFuture chan *grpc_monitoring_go.ContainerStatsResponse) {
	getContainerStatsCtx, getContainerStatsCancel := context.WithTimeout(ctx, defaultTimeout)
	defer getContainerStatsCancel()
	clusterContainerStats, err := metricsCollector.GetContainerStats(getContainerStatsCtx, &grpc_common_go.Empty{})
	if err != nil {
		log.Error().
			Str("organizationId", cluster.OrganizationId).
			Str("clusterId", cluster.ClusterId).
			Err(err).
			Msg("metrics-collector responded with an error when querying for stats. The aggregation will not include this cluster stats.")
		statsFuture <- nil
	}
	// Adjust millicores metric with the cluster conversion factor
	for _, containerStat := range clusterContainerStats.ContainerStats {
		containerStat.CpuMillicore = containerStat.GetCpuMillicore() * cluster.MillicoresConversionFactor
	}
	statsFuture <- clusterContainerStats
}

func (m *Manager) aggregateStatsByServiceInstanceId(orgContainerStats []*grpc_monitoring_go.ContainerStats, request *grpc_monitoring_go.OrganizationApplicationStatsRequest, organizationName string) []*grpc_monitoring_go.OrganizationApplicationStats {
	statsMapByServiceInstanceId := make(map[string]*grpc_monitoring_go.OrganizationApplicationStats, 0)
	for _, containerStats := range orgContainerStats {
		stats, found := statsMapByServiceInstanceId[containerStats.ServiceInstanceId]
		if !found {
			// Include
			statsMapByServiceInstanceId[containerStats.ServiceInstanceId] = &grpc_monitoring_go.OrganizationApplicationStats{
				OrganizationId:           request.OrganizationId,
				OrganizationName:         organizationName,
				AppInstanceId:            containerStats.AppInstanceId,
				AppInstanceName:          containerStats.AppInstanceName,
				ServiceGroupInstanceId:   containerStats.ServiceGroupInstanceId,
				ServiceGroupInstanceName: containerStats.ServiceGroupInstanceName,
				ServiceInstanceId:        containerStats.ServiceInstanceId,
				ServiceInstanceName:      containerStats.ServiceInstanceName,
				CpuMillicore:             containerStats.CpuMillicore,
				MemoryByte:               containerStats.MemoryByte,
				StorageByte:              containerStats.StorageByte,
			}
		} else {
			// Aggregate
			stats.CpuMillicore += containerStats.CpuMillicore
			stats.MemoryByte += containerStats.MemoryByte
			stats.StorageByte += containerStats.StorageByte
		}
	}
	serviceInstanceStats := make([]*grpc_monitoring_go.OrganizationApplicationStats, 0, len(statsMapByServiceInstanceId))
	for _, value := range statsMapByServiceInstanceId {
		serviceInstanceStats = append(serviceInstanceStats, value)
	}
	return serviceInstanceStats
}
