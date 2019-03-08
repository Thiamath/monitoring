/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translates Kubernetes events to Platform counters

package kubernetes

import (
	"fmt"
	"reflect"

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

type TranslateAction string
const (
	TranslateAdd TranslateAction = "add"
	TranslateUpdate TranslateAction = "update"
	TranslateDelete TranslateAction = "delete"
)
type TranslateFunc func(interface{}, TranslateAction)

// Translator implements k8s.io/client-go/tools/cache.ResourceEventHandler
// so it can be used directly in the informer.
type Translator struct {
	// The mapping from kind to function pointer
	funcMap map[schema.GroupVersionKind]TranslateFunc
}

func NewTranslator() (*Translator, derrors.Error) {
	translator := &Translator{
		funcMap: make(map[schema.GroupVersionKind]TranslateFunc, len(Translatable)),
	}

	// Create the dynamic function map to this exact instance
	for _, kind := range(Translatable) {
		fName := fmt.Sprintf("Translate%s", kind.Kind)
		tValue := reflect.ValueOf(translator)
		fValue := tValue.MethodByName(fName)
		if !fValue.IsValid() {
			return nil, derrors.NewInternalError(fmt.Sprintf("function %s not defined in translator", fName))
		}
		translator.funcMap[kind] = fValue.Interface().(func(interface{}, TranslateAction))
	}

	return translator, nil
}

func (t *Translator) OnAdd(obj interface{}) {
	t.dispatch(obj, TranslateAdd)
}

func (t *Translator) OnUpdate(oldObj, newObj interface{}) {
	t.dispatch(newObj, TranslateUpdate)
}

func (t *Translator) OnDelete(obj interface{}) {
	t.dispatch(obj, TranslateDelete)
}

func (t *Translator) dispatch(obj interface{}, action TranslateAction) {
	// We should be able to cast every object to meta
	meta, ok := obj.(meta_v1.Object)
	if !ok {
		log.Error().Msg("non-kubernetes object received")
		return
	}
	l := log.With().Str("action", string(action)).Str("resource", meta.GetSelfLink()).Logger()

	kinds, _, err := scheme.Scheme.ObjectKinds(obj.(runtime.Object))
	if err != nil {
		l.Warn().Msg("invalid object received")
		return
	}

	// Not sure what to do if an object matches multiple kinds, let's
	// at least warn
	if len(kinds) > 1 {
		kindLog := l.Warn()
		for _, k := range(kinds) {
			kindLog = kindLog.Str("candidate", k.String())
		}
		kindLog.Msg("received ambiguous object, picking first candidate")
	}

	// Dispatch to translator function
	kind := kinds[0]
	f, found := t.funcMap[kind]
	if !found {
		l.Warn().Msg("no translator function found")
		return
	}
	l.Debug().Msg("dispatching")
	f(obj, action)
}

// Translating functions
func (t *Translator) TranslateDeployment(obj interface{}, action TranslateAction) {
	d := obj.(*apps_v1.Deployment)
	log.Debug().Str("name", d.Name).Msg("deployment")
}

func (t *Translator) TranslateNamespace(obj interface{}, action TranslateAction) {
	n := obj.(*core_v1.Namespace)
	log.Debug().Str("name", n.Name).Msg("namespace")
}
