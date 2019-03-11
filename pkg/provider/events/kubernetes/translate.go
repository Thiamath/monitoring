/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates event actions to platform metrics

package kubernetes

import (
	"time"

	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	"github.com/rs/zerolog/log"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type TranslateFuncs struct {
	// Collect metrics based on events
	collector metrics.Collector
	// Server startup
	startupTime time.Time
}

// NOTE: On start, we should count the incoming create/delete to update
// total running, but created/deleted should be 0 to properly use
// the prometheus counters

func NewTranslateFuncs(collector metrics.Collector) *TranslateFuncs {
	return &TranslateFuncs{
		collector: collector,
		startupTime: time.Now(),
	}
}

func (t *TranslateFuncs) SupportedKinds() KindList {
	return KindList{
		apps_v1.SchemeGroupVersion.WithKind("Deployment"),
		core_v1.SchemeGroupVersion.WithKind("Namespace"),
	}
}

// Translating functions
func (t *TranslateFuncs) OnDeployment(obj interface{}, action EventAction) {
	d := obj.(*apps_v1.Deployment)
	log.Debug().Str("name", d.Name).Msg("deployment")
}

func (t *TranslateFuncs) OnNamespace(obj interface{}, action EventAction) {
	n := obj.(*core_v1.Namespace)
	log.Debug().Str("name", n.Name).Msg("namespace")
	switch action {
	case EventAdd:
		t.createOrExisting(metrics.MetricServices, n.CreationTimestamp)
	case EventDelete:
		t.collector.Delete(metrics.MetricServices)
	}
}

// Calls create function if after server startup, existing otherwise
func (t *TranslateFuncs) createOrExisting(metric metrics.MetricType, ts meta_v1.Time) {
	if ts.After(t.startupTime) {
		t.collector.Create(metric)
	} else {
		t.collector.Existing(metric)
	}
}
