/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Collect and store metrics in memory

package collector

import (
	"github.com/nalej/derrors"
)

type Collector struct {

}

func NewCollector() (*Collector, derrors.Error) {
	collector := &Collector{
	}

	return collector, nil
}
