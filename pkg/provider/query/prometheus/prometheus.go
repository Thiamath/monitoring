/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query provider implementation

package query

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
)

const providerType query.QueryProviderType = "PROMETHEUS"

type PrometheusProvider struct {
}

func NewProvider() (*PrometheusProvider, derrors.Error) {
	return &PrometheusProvider{}, nil
}

// Returns the query provider type
func (p *PrometheusProvider) Type() query.QueryProviderType {
	return providerType
}

// Execute query q.
func (p *PrometheusProvider) Query(ctx context.Context, q *query.Query) (query.QueryResult, derrors.Error) {
	return &PrometheusResult{}, nil
}
