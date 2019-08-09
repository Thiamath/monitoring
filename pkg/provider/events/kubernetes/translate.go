/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates event actions to platform metrics

package kubernetes

import (
	"fmt"

	"github.com/nalej/deployment-manager/pkg/utils"
	"github.com/nalej/monitoring/pkg/metrics"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"
	extensions_v1beta1 "k8s.io/api/extensions/v1beta1"
	"k8s.io/apimachinery/pkg/api/meta"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
        "k8s.io/client-go/tools/cache"

	"github.com/rs/zerolog/log"
)

var (
	DeploymentKind = apps_v1.SchemeGroupVersion.WithKind("Deployment")
	NamespaceKind = core_v1.SchemeGroupVersion.WithKind("Namespace")
	PVCKind = core_v1.SchemeGroupVersion.WithKind("PersistentVolumeClaim")
	PodKind = core_v1.SchemeGroupVersion.WithKind("Pod")
	ServiceKind = core_v1.SchemeGroupVersion.WithKind("Service")
	EventKind = core_v1.SchemeGroupVersion.WithKind("Event")
	IngressKind = extensions_v1beta1.SchemeGroupVersion.WithKind("Ingress")
)

// We'll filter out these labels
var FilterLabelSet = map[string]string{
	"agent": "zt-agent",
}

type TranslateFuncs struct {
	// Collect metrics based on events
	collector metrics.Collector
	// Server startup
	startupTime meta_v1.Time

	// We have a reference back to the data stores of the informers for
	// each kind of resource we support. We use this to look up and
	// cross-reference resources to figure out the needed translation.
	stores map[string]cache.Store
}

// NOTE: On start, we count the incoming create/delete to update
// total running, but created/deleted should be 0 to properly use
// the prometheus counters
func NewTranslateFuncs(collector metrics.Collector) *TranslateFuncs {
	return &TranslateFuncs{
		collector: collector,
		startupTime: meta_v1.Now(),
		stores: map[string]cache.Store{},
	}
}

func (t *TranslateFuncs) SetStore(kind schema.GroupVersionKind, store cache.Store) error {
	_, found := t.stores[kind.Kind]
	if found {
		return fmt.Errorf("Store for %s already set", kind.Kind)
	}

	t.stores[kind.Kind] = store
	return nil
}

func (t *TranslateFuncs) SupportedKinds() KindList {
	return KindList{
		DeploymentKind,
		NamespaceKind,
		PVCKind,
		// We only watch this so we have the resource store
		PodKind,
		ServiceKind,
		EventKind,
		IngressKind,
	}
}

// Translating functions
func (t *TranslateFuncs) OnDeployment(oldObj, obj interface{}, action EventType) error {
	d := obj.(*apps_v1.Deployment)

	// filter out zt-agent deployment
	if !isAppInstance(d) {
		return nil
	}

	return t.translate(action, metrics.MetricServices, &d.CreationTimestamp)
}

func (t *TranslateFuncs) OnNamespace(oldObj, obj interface{}, action EventType) error {
	n := obj.(*core_v1.Namespace)

	return t.translate(action, metrics.MetricFragments, &n.CreationTimestamp)
}

func (t *TranslateFuncs) OnPersistentVolumeClaim(oldObj, obj interface{}, action EventType) error {
	pvc := obj.(*core_v1.PersistentVolumeClaim)
	return t.translate(action, metrics.MetricVolumes, &pvc.CreationTimestamp)
}

func (t *TranslateFuncs) OnPod(oldObj, obj interface{}, action EventType) error {
	// No action - only watched to have the resource store for reference
	return nil
}

func (t *TranslateFuncs) OnIngress(oldObj, obj interface{}, action EventType) error {
	i := obj.(*extensions_v1beta1.Ingress)
	return t.translate(action, metrics.MetricEndpoints, &i.CreationTimestamp)
}

func (t *TranslateFuncs) OnService(oldObj, obj interface{}, action EventType) error {
	s := obj.(*core_v1.Service)

	if s.Spec.Type != core_v1.ServiceTypeLoadBalancer {
		return nil
	}

	return t.translate(action, metrics.MetricEndpoints, &s.CreationTimestamp)
}

