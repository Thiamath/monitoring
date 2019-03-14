/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Wrapper for the configuration properties.

package slave

import (
	"fmt"
	"os"

	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-monitor/version"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Port where the API service will listen requests.
	Port int
	// Port where the metrics endpoint is served over HTTP
	MetricsPort int
	// Path to kubeconfig
	Kubeconfig string
	// Running inside Kubernetes cluster
	InCluster bool

	// Retrieval backends
	QueryProviders query.QueryProviderConfigs
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}

	if conf.MetricsPort <= 0 {
		return derrors.NewInvalidArgumentError("metricsPort must be specified")
	}

	// Retrieval backends validation
	for _, queryConfig := range(conf.QueryProviders) {
		derr := queryConfig.Validate()
		if derr != nil {
			return derr
		}
	}

	// NOTE: All validation except kubeconfig should go before this line

	if conf.InCluster {
		return nil
	}

	// Not in cluster, check kube config
	if conf.Kubeconfig == "" {
		return derrors.NewInvalidArgumentError("one of in-cluster or kubeconfig should be specified")
	}

	f, err := os.Open(conf.Kubeconfig)
	if err != nil {
		return derrors.NewInvalidArgumentError(fmt.Sprintf("cannot open kubeconfig %s", conf.Kubeconfig), err)
	}
	f.Close()

	return nil
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Int("port", conf.MetricsPort).Msg("metrics port")
	log.Info().Str("file", conf.Kubeconfig).Bool("in-cluster", conf.InCluster).Msg("kubeconfig")

	// Retrieval backends
	for _, queryConfig := range(conf.QueryProviders) {
		queryConfig.Print(log.Info())
	}
}
