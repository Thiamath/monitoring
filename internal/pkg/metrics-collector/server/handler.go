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

// Handler for both slave and coord, implementing ExecQuery, GetClusterSummary
// and GetClusterStats.
// Implements grpc-monitoring-go.MetricsCollectorServer and MonitoringManagerServer

package server

import (
	"context"
	"github.com/nalej/monitoring/internal/pkg/entities"

	"github.com/nalej/derrors"

	grpc "github.com/nalej/grpc-monitoring-go"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	manager Manager
}

func NewHandler(m Manager) (*Handler, derrors.Error) {
	return &Handler{
		manager: m,
	}, nil
}

// Retrieve a summary of high level cluster resource availability
func (h *Handler) GetClusterSummary(ctx context.Context, request *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, error) {
	log.Debug().
		Str("organization_id", request.GetOrganizationId()).
		Str("cluster_id", request.GetClusterId()).
		Int32("avg", request.GetRangeMinutes()).
		Msg("received cluster summary request")

	// Validate
	derr := entities.ValidateClusterSummary(request)
	if derr != nil {
		log.Info().Str("err", derr.DebugReport()).Err(derr).Msg("invalid request")
		return nil, derr
	}

	res, derr := h.manager.GetClusterSummary(ctx, request)
	if derr != nil {
		log.Info().Str("err", derr.DebugReport()).Err(derr).Msg("error retrieving cluster summary")
		return nil, derr
	}

	return res, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (h *Handler) GetClusterStats(ctx context.Context, request *grpc.ClusterStatsRequest) (*grpc.ClusterStats, error) {
	log.Debug().
		Str("organization_id", request.GetOrganizationId()).
		Str("cluster_id", request.GetClusterId()).
		Int32("avg", request.GetRangeMinutes()).
		Msg("received cluster statistics request")

	// Validate
	derr := entities.ValidateClusterStats(request)
	if derr != nil {
		log.Info().Str("err", derr.DebugReport()).Err(derr).Msg("invalid request")
		return nil, derr
	}

	res, derr := h.manager.GetClusterStats(ctx, request)
	if derr != nil {
		log.Info().Str("err", derr.DebugReport()).Err(derr).Msg("error retrieving cluster statistics")
		return nil, derr
	}

	return res, nil
}

// Execute a query directly on the monitoring storage backend
func (h *Handler) Query(ctx context.Context, request *grpc.QueryRequest) (*grpc.QueryResponse, error) {
	log.Debug().
		Str("organization_id", request.GetOrganizationId()).
		Str("cluster_id", request.GetClusterId()).
		Str("type", request.GetType().String()).
		Str("query", request.GetQuery()).
		Msg("received query request")

	// Validate
	derr := entities.ValidateQuery(request)
	if derr != nil {
		log.Info().Str("err", derr.DebugReport()).Err(derr).Msg("invalid request")
		return nil, derr
	}

	// Execute
	res, err := h.manager.Query(ctx, request)
	if err != nil {
		log.Info().Str("err", err.DebugReport()).Err(err).Msg("error executing query")
		return nil, err
	}

	return res, nil
}
