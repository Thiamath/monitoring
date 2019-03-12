/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates event actions to platform metrics

package kubernetes

import (
	"time"

	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	extensions_v1beta1 "k8s.io/api/extensions/v1beta1"

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
		core_v1.SchemeGroupVersion.WithKind("PersistentVolumeClaim"),
		core_v1.SchemeGroupVersion.WithKind("Service"),
		extensions_v1beta1.SchemeGroupVersion.WithKind("Ingress"),
	}
}

// Translating functions
func (t *TranslateFuncs) OnDeployment(obj interface{}, action EventAction) {
	d := obj.(*apps_v1.Deployment)

	// filter out zt-agent deployment
	agentLabel, found := d.Labels["agent"]
	if found && agentLabel == "zt-agent" {
		return
	}

	t.translate(action, metrics.MetricServices, d.CreationTimestamp)
}

func (t *TranslateFuncs) OnNamespace(obj interface{}, action EventAction) {
	n := obj.(*core_v1.Namespace)
	t.translate(action, metrics.MetricFragments, n.CreationTimestamp)
}

func (t *TranslateFuncs) OnPersistentVolumeClaim(obj interface{}, action EventAction) {
	pvc := obj.(*core_v1.PersistentVolumeClaim)
	t.translate(action, metrics.MetricVolumes, pvc.CreationTimestamp)
}

func (t *TranslateFuncs) OnIngress(obj interface{}, action EventAction) {
	i := obj.(*extensions_v1beta1.Ingress)
	t.translate(action, metrics.MetricEndpoints, i.CreationTimestamp)
}

func (t *TranslateFuncs) OnService(obj interface{}, action EventAction) {
	s := obj.(*core_v1.Service)

	if s.Spec.Type != core_v1.ServiceTypeLoadBalancer {
		return
	}

	t.translate(action, metrics.MetricEndpoints, s.CreationTimestamp)
}

// Calls create function if after server startup, existing otherwise
func (t *TranslateFuncs) translate(action EventAction, metric metrics.MetricType, ts meta_v1.Time) {
	switch action {
	case EventAdd:
		if ts.After(t.startupTime) {
			t.collector.Create(metric)
		} else {
			t.collector.Existing(metric)
		}
	case EventDelete:
		t.collector.Delete(metric)
	}
}
