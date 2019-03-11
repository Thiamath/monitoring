/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Watcher handles events from informers

package kubernetes

import (
	"fmt"

	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"
        "k8s.io/client-go/kubernetes/scheme"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/tools/cache"
)

// Watcher sets up and manages the lifecycle of an informer; it deals with
// events and makes sure the appriate event handler is called.
// See NOTE in kubernetes.go
type Watcher struct {
	// The GroupVersionKind we're creating a watcher for
	gvk *schema.GroupVersionKind
	// The informer controller
	controller cache.Controller
}

func NewWatcher(config *rest.Config, gvk *schema.GroupVersionKind, handler cache.ResourceEventHandler, labelSelector string) (*Watcher, derrors.Error) {
	log.Debug().Str("kind", gvk.String()).Msg("new watcher")

	// Create client
	client, derr := createClient(config, gvk.GroupVersion())
	if derr != nil {
		return nil, derr
	}

	// Create empty object
	objType, err := scheme.Scheme.New(*gvk)
	if err != nil {
		return nil, derrors.NewInternalError(fmt.Sprintf("failed creating object for %s", gvk.String()), err)
	}

	// Check selectors
	parsedLabelSelector, err := labels.Parse(labelSelector)
	if err != nil {
		return nil, derrors.NewInternalError("failed parsing label selector", err)
	}

	// Create a lister-watcher
	// We're adding an 's' to the kind - REST API seems to have e.g., the
	// endpoint "deployments" for the kind "deployment". Not sure how to do
	// this properly - we'll see when this fails.
	optionsModifier := func(options *meta_v1.ListOptions) {
		options.FieldSelector = fields.Everything().String()
		options.LabelSelector = parsedLabelSelector.String()
	}
	watchlist := cache.NewFilteredListWatchFromClient(client, gvk.Kind + "s", meta_v1.NamespaceAll, optionsModifier)

	// Create an informer
	_ /* store */, controller := cache.NewInformer(watchlist, objType, 0 /* No resync */, handler)

	watcher := &Watcher{
		gvk: gvk,
		controller: controller,
	}

	return watcher, nil
}

func (w *Watcher) Start(stopChan <-chan struct{}) {
	log.Debug().Str("resource", w.gvk.String()).Msg("starting watcher")
	go w.controller.Run(stopChan)

	// TODO: do we need to wait for sync even if we don't have a queue
	// and index? Seems to me we want all events anyway, and we have
	// no queue to start after we synced.
	// if !cache.WaitForSync
}

func createClient(config *rest.Config, gv schema.GroupVersion) (rest.Interface, derrors.Error) {
	// Create shallow copy
	c := *config

	c.GroupVersion = &gv

	// The core api has no group and has a slightly different base URL
	if gv.Group != "" {
		c.APIPath = "/apis"
	} else {
		c.APIPath = "/api"
	}
	c.NegotiatedSerializer = serializer.DirectCodecFactory{CodecFactory: scheme.Codecs}

	if c.UserAgent == "" {
		c.UserAgent = rest.DefaultKubernetesUserAgent()
	}

	client, err := rest.RESTClientFor(&c)
	if err != nil {
		return nil, derrors.NewInternalError(fmt.Sprintf("failed creating kubernetes client for %s", gv.String()), err)
	}

	return client, nil
}
