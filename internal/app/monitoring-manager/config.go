/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Wrapper for the configuration properties.

package monitoring_manager

import (
	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/version"
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Port where the API service will listen requests.
	Port int
	// Address with host:port of the system model component
	SystemModelAddress string
	// EdgeInventoryProxyAddress with host:port of the edge inventory proxy
	EdgeInventoryProxyAddress string
	// Prefix for application cluster hostnames
	AppClusterPrefix string
	// Port used by app-cluster-api
	AppClusterPort int
	// Use TLS
	UseTLS bool
	// Don't validate TLS certificates
	SkipServerCertValidation bool
	// Alternative certificate file to use for validation
	CACertPath string
	// Client Cert Path
	ClientCertPath string
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}
	if conf.SystemModelAddress == "" {
		return derrors.NewInvalidArgumentError("systemModelAddress is required")
	}
	if conf.EdgeInventoryProxyAddress == "" {
		return derrors.NewInvalidArgumentError("edgeInventoryProxyAddress is required")
	}
	if conf.AppClusterPort <= 0 {
		return derrors.NewInvalidArgumentError("appClusterPort is required")
	}
	if conf.CACertPath == "" {
		return derrors.NewInvalidArgumentError("caCertPath is required")
	}
	if conf.ClientCertPath == "" {
		return derrors.NewInvalidArgumentError("clientCertPath is required")
	}
	return nil
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("systemModelAddress")
	log.Info().Str("URL", conf.EdgeInventoryProxyAddress).Msg("edgeInventoryProxyAddress")
	log.Info().Str("prefix", conf.AppClusterPrefix).Msg("appClusterPrefix")
	log.Info().Int("port", conf.AppClusterPort).Msg("appClusterPort")
	log.Info().Bool("tls", conf.UseTLS).Bool("insecure", conf.SkipServerCertValidation).Str("cert", conf.CACertPath).Str("cert", conf.ClientCertPath).Msg("TLS parameters")
}
