/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query result implementation

package query

import (
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
)

type PrometheusResult struct {
}

func (r *PrometheusResult) Type() query.QueryProviderType {
	return providerType
}
