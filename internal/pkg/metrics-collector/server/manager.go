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

// Manager handles metrics queries

package server

import (
	"context"
	"fmt"
	"github.com/nalej/grpc-common-go"
	"github.com/nalej/monitoring/internal/pkg/metrics-collector"
	"github.com/nalej/monitoring/pkg/provider/query/prometheus"
	"strconv"
	"time"

	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"

	"github.com/nalej/grpc-utils/pkg/conversions"

	"github.com/nalej/monitoring/internal/pkg/metrics-collector/translators"
	"github.com/nalej/monitoring/pkg/provider/query"

	"github.com/nalej/grpc-monitoring-go"
)

const (
	CpuQuery     = "nalej_servinst_cpu_core"
	MemoryQuery  = "nalej_servinst_memory_byte"
	StorageQuery = "nalej_servinst_storage_byte"
)

// Manager structure with the required clients for roles operations.
type Manager struct {
	providers        query.QueryProviders
	featureProviders map[query.QueryProviderFeature]query.QueryProvider
}

// NewManager creates a new query manager.
func NewManager(providers query.QueryProviders) (Manager, derrors.Error) {
	// Check providers for specific features
	// NOTE: this only gives us the last provider with a certain feature,
	// but at least we have one we can use
	featureProviders := map[query.QueryProviderFeature]query.QueryProvider{}
	for _, provider := range providers {
		for _, feature := range provider.Supported() {
			featureProviders[feature] = provider
		}
	}

	manager := Manager{
		providers:        providers,
		featureProviders: featureProviders,
	}

	return manager, nil
}

// GetClusterSummary retrieves a summary of high level cluster resource availability
func (m *Manager) GetClusterSummary(ctx context.Context, request *grpc_monitoring_go.ClusterSummaryRequest) (*grpc_monitoring_go.ClusterSummary, error) {
	// Get right provider
	provider, found := m.featureProviders[query.FeatureSystemStats]
	if !found {
		return nil, derrors.NewUnavailableError("no query provider for system statistics")
	}

	vars := &query.TemplateVars{
		AvgSeconds: request.GetRangeMinutes() * 60,
	}

	// Create result
	res := &grpc_monitoring_go.ClusterSummary{
		OrganizationId: request.GetOrganizationId(),
		ClusterId:      request.GetClusterId(),
	}

	// Create mapping to fill
	resultMap := map[query.TemplateName]**grpc_monitoring_go.ClusterStat{
		query.TemplateName_CPU:           &res.CpuMillicores,
		query.TemplateName_Memory:        &res.MemoryBytes,
		query.TemplateName_Storage:       &res.StorageBytes,
		query.TemplateName_UsableStorage: &res.UsableStorageBytes,
	}

	for name, stat := range resultMap {
		available, derr := provider.ExecuteTemplate(ctx, name+query.TemplateName_Available, vars)
		if derr != nil {
			return nil, derr
		}
		total, derr := provider.ExecuteTemplate(ctx, name+query.TemplateName_Total, vars)
		if derr != nil {
			return nil, derr
		}

		*stat = &grpc_monitoring_go.ClusterStat{
			Total:     total,
			Available: available,
		}
	}

	return res, nil
}

