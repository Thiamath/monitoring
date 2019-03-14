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
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"

	grpc "github.com/nalej/grpc-infrastructure-monitor-go"
)

type RetrieveManager struct {
	providers query.QueryProviders
	defaultProvider query.QueryProviderType
}

// Create a new query manager.
func NewRetrieveManager(providers query.QueryProviders, defaultProvider query.QueryProviderType) (*RetrieveManager, derrors.Error) {
	manager := &RetrieveManager{
		providers: providers,
		defaultProvider: defaultProvider,
	}

	_, found := providers[defaultProvider]
	if !found {
		return nil, derrors.NewUnavailableError("default provider not available")
	}

	return manager, nil
}

// Retrieve a summary of high level cluster resource availability
func (m *RetrieveManager) GetClusterSummary(ctx context.Context, request *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, derrors.Error) {
	avg := time.Minute * time.Duration(request.GetRangeMinutes())

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
		available, derr := m.providers[m.defaultProvider].ExecuteTemplate(ctx, name + query.TemplateName_Available, avg)
		if derr != nil {
			return nil, derr
		}
		total, derr := m.providers[m.defaultProvider].ExecuteTemplate(ctx, name + query.TemplateName_Total, avg)
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
func (m *RetrieveManager) GetClusterStats(context.Context, *grpc.ClusterStatsRequest) (*grpc.ClusterStats, derrors.Error) {
	return nil, nil

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
