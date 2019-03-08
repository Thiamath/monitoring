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
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Port where the API service will listen requests.
	Port int
	// Path to kubeconfig
	Kubeconfig string
	// Running inside Kubernetes cluster
	InCluster bool
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}

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
	log.Info().Str("file", conf.Kubeconfig).Bool("in-cluster", conf.InCluster).Msg("kubeconfig")
}
