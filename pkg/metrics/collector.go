/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Interface for a collector that stores metrics

package metrics

import (
	"github.com/nalej/derrors"
)

type Collector interface {
	Create(t MetricType)
	Delete(t MetricType)
	GetMetrics(types ...MetricType) (Metrics, derrors.Error)
}
