/*
 * Copyright (C) 2019 Nalej Group - All Rights Reserved
 */

// Metrics manager responsible for collecting platform-relevant monitoring
// data and storing it in-memory, ready to be exported to a data store

package metrics

import (
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/provider/collector"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	provider collector.CollectorProvider
}

func NewManager(provider collector.CollectorProvider) (*Manager, derrors.Error) {
	manager := &Manager{
		provider: provider,
	}

	return manager, nil
}

func (m *Manager) Start() (derrors.Error) {
	log.Debug().Msg("starting metrics manager")

	go m.tickerLogger(time.Tick(time.Second * 10))

	// Start collector
	err := m.provider.Start()
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) tickerLogger(c <-chan time.Time) {
	for {
		<-c
		log.Debug().Msg("ticker")
		metrics, _ := m.provider.GetMetrics()
		for t, m := range(metrics) {
			log.Debug().
				Int64("created", m.Created).
				Int64("deleted", m.Deleted).
				Int64("running", m.CurrentRunning).
				Int64("error", m.CurrentError).
				Msg(string(t))
		}
	}
}