// GetClusterStats retrieves statistics on cluster with respect to platform resources
func (m *Manager) GetClusterStats(ctx context.Context, request *grpc_monitoring_go.ClusterStatsRequest) (*grpc_monitoring_go.ClusterStats, error) {
	// Get right provider
	provider, found := m.featureProviders[query.FeaturePlatformStats]
	if !found {
		return nil, derrors.NewUnavailableError("no query provider for platform statistics")
	}

	vars := &query.TemplateVars{
		AvgSeconds: request.GetRangeMinutes() * 60,
	}

	// If no specific fields are requested, get all
	fields := request.GetFields()
	if len(fields) == 0 {
		fields = metrics_collector.AllGRPCStatsFields()
	}

	// TODO: parallel queries
	var stats = map[int32]*grpc_monitoring_go.PlatformStat{}
	for _, field := range fields {
		stat := &grpc_monitoring_go.PlatformStat{}

		// Create mapping to fill
		resultMap := map[query.MetricCounter]*int64{
			query.MetricCreated: &stat.Created, // counter
			query.MetricDeleted: &stat.Deleted, // counter
			query.MetricErrors:  &stat.Errors,  // counter
			query.MetricRunning: &stat.Running, // gauge
		}

		vars.MetricName = metrics_collector.GRPCStatsFieldToMetric(field)
		for counter, valPtr := range resultMap {
			// Determine template based on value type (counter, gauge)
			templateName, derr := query.GetPlatformTemplateName(counter)
			if derr != nil {
				return nil, derr
			}

			vars.StatName = counter.String()
			val, derr := provider.ExecuteTemplate(ctx, templateName, vars)
			if derr != nil {
				return nil, derr
			}
			*valPtr = val
		}

		stats[int32(field)] = stat
	}

	// Create result
	res := &grpc_monitoring_go.ClusterStats{
		OrganizationId: request.GetOrganizationId(),
		ClusterId:      request.GetClusterId(),
		Stats:          stats,
	}

	return res, nil
}

// Query executes a query directly on the monitoring storage backend
func (m *Manager) Query(ctx context.Context, request *grpc_monitoring_go.QueryRequest) (*grpc_monitoring_go.QueryResponse, error) {
	// Validate we have the right request type for the backend
	providerType := query.QueryProviderType(request.GetType().String())
	provider, found := m.providers[providerType]
	if !found {
		return nil, derrors.NewUnavailableError(fmt.Sprintf("requested query provider %s not available", string(providerType)))
	}

	// Translate to backend query and execute
	queryRange := request.GetRange()
	q := &query.Query{
		QueryString: request.GetQuery(),
		Range: query.QueryRange{
			Start: conversions.GoTime(queryRange.GetStart()),
			End:   conversions.GoTime(queryRange.GetEnd()),
			// Step is a float32 in seconds, convert to int64 in nanos
			Step: time.Duration(queryRange.GetStep() * float32(1000*1000*1000)),
		},
	}

	res, derr := provider.Query(ctx, q)
	if derr != nil {
		return nil, derr
	}

	// Translate result
	translator, found := translators.GetTranslator(providerType)
	if !found {
		return nil, derrors.NewUnimplementedError(fmt.Sprintf("no result translator found for type %s", string(providerType)))
	}

	queryResponse, derr := translator(res)
	if derr != nil {
		return nil, derr
	}

	// Set original orginazation and cluster
	queryResponse.OrganizationId = request.GetOrganizationId()
	queryResponse.ClusterId = request.GetClusterId()

	return queryResponse, nil
}

