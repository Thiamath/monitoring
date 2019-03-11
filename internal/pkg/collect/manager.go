/*
 * Copyright (C) 2019 Nalej Group - All Rights Reserved
 */

// Collect manager responsible for collecting platform-relevant monitoring
// data and storing it in-memory, ready to be exported to a data store

package collect

import (
	"net/http"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/provider/events"
	"github.com/nalej/infrastructure-monitor/pkg/provider/metrics"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	eventsProvider events.EventsProvider
	metricsProvider metrics.MetricsProvider
}

func NewManager(eventsProvider events.EventsProvider, metricsProvider metrics.MetricsProvider) (*Manager, derrors.Error) {
	manager := &Manager{
		eventsProvider: eventsProvider,
		metricsProvider: metricsProvider,
	}

	return manager, nil
}

func (m *Manager) Start() (derrors.Error) {
	log.Debug().Msg("starting metrics manager")

	go m.tickerLogger(time.Tick(time.Second * 10))

	// Start collecting events
	err := m.eventsProvider.Start()
	if err != nil {
		return err
	}

	return nil
}

func (m *Manager) Metrics(w http.ResponseWriter, r *http.Request) {
	m.metricsProvider.Metrics(w, r)
}

func (m *Manager) tickerLogger(c <-chan time.Time) {
	for {
		<-c
		log.Debug().Msg("ticker")
		metrics, _ := m.eventsProvider.GetMetrics()
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
