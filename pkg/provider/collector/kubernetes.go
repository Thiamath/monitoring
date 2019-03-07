/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Kubernetes Collector provider

package collector

import (
	"github.com/nalej/derrors"

	"github.com/rs/zerolog/log"

        "k8s.io/client-go/kubernetes"
)

// KubeCollectorProvider implements the CollectorProvider interface; it
// subscribes to Kubernetes events and translates each incoming event to
// a platform metric
type KubeCollectorProvider struct {
	// Kubernetes client for event subscription
	client *kubernetes.Clientset

	// Metrics collector
	// TBD
}

/*
        // Create clients
        var kubeconfig *rest.Config
        if s.Configuration.InCluster {
                kubeconfig, err = rest.InClusterConfig()
        } else {
                kubeconfig, err = clientcmd.BuildConfigFromFlags("", s.Configuration.Kubeconfig)
        }
        if err != nil {
                return derrors.NewInternalError("failed to create kubeclient configuration", err)
        }

        kubeclient, err := kubernetes.NewForConfig(kubeconfig)
        if err != nil {
                return derrors.NewInternalError("failed to create kubeclient", err)
        }
        log.Debug().Str("host", kubeconfig.Host).Msg("created kubeclient")
*/

func NewKubeCollectorProvider(configfile string, incluster bool) (*KubeCollectorProvider, derrors.Error) {
	log.Debug().Str("config", configfile).Bool("in-cluster", incluster).Msg("creating kubernetes collector provider")
	return nil, nil
}

// Start collecting metrics
func (kcp *KubeCollectorProvider) Start() (derrors.Error) {
	return nil
}

// Stop collecting metrics
func (kcp *KubeCollectorProvider) Stop() (derrors.Error) {
	return nil
}

// Get specific metrics, or all available when no specific metrics are requested
func (kcp *KubeCollectorProvider) GetMetrics(metrics ...MetricType) (Metrics, derrors.Error) {
	return nil, nil
}
