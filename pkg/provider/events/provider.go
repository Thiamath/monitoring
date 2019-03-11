/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Events provider interface

package events

import (
	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/metrics"
)

// An events provider listens to events that provider metrics.
// Depending on the provider, it stores those events internally or in
// a Collector. Useful metrics may be retrieved through GetMetrics
// or through the Collector if used.
type EventsProvider interface {
	// Start collecting metrics
	Start() (derrors.Error)
	// Stop collecting metrics
	Stop() (derrors.Error)

	// Get specific metrics, or all available when no specific metrics
	// are requested
	GetMetrics(types ...metrics.MetricType) (metrics.Metrics, derrors.Error)
}
