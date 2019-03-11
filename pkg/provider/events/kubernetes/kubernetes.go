/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Kubernetes Events provider

package kubernetes

import (
	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	"github.com/rs/zerolog/log"

        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/tools/clientcmd"
)

// EventsProvider implements the EventsProvider interface; it
// subscribes to Kubernetes events and translates each incoming event to
// a platform metric
type EventsProvider struct {
	// Configuration to create Kubernetes client
	kubeconfig *rest.Config
	// Kubernetes client for event subscription
	client *kubernetes.Clientset

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

func NewEventsProvider(configfile string, incluster bool, collector metrics.Collector) (*EventsProvider, derrors.Error) {
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
		stopChan: make(chan struct{}),
		collector: collector,
	}
	return provider, nil
}

// Start collecting metrics
func (p *EventsProvider) Start() (derrors.Error) {
	log.Info().Msg("starting kubernetes events listener")

	translator, err := NewTranslator(p.collector)
	if err != nil {
		return err
	}

	// Set up watchers
	for _, kind := range(Translatable) {
		watcher, err := NewWatcher(p.kubeconfig, &kind, translator)
		if err != nil {
			p.Stop()
			return err
		}

		watcher.Start(p.stopChan)
	}

	return nil
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
