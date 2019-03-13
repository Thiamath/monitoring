/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Slave implementation for RetrieveManager

package slave

import (
	"context"

	"github.com/nalej/derrors"

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

	// Translate to backend query and execute
	// querystring and range

	// Translate result
	// add org, cluster, type
	// resulttype, resultvalue

	return nil, nil
}
