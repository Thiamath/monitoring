/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Slave implementation for RetrieveManager

package slave

import (
	"context"
	"fmt"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/internal/pkg/retrieve/translators"
	"github.com/nalej/infrastructure-monitor/pkg/metrics"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"

	grpc "github.com/nalej/grpc-infrastructure-monitor-go"
)

type RetrieveManager struct {
	providers query.QueryProviders
	featureProviders map[query.QueryProviderFeature]query.QueryProvider
}

// Create a new query manager.
func NewRetrieveManager(providers query.QueryProviders) (*RetrieveManager, derrors.Error) {
	// Check providers for specific features
	// NOTE: this only gives us the last provider with a certain feature,
	// but at least we have one we can use
	featureProviders := map[query.QueryProviderFeature]query.QueryProvider{}
	for _, provider := range(providers) {
		for _, feature := range(provider.Supported()) {
			featureProviders[feature] = provider
		}
	}

	manager := &RetrieveManager{
		providers: providers,
		featureProviders: featureProviders,
	}

	return manager, nil
}

// Retrieve a summary of high level cluster resource availability
func (m *RetrieveManager) GetClusterSummary(ctx context.Context, request *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, derrors.Error) {
	// Get right provider
	provider, found := m.featureProviders[query.FeatureSystemStats]
	if !found {
		return nil, derrors.NewUnavailableError("no query provider for system statistics")
	}

	vars := &query.TemplateVars{
		AvgSeconds: request.GetRangeMinutes() * 60,
	}

	// Create result
	res := &grpc.ClusterSummary{
		OrganizationId: request.GetOrganizationId(),
		ClusterId: request.GetClusterId(),
	}

	// Create mapping to fill
	resultMap := map[query.TemplateName]**grpc.ClusterStat{
		query.TemplateName_CPU: &res.CpuMillicores,
		query.TemplateName_Memory: &res.MemoryBytes,
		query.TemplateName_Storage: &res.StorageBytes,
	}

	for name, stat := range(resultMap) {
		available, derr := provider.ExecuteTemplate(ctx, name + query.TemplateName_Available, vars)
		if derr != nil {
			return nil, derr
		}
		total, derr := provider.ExecuteTemplate(ctx, name + query.TemplateName_Total, vars)
		if derr != nil {
			return nil, derr
		}

		*stat = &grpc.ClusterStat{
			Total: total,
			Available: available,
		}
	}

	return res, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (m *RetrieveManager) GetClusterStats(ctx context.Context, request *grpc.ClusterStatsRequest) (*grpc.ClusterStats, derrors.Error) {
	// Get right provider
	provider, found := m.featureProviders[query.FeaturePlatformStats]
	if !found {
		return nil, derrors.NewUnavailableError("no query provider for platform statistics")
	}

	vars := &query.TemplateVars{
		AvgSeconds: request.GetRangeMinutes() * 60,
	}

	var stats = map[int32]*grpc.PlatformStat{}

	// TODO: use request fields
	// TODO: parallel queries
	for _, metric := range(metrics.AllMetrics) {
		stat := &grpc.PlatformStat{}
		// Create mapping to fill
		resultMap := map[metrics.MetricCounter]*int64{
			metrics.MetricCreated: &stat.Created,
			metrics.MetricDeleted: &stat.Deleted,
			metrics.MetricErrors: &stat.Errors,
			metrics.MetricRunning: &stat.Running,
		}

		vars.MetricName = metric.String()
		for counter, valPtr := range(resultMap) {
			// Determine template based on value type (counter, gauge)
			var templateName query.TemplateName
			valType, found := metrics.CounterMap[counter]
			if !found {
				return nil, derrors.NewUnavailableError("no appropriate statistic available")
			}

			vars.StatName = counter.String()
			switch valType {
			case metrics.ValueCounter:
				templateName = query.TemplateName_PlatformStatsCounter
			case metrics.ValueGauge:
				templateName = query.TemplateName_PlatformStatsGauge
			default:
				return nil, derrors.NewUnavailableError("no appropriate query template available")
			}

			val, derr := provider.ExecuteTemplate(ctx, templateName, vars)
			if derr != nil {
				return nil, derr
			}
			*valPtr = val
		}

		statsFieldNumber, found := grpc.PlatformStatsField_value[metric.ToAPI()]
		if !found {
			return nil, derrors.NewUnavailableError("no mapping between statistic and API result message")
		}
		stats[statsFieldNumber] = stat
	}

	// Create result
	res := &grpc.ClusterStats{
		OrganizationId: request.GetOrganizationId(),
		ClusterId: request.GetClusterId(),
		Stats: stats,
	}

	return res, nil
}

// Execute a query directly on the monitoring storage backend
func (m *RetrieveManager) Query(ctx context.Context, request *grpc.QueryRequest) (*grpc.QueryResponse, derrors.Error) {
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
			Start: translators.GoTime(queryRange.GetStart()),
			End: translators.GoTime(queryRange.GetEnd()),
			// Step is a float32 in seconds, convert to int64 in nanos
			Step: time.Duration(queryRange.GetStep() * float32(1000 * 1000 * 1000)),
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
