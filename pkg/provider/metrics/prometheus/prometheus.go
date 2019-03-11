/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus implementation for metrics interface

package prometheus

import (
	"net/http"

	"github.com/nalej/derrors"
)

type MetricsProvider struct {
}

func NewMetricsProvider() (*MetricsProvider, derrors.Error) {
	return nil, nil
}

func (p *MetricsProvider) Metrics(w http.ResponseWriter, r *http.Request) {
}
