/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Events provider interface

package events

import (
	"github.com/nalej/derrors"
)

// An events provider listens to events that provide metrics.
// NOTE: If we use a collector, do not assume it is thread safe and do not
// call concurrently
type EventsProvider interface {
	// Start collecting metrics
	Start() (derrors.Error)
	// Stop collecting metrics
	Stop() (derrors.Error)
}
