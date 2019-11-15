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

package static_lister

import (
	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/version"
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Port where the Prometheus endpoint will be served
	Port int
	// Namespace, subsystem and name for the metric that is served.
	// The metric name is namespace_subsystem_name
	Namespace string
	Subsystem string
	Name      string
	// The name of the label that will be set for this series
	LabelName string
	// The file with the values for the label
	LabelFile string
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}

	// Namespace and Subsystem may be empty
	if conf.Name == "" {
		return derrors.NewInvalidArgumentError("name must be specified")
	}

	if conf.LabelName == "" {
		return derrors.NewInvalidArgumentError("label-name must be specified")
	}

	if conf.LabelFile == "" {
		return derrors.NewInvalidArgumentError("label-file must be specified")
	}

	return nil
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Int("port", conf.Port).Msg("metrics endpoint port")
	log.Info().Str("namespace", conf.Namespace).Str("subsystem", conf.Subsystem).Str("name", conf.Name).Msg("metric name")
	log.Info().Str("label", conf.LabelName).Msg("label name")
	log.Info().Str("file", conf.LabelFile).Msg("label values file")
}
