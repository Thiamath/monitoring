/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Interface for a collector that stores metrics

package metrics

import (
	"github.com/nalej/derrors"
)

// Interface to collect creation/deletion events of metrics.
// NOTE: We assume these are not thread-safe and should not be called
// concurrently
type Collector interface {
	Create(t MetricType)
	Delete(t MetricType)
	GetMetrics(types ...MetricType) (Metrics, derrors.Error)
}
