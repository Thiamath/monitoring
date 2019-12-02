/*
 * Copyright 2019 Nalej
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Wrapper for the configuration properties.

package server

import (
	"fmt"
	"os"

	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/nalej/monitoring/version"
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

	// Retrieval backends
	QueryProviders query.ProviderConfigs
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}

	// Retrieval backends validation
	for _, queryConfig := range conf.QueryProviders {
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
	_ = f.Close()

	return nil
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("file", conf.Kubeconfig).Bool("in-cluster", conf.InCluster).Msg("kubeconfig")

	// Retrieval backends
	for _, queryConfig := range conf.QueryProviders {
		queryConfig.Print(log.Info())
	}
}
