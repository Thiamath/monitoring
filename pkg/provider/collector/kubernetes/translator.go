/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates Kubernetes events to Platform counters

package kubernetes

import (
	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"

	apps_v1 "k8s.io/api/apps/v1"
	core_v1 "k8s.io/api/core/v1"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

        "k8s.io/client-go/kubernetes/scheme"
)

// List of objects we are able to translate
var Translatable = []schema.GroupVersionKind{
	apps_v1.SchemeGroupVersion.WithKind("Deployment"),
	core_v1.SchemeGroupVersion.WithKind("Namespace"),
}

// Translator implements k8s.io/client-go/tools/cache.ResourceEventHandler
// so it can be used directly in the informer.
type Translator struct {

}

func NewTranslator() (*Translator, derrors.Error) {
	translator := &Translator{
	}

	return translator, nil
}

func (t *Translator) OnAdd(obj interface{}) {
	// We should be able to cast every object to meta
	meta, ok := obj.(meta_v1.Object)
	if !ok {
		log.Error().Msg("non-kubernetes object received")
		return
	}

	kinds, _, err := scheme.Scheme.ObjectKinds(obj.(runtime.Object))
	if err != nil {
		log.Warn().Str("link", meta.GetSelfLink()).Msg("invalid object received")
		return
	}

	// Not sure what to do if an object matches multiple kinds, let's
	// at least warn
	if len(kinds) > 1 {
		l := log.Warn().Str("link", meta.GetSelfLink())
		for _, k := range(kinds) {
			l = l.Str("candidate", k.String())
		}
		l.Msg("received ambiguous object, picking first candidate")
	}

	kind := kinds[0]
	log.Debug().
		Str("link", meta.GetSelfLink()).
		Str("kind", kind.String()).
		Str("namespace", meta.GetNamespace()).
		Str("name", meta.GetName()).
		Msg("resource added")
}

func (t *Translator) OnUpdate(oldObj, newObj interface{}) {

}

func (t *Translator) OnDelete(obj interface{}) {

}
