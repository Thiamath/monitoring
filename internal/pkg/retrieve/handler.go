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
func (h *Handler) GetClusterSummary(context.Context, *grpc.ClusterSummaryRequest) (*grpc.ClusterSummary, error) {
	return nil, nil
}

// Retrieve statistics on cluster with respect to platform resources
func (h *Handler) GetClusterStats(context.Context, *grpc.ClusterStatsRequest) (*grpc.ClusterStats, error) {
	return nil, nil
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
	// TODO: check cluster id
	// Check querytype is set
	// check querystring is not nil

	// Execute
	res, err := h.manager.Query(ctx, request)
	if err != nil {
		log.Info().Str("err", err.DebugReport()).Err(err).Msg("error executing query")
		// TODO: convert error
		return nil, err
	}

	return res, nil
}
