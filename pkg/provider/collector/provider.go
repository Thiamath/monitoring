/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Collector provider interface

package collector

import (
	"github.com/nalej/derrors"
)

// A collector provider collects metrics in the background and
// stores them in memory.
type CollectorProvider interface {
	// Start collecting metrics
	Start() (derrors.Error)
	// Stop collecting metrics
	Stop() (derrors.Error)

	// Get specific metrics, or all available when no specific metrics
	// are requested
	GetMetrics(metrics ...MetricType) (Metrics, derrors.Error)
}
