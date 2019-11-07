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
 *
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
