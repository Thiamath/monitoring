/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package asset

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/grpc-utils/pkg/conversions"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) (*Handler, derrors.Error) {
	h := &Handler{
		manager: manager,
	}

	return h, nil
}

func (h *Handler) ListMetrics(ctx context.Context, selector *grpc_inventory_go.AssetSelector) (*grpc_monitoring_go.MetricsList, error) {
	derr := ValidAssetSelector(selector)
	if derr != nil {
		return nil, conversions.ToGRPCError(derr)
	}

	return h.manager.ListMetrics(selector)
}

func (h *Handler) QueryMetrics(ctx context.Context, request *grpc_monitoring_go.QueryMetricsRequest) (*grpc_monitoring_go.QueryMetricsResult, error) {
	derr := ValidQueryMetricsRequest(request)
	if derr != nil {
		return nil, conversions.ToGRPCError(derr)
	}

	return h.manager.QueryMetrics(request)
}
