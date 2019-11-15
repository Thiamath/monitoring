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

// RetrieveManager interface definition

package retrieve

import (
	"context"

	"github.com/nalej/derrors"

	grpc "github.com/nalej/grpc-monitoring-go"
)

type RetrieveManager interface {
	// Retrieve a summary of high level cluster resource availability
	GetClusterSummary(context.Context, *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, derrors.Error)
	// Retrieve statistics on cluster with respect to platform resources
	GetClusterStats(context.Context, *grpc.ClusterStatsRequest) (*grpc.ClusterStats, derrors.Error)
	// Execute a query directly on the monitoring storage backend
	Query(context.Context, *grpc.QueryRequest) (*grpc.QueryResponse, derrors.Error)
}
