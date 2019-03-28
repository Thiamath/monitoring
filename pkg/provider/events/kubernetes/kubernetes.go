/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Kubernetes Events provider

package kubernetes

import (
	"fmt"

	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	"github.com/rs/zerolog/log"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/runtime/serializer"

        "k8s.io/client-go/discovery"
        "k8s.io/client-go/kubernetes/scheme"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/restmapper"
        "k8s.io/client-go/tools/clientcmd"
)

// EventsProvider implements the EventsProvider interface; it
// subscribes to Kubernetes events and translates each incoming event to
// a platform metric
type EventsProvider struct {
	// Configuration to create Kubernetes client
	kubeconfig *rest.Config

	// Cached Kubernetes clients for event subscription
	clients map[schema.GroupVersion]rest.Interface

	// Filters for the watchers
	labelSelector string

	// Channel to stop informers. Close to stop.
	stopChan chan struct{}

	// Metrics collector
	collector metrics.Collector
}

// NOTE: There is a simple example on how to deal with Kubernetes events here:
//   https://rsmitty.github.io/Kubernetes-Events/
// There is a more complex example here:
//   https://engineering.bitnami.com/articles/kubewatch-an-example-of-kubernetes-custom-controller.html
//
// At this moment I don't think we need a workqueue and a rate limiter - we
// need all the events and the work we do really isn't that much (we pretty
// much just determine event type and increase a counter). Furthermore, I don't
// think we need to bother with a shared informer, as we only have one handler
// per informer anyway; nor do we need an index as we don't have a queue and
// can pass the object straight to the handler.
// We can introduce these concepts when needed for optimization.
//
// An extensive description of the event mechanism can be found here:
//   https://lairdnelson.wordpress.com/2018/01/07/understanding-kubernetes-tools-cache-package-part-0/

func NewEventsProvider(configfile string, incluster bool, labelSelector string, collector metrics.Collector) (*EventsProvider, derrors.Error) {
	log.Debug().Str("config", configfile).Bool("in-cluster", incluster).Msg("creating kubernetes events provider")

        var kubeconfig *rest.Config
	var err error
	if incluster {
		kubeconfig, err = rest.InClusterConfig()
	} else {
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", configfile)
	}
	if err != nil {
		return nil, derrors.NewInternalError("failed to create kubeclient configuration", err)
	}
        log.Info().Str("host", kubeconfig.Host).Msg("created kubeconfig")

	provider := &EventsProvider{
		kubeconfig: kubeconfig,
		clients: map[schema.GroupVersion]rest.Interface{},
		labelSelector: labelSelector,
		stopChan: make(chan struct{}),
		collector: collector,
	}
	return provider, nil
}

// Start collecting metrics
func (p *EventsProvider) Start() (derrors.Error) {
	log.Info().Msg("starting kubernetes events listener")

	translator := NewTranslateFuncs(p.collector)
	dispatcher, derr := NewDispatcher(translator)
	if derr != nil {
		return derr
	}

	// Create discovery client
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(p.kubeconfig)
	if err != nil {
		return derrors.NewInternalError("failed to create discovery client", err)
	}
	resources, err := restmapper.GetAPIGroupResources(discoveryClient)
	if err != nil {
		return derrors.NewInternalError("failed to get api group resources", err)
	}
	mapper := restmapper.NewDiscoveryRESTMapper(resources)

	// Set up watchers
	for _, kind := range(dispatcher.Dispatchable()) {
		// Create cached client
		client, derr := p.createClient(kind.GroupVersion())
		if derr != nil {
			p.Stop()
			return derr
		}

		// Figure out resource with RESTMapper
		mapping, err := mapper.RESTMapping(kind.GroupKind(), kind.Version)
		if err != nil {
			p.Stop()
			return derrors.NewInternalError("unable to get rest mapping", err)
		}
		resource := mapping.Resource.Resource

		watcher, derr := NewWatcher(client, &kind, resource, dispatcher, p.labelSelector)
		if derr != nil {
			p.Stop()
			return derr
		}

		// Link the client resource store to the translator for
		// cross-referencing objects
		err = translator.SetStore(kind, watcher.GetStore())
		if err != nil {
			p.Stop()
			return derrors.NewAlreadyExistsError("store already set", err)
		}

		watcher.Start(p.stopChan)
	}

	return nil
}

func (p *EventsProvider) createClient(gv schema.GroupVersion) (rest.Interface, derrors.Error) {
	client, found := p.clients[gv]
	if found {
		return client, nil
	}

	log.Debug().Str("gv", gv.String()).Msg("creating new client")

	// Create shallow copy
	c := *p.kubeconfig

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

	p.clients[gv] = client
	return client, nil
}

// Stop collecting metrics
func (p *EventsProvider) Stop() (derrors.Error) {
	log.Info().Msg("stopping kubernetes collector")
	// Stop informers
	close(p.stopChan)
	return nil
}

// Get specific metrics, or all available when no specific metrics are requested
func (p *EventsProvider) GetMetrics(types ...metrics.MetricType) (metrics.Metrics, derrors.Error) {
	return p.collector.GetMetrics(types...)
}
