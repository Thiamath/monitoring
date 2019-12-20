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

package server

import (
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/internal/pkg/entities"
	"github.com/rs/zerolog/log"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/api/httpbody"
)

type Handler struct {
	manager *Manager
}

func NewHandler(manager *Manager) (*Handler, derrors.Error) {
	return &Handler{manager: manager}, nil
}

func (h *Handler) Metrics(ctx context.Context, request *grpc_monitoring_go.OrganizationApplicationStatsRequest) (*httpbody.HttpBody, error) {
	log.Info().Interface("request", request).Msg("Got metrics request")
	derr := entities.ValidateOrganizationApplicationStatsRequest(request)
	if derr != nil {
		return nil, derr
	}

	response, err := h.manager.Metrics(request.OrganizationId)
	if err != nil {
		return nil, err
	}
	return &httpbody.HttpBody{
		ContentType: "text/plain; version=0.0.4",
		Data:        response,
	}, nil
}
