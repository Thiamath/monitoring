/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// RetrieveManager interface definition

package retrieve

import (
	"context"

	"github.com/nalej/derrors"

        grpc "github.com/nalej/grpc-infrastructure-monitor-go"
)

type RetrieveManager interface {
	// Retrieve a summary of high level cluster resource availability
	GetClusterSummary(context.Context, *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, derrors.Error)
	// Retrieve statistics on cluster with respect to platform resources
	GetClusterStats(context.Context, *grpc.ClusterStatsRequest) (*grpc.ClusterStats, derrors.Error)
	// Execute a query directly on the monitoring storage backend
	Query(context.Context, *grpc.QueryRequest) (*grpc.QueryResponse, derrors.Error)
}
