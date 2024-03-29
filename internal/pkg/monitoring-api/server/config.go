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
	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/version"
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Debug flag of the run command
	Debug bool
	// GrpcPort where the API service will listen requests.
	GrpcPort int
	// HttpPort where the API service will listen requests.
	HttpPort int
	// UseTLS Use or not TLS.
	UseTLS bool
	// SkipServerCertValidation Don't validate TLS certificates.
	SkipServerCertValidation bool
	// CACertPath Alternative certificate file to use for validation.
	CACertPath string
	// ClientCertPath Client Cert Path.
	ClientCertPath string
	// MonitoringManagerAddress is the address to the monitoring manager service
	MonitoringManagerAddress string
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.GrpcPort <= 0 {
		return derrors.NewInvalidArgumentError("grpc port must be specified")
	}
	if conf.HttpPort <= 0 {
		return derrors.NewInvalidArgumentError("http port must be specified")
	}
	if conf.CACertPath == "" {
		return derrors.NewInvalidArgumentError("caCertPath is required")
	}
	if conf.ClientCertPath == "" {
		return derrors.NewInvalidArgumentError("clientCertPath is required")
	}
	if conf.MonitoringManagerAddress == "" {
		return derrors.NewInvalidArgumentError("monitoringManagerAddress is required")
	}
	return nil
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Int("port", conf.GrpcPort).Msg("gRPC port")
	log.Info().Int("port", conf.HttpPort).Msg("HTTP port")
	log.Info().Bool("tls", conf.UseTLS).Bool("skipServerCertValidation", conf.SkipServerCertValidation).Str("cert", conf.CACertPath).Str("cert", conf.ClientCertPath).Msg("TLS parameters")
	log.Info().Str("MonitoringManagerAddress", conf.MonitoringManagerAddress).Msg("address of the  monitoring manager service")
}
