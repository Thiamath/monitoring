/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Kubernetes Collector provider

package kubernetes

import (
	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-monitor/pkg/provider/collector"

	"github.com/rs/zerolog/log"

        "k8s.io/client-go/kubernetes"
        "k8s.io/client-go/rest"
        "k8s.io/client-go/tools/clientcmd"
)

// CollectorProvider implements the CollectorProvider interface; it
// subscribes to Kubernetes events and translates each incoming event to
// a platform metric
type CollectorProvider struct {
	// Configuration to create Kubernetes client
	kubeconfig *rest.Config
	// Kubernetes client for event subscription
	client *kubernetes.Clientset

	// Channel to stop informers. Close to stop.
	stopChan chan struct{}

	// Metrics collector
	collector *collector.Collector
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

func NewCollectorProvider(configfile string, incluster bool) (*CollectorProvider, derrors.Error) {
	log.Debug().Str("config", configfile).Bool("in-cluster", incluster).Msg("creating kubernetes collector provider")

	collector, derr := collector.NewCollector()
	if derr != nil {
		return nil, derr
	}

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

	provider := &CollectorProvider{
		kubeconfig: kubeconfig,
		stopChan: make(chan struct{}),
		collector: collector,
	}
	return provider, nil
}

// Start collecting metrics
func (c *CollectorProvider) Start() (derrors.Error) {
	log.Info().Msg("starting kubernetes collector")

	translator, err := NewTranslator(c.collector)
	if err != nil {
		return err
	}

	// Set up watchers
	for _, kind := range(Translatable) {
		watcher, err := NewWatcher(c.kubeconfig, &kind, translator)
		if err != nil {
			c.Stop()
			return err
		}

		watcher.Start(c.stopChan)
	}

	return nil
}

// Stop collecting metrics
func (c *CollectorProvider) Stop() (derrors.Error) {
	log.Info().Msg("stopping kubernetes collector")
	// Stop informers
	close(c.stopChan)
	return nil
}

// Get specific metrics, or all available when no specific metrics are requested
func (c *CollectorProvider) GetMetrics(metrics ...collector.MetricType) (collector.Metrics, derrors.Error) {
	return nil, nil
}
