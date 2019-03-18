/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Dispatches Kubernetes events functions from a DispatchFuncs

package kubernetes

import (
	"fmt"
	"reflect"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"

        "k8s.io/client-go/kubernetes/scheme"
)

type KindList []schema.GroupVersionKind

// Actions that a Kubernetes event can describe
type EventAction string
const (
	EventAdd EventAction = "add"
	EventUpdate EventAction = "update"
	EventDelete EventAction = "delete"
)

// Interface for a collection of dispatcher functions. As an implementation
// can contain an arbirtary number of functions, the only interface-defined
// method is to list the GroupVersionKinds this function collection supports.
type DispatchFuncs interface {
	SupportedKinds() KindList
}

// Function type for the SupportedKinds of a DispatchFuncs collection
type DispatchFunc func(interface{}, EventAction)

// Dispatcher implements k8s.io/client-go/tools/cache.ResourceEventHandler
// so it can be used directly in the informer.
type Dispatcher struct {
	// The mapping from kind to function pointer
	funcMap map[schema.GroupVersionKind]DispatchFunc
}

func NewDispatcher(funcs DispatchFuncs) (*Dispatcher, derrors.Error) {
	kinds := funcs.SupportedKinds()

	dispatcher := &Dispatcher{
		funcMap: make(map[schema.GroupVersionKind]DispatchFunc, len(kinds)),
	}

	funcsValue := reflect.ValueOf(funcs)

	// Create the dynamic function map to this exact instance
	for _, kind := range(kinds) {
		fName := fmt.Sprintf("On%s", kind.Kind)
		fValue := funcsValue.MethodByName(fName)
		if !fValue.IsValid() {
			return nil, derrors.NewInternalError(fmt.Sprintf("function %s not defined in dispatchfuncs", fName))
		}
		dispatcher.funcMap[kind] = fValue.Interface().(func(interface{}, EventAction))
	}

	return dispatcher, nil
}

func (d *Dispatcher) Dispatchable() KindList {
	kinds := make(KindList, 0, len(d.funcMap))
	for k, _ := range(d.funcMap) {
		kinds = append(kinds, k)
	}

	return kinds
}

func (d *Dispatcher) OnAdd(obj interface{}) {
	d.dispatch(obj, EventAdd)
}

func (d *Dispatcher) OnUpdate(oldObj, newObj interface{}) {
	d.dispatch(newObj, EventUpdate)
}

func (d *Dispatcher) OnDelete(obj interface{}) {
	d.dispatch(obj, EventDelete)
}

func (d *Dispatcher) dispatch(obj interface{}, action EventAction) {
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
	f, found := d.funcMap[kind]
	if !found {
		l.Warn().Msg("no translator function found")
		return
	}
	l.Debug().Msg("dispatching")
	f(obj, action)
}