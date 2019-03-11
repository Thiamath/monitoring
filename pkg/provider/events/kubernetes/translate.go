/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates event actions to platform metrics

package kubernetes

import (
	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	"github.com/rs/zerolog/log"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
)

type TranslateFuncs struct {
	// Collect metrics based on events
	collector metrics.Collector
}

func NewTranslateFuncs(collector metrics.Collector) *TranslateFuncs {
	return &TranslateFuncs{
		collector: collector,
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
		t.collector.Create(metrics.MetricServices)
	case EventDelete:
		t.collector.Delete(metrics.MetricServices)
	}
}
