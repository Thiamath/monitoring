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

// RetrieveManager handles metrics queries

package metrics_collector

import (
	"context"
	"fmt"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-utils/pkg/conversions"

	"github.com/nalej/monitoring/internal/pkg/retrieve/translators"
	"github.com/nalej/monitoring/pkg/provider/query"

	grpc "github.com/nalej/grpc-monitoring-go"
)

type RetrieveManager struct {
	providers        query.QueryProviders
	featureProviders map[query.QueryProviderFeature]query.QueryProvider
}

// Create a new query manager.
func NewRetrieveManager(providers query.QueryProviders) (*RetrieveManager, derrors.Error) {
	// Check providers for specific features
	// NOTE: this only gives us the last provider with a certain feature,
	// but at least we have one we can use
	featureProviders := map[query.QueryProviderFeature]query.QueryProvider{}
	for _, provider := range providers {
		for _, feature := range provider.Supported() {
			featureProviders[feature] = provider
		}
	}

	manager := &RetrieveManager{
		providers:        providers,
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
		ClusterId:      request.GetClusterId(),
	}

	// Create mapping to fill
	resultMap := map[query.TemplateName]**grpc.ClusterStat{
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

		*stat = &grpc.ClusterStat{
			Total:     total,
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

	// If no specific fields are requested, get all
	fields := request.GetFields()
	if len(fields) == 0 {
		fields = AllGRPCStatsFields()
	}

	// TODO: parallel queries
	var stats = map[int32]*grpc.PlatformStat{}
	for _, field := range fields {
		stat := &grpc.PlatformStat{}

		// Create mapping to fill
		resultMap := map[query.MetricCounter]*int64{
			query.MetricCreated: &stat.Created, // counter
			query.MetricDeleted: &stat.Deleted, // counter
			query.MetricErrors:  &stat.Errors,  // counter
			query.MetricRunning: &stat.Running, // gauge
		}

		vars.MetricName = GRPCStatsFieldToMetric(field)
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
	res := &grpc.ClusterStats{
		OrganizationId: request.GetOrganizationId(),
		ClusterId:      request.GetClusterId(),
		Stats:          stats,
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