// NOTE: At this point we log Warning events as errors. For true errors we would
// need to decide what an error actually is (unavailable container or endpoint?
// application that quits unexpectedly?), if it's transient or permanent,
// whether we actually care about it, etc. Then we'd need to analyze the event
// and other resources to figure out what we're dealing with. So, for now, we
// just count warnings.
func (t *TranslateFuncs) OnEvent(oldObj, obj interface{}, action EventType) error {
	e := obj.(*core_v1.Event)

	if action == EventDelete {
		return nil
	}

	// Discard any normal events, and any events that
	// happened before we started watching (to avoid double counting after
	// a restart)
	if e.Type != "Warning" || e.LastTimestamp.Before(&t.startupTime) {
		return nil
	}

	if action == EventUpdate {
		oldE := oldObj.(*core_v1.Event)
		// If count increased, we log another warning
		if oldE.Count == e.Count {
			return nil
		}
	}

	// Get object event references
	ref, exists := t.getReferencedObject(&e.InvolvedObject)
	if !exists {
		return nil
	}

	// Check if the referred object is of interest
	if !isAppInstance(ref) {
		return nil
	}

	kind := e.InvolvedObject.Kind
	log.Debug().Str("object", kind).Str("name", e.InvolvedObject.Name).
		Int32("count", e.Count).Str("reason", e.Reason).Str("message", e.Message).Msg("event")

	switch kind {
	case PodKind.Kind:
		// Filter out references to zt container
		if e.InvolvedObject.FieldPath == "spec.containers{zt-sidecar}" {
			return nil
		}
		fallthrough
	case DeploymentKind.Kind:
		t.collector.Error(metrics.MetricServices)

	case ServiceKind.Kind:
		s := ref.(*core_v1.Service)
		if s.Spec.Type != core_v1.ServiceTypeLoadBalancer {
			return nil
		}
		fallthrough
	case IngressKind.Kind:
		t.collector.Error(metrics.MetricEndpoints)

	case PVCKind.Kind:
		t.collector.Error(metrics.MetricVolumes)

	case NamespaceKind.Kind:
		t.collector.Error(metrics.MetricFragments)
	}

	return nil
}

func (t *TranslateFuncs) getReferencedObject(ref *core_v1.ObjectReference) (interface{}, bool) {
	// If we don't have a store for this kind, we are not intereseted in
	// the object - we also cannot easily retrieve it.
	store, found := t.stores[ref.Kind]
	if !found {
		return nil, false
	}

	var key string
	if len(ref.Namespace) > 0 {
		key = fmt.Sprintf("%s/%s", ref.Namespace, ref.Name)
	} else {
		key = ref.Name
	}

	obj, exists, err := store.GetByKey(key)
	if err != nil {
		log.Error().Err(err).Msg("error retrieving object from resource store")
		return nil, false
	}
	if !exists {
		return nil, false
	}

	return obj, true
}

// Calls create function if after server startup, existing otherwise
func (t *TranslateFuncs) translate(action EventType, metric metrics.MetricType, ts *meta_v1.Time) error {
	switch action {
	case EventAdd:
		if t.startupTime.Before(ts) {
			t.collector.Create(metric)
		} else {
			t.collector.Existing(metric)
		}
	case EventDelete:
		t.collector.Delete(metric)
	}

	return nil
}

// Filter out non-application-instance objects
func isAppInstance(obj interface{}) bool {
	metaobj, err := meta.Accessor(obj)
	if err != nil {
		log.Error().Err(err).Msg("invalid object retrieved from resource store")
		return false
	}
	labels := metaobj.GetLabels()
	_, found := labels[utils.NALEJ_ANNOTATION_ORGANIZATION_ID]
	if !found {
		log.Debug().Msg("no nalej-organization")
		return false
	}

	// filter out unwanted instances
	for k, v := range FilterLabelSet {
		label, found := labels[k]
		if found && label == v {
			return false
		}
	}

	return true
}