func (m *Manager) GetContainerStats(ctx context.Context, _ *grpc_common_go.Empty) (*grpc_monitoring_go.ContainerStatsResponse, error) {
	// Validate we have the right request type for the backend
	providerType := prometheus.ProviderType
	provider, found := m.providers[providerType]
	if !found {
		return nil, derrors.NewUnavailableError(fmt.Sprintf("requested query provider %s not available", string(providerType)))
	}
	// Translate result
	translator, found := translators.GetTranslator(provider.ProviderType())
	if !found {
		return nil, derrors.NewNotFoundError(fmt.Sprintf("no result translator found for provider %s", string(provider.ProviderType())))
	}

	queryTime := time.Now()

	cpuStatsFuture := getCpuStats(queryTime, ctx, provider, translator)
	memoryStatsFuture := getMemoryStats(queryTime, ctx, provider, translator)
	storageStatsFuture := getStorageStats(queryTime, ctx, provider, translator)

	cpuStats := <-cpuStatsFuture
	memoryStats := <-memoryStatsFuture
	storageStats := <-storageStatsFuture

	statsMapByNamespace := make(map[string]map[string]*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue, 0)
	if cpuStats == nil {
		log.Warn().Msg(CpuQuery + " stats could not be retrieved and will not be aggregated")
	} else {
		mapQueryResultsByNamespace(CpuQuery, cpuStats, statsMapByNamespace)
	}
	if memoryStats == nil {
		log.Warn().Msg(MemoryQuery + " stats could not be retrieved and will not be aggregated")
	} else {
		mapQueryResultsByNamespace(MemoryQuery, memoryStats, statsMapByNamespace)
	}
	if storageStats == nil {
		log.Warn().Msg(StorageQuery + " stats could not be retrieved and will not be aggregated")
	} else {
		mapQueryResultsByNamespace(StorageQuery, cpuStats, statsMapByNamespace)
	}

	containerStats := make([]*grpc_monitoring_go.ContainerStats, 0)
	for namespace, statsByType := range statsMapByNamespace {
		pod := statsByType[CpuQuery].Metric["pod"]
		container := statsByType[CpuQuery].Metric["container"]
		cpuMillicore, _ := strconv.ParseFloat(statsByType[CpuQuery].Value[0].Value, 64)
		memoryByte, _ := strconv.ParseFloat(statsByType[MemoryQuery].Value[0].Value, 64)
		storageByte, _ := strconv.ParseFloat(statsByType[StorageQuery].Value[0].Value, 64)

		stats := grpc_monitoring_go.ContainerStats{
			Namespace:    namespace,
			Pod:          pod,
			Container:    container,
			CpuMillicore: cpuMillicore,
			MemoryByte:   memoryByte,
			StorageByte:  storageByte,
		}
		containerStats = append(containerStats, &stats)
	}

	containerStatsResponse := &grpc_monitoring_go.ContainerStatsResponse{
		ContainerStats: containerStats,
	}
	return containerStatsResponse, nil
}

func mapQueryResultsByNamespace(metricName string, results *grpc_monitoring_go.QueryResponse, statsMap map[string]map[string]*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue) {
	for _, result := range results.GetPrometheusResult().GetResult() {
		namespace := result.Metric["namespace"]
		namespaceMetrics, exists := statsMap[namespace]
		if !exists {
			namespaceMetrics = make(map[string]*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue, 0)
			statsMap[namespace] = namespaceMetrics
		}
		namespaceMetrics[metricName] = result
	}
}

func getCpuStats(queryTime time.Time, ctx context.Context, provider query.QueryProvider, translator translators.TranslatorFunc) chan *grpc_monitoring_go.QueryResponse {
	future := make(chan *grpc_monitoring_go.QueryResponse)
	go launchQuery(CpuQuery, queryTime, provider, ctx, translator, future)
	return future
}

func getMemoryStats(queryTime time.Time, ctx context.Context, provider query.QueryProvider, translator translators.TranslatorFunc) chan *grpc_monitoring_go.QueryResponse {
	future := make(chan *grpc_monitoring_go.QueryResponse)
	go launchQuery(MemoryQuery, queryTime, provider, ctx, translator, future)
	return future
}

func getStorageStats(queryTime time.Time, ctx context.Context, provider query.QueryProvider, translator translators.TranslatorFunc) chan *grpc_monitoring_go.QueryResponse {
	future := make(chan *grpc_monitoring_go.QueryResponse)
	go launchQuery(StorageQuery, queryTime, provider, ctx, translator, future)
	return future
}

func launchQuery(queryString string, queryTime time.Time, provider query.QueryProvider, ctx context.Context, translator translators.TranslatorFunc, future chan *grpc_monitoring_go.QueryResponse) {
	q := &query.Query{
		QueryString: queryString,
		Range: query.QueryRange{
			Start: queryTime,
			End:   time.Time{},
			// Step is a float32 in seconds, convert to int64 in nanos
			Step: time.Duration(0),
		},
	}
	res, derr := provider.Query(ctx, q)
	if derr != nil {
		log.Error().
			Interface("query", q).
			Msg("error getting cpu stats")
		future <- nil
	}
	queryResponse, derr := translator(res)
	if derr != nil {
		log.Error().
			Interface("query", q).
			Interface("response", queryResponse).
			Msg("error translating cpu stats")
		future <- nil
	}
	future <- queryResponse
}
