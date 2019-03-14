/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Handler for both slave and coord, implementing ExecQuery, GetClusterSummary
// and GetClusterStats.
// Implements grpc-infrastructure-monitor-go.SlaveServer and CoordinatorServer

package retrieve

import (
	"context"

	"github.com/nalej/derrors"

        grpc "github.com/nalej/grpc-infrastructure-monitor-go"

	"github.com/rs/zerolog/log"
)

type Handler struct {
	manager RetrieveManager
}

func NewHandler(m RetrieveManager) (*Handler, derrors.Error) {
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
	derr := validateClusterSummary(request)
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
func (h *Handler) GetClusterStats(context.Context, *grpc.ClusterStatsRequest) (*grpc.ClusterStats, error) {
	return nil, derrors.NewUnimplementedError("GetClusterStats is not implemented")
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
	derr := validateQuery(request)
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
