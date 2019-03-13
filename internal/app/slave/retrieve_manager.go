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
	registry query.QueryProviderRegistry
}

// Create a new query manager. Without arguments, use the default
// registry for providers. With arguments, only the first registry
// is actually used
func NewRetrieveManager(r ...query.QueryProviderRegistry) (*RetrieveManager, derrors.Error) {
	registry := query.DefaultRegistry
	if len(r) > 0 {
		registry = r[0]
	}

	manager := &RetrieveManager{
		registry: registry,
	}

	return manager, nil
}

// Retrieve a summary of high level cluster resource availability
func (m *RetrieveManager) GetClusterSummary(context.Context, *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, derrors.Error) {
	return nil, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (m *RetrieveManager) GetClusterStats(context.Context, *grpc.ClusterStatsRequest) (*grpc.ClusterStats, derrors.Error) {
	return nil, nil

}

// Execute a query directly on the monitoring storage backend
func (m *RetrieveManager) Query(ctx context.Context, request *grpc.QueryRequest) (*grpc.QueryResponse, derrors.Error) {
	// Validate we have the right request type for the backend
	providerType := query.QueryProviderType(request.GetType().String())
	provider := m.registry.GetProvider(providerType)
	if provider == nil {
		return nil, derrors.NewUnavailableError(fmt.Sprintf("requested query provider %s not available", string(providerType)))
	}

	// Translate to backend query and execute
	queryRange := request.GetRange()
	q := &query.Query{
		QueryString: request.GetQuery(),
		Range: query.QueryRange{
			Step: time.Duration(queryRange.GetStep() * 1000 * 1000 * 1000),
		},
	}

	res, derr := provider.Query(ctx, q)
	if derr != nil {
		return nil, derr
	}

	// Translate result
	translator, found := translators.GetTranslator(providerType)
	if !found {
		return nil, derrors.NewInternalError(fmt.Sprintf("no result translator found for type %s", string(providerType)))
	}

	queryResponse, derr := translator(res)
	if derr != nil {
		return nil, derr
	}

	// Set original orginazation and cluster
	queryResponse.OrganizationId = request.GetOrganizationId()
	queryResponse.ClusterId = request.GetClusterId()

	fmt.Printf("%+v\n", queryResponse)
	return queryResponse, nil
}
