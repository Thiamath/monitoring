/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Wrapper for the configuration properties.

package coord

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Port where the API service will listen requests.
	Port int
	// Address with host:port of the ElasticSearch server
	SystemModelAddress string
	// Prefix for application cluster hostnames
	AppClusterPrefix string
	// Port used by app-cluster-api
	AppClusterPort int
	// Use TLS
	UseTLS bool
	// Don't validate TLS certificates
	Insecure bool
	// Alternative certificate file to use for validation
	CACert string
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	if conf.Port <= 0 {
		return derrors.NewInvalidArgumentError("port must be specified")
	}
	if conf.SystemModelAddress == "" {
		return derrors.NewInvalidArgumentError("systemModelAddress is required")
	}
	if conf.AppClusterPort <= 0 {
		return derrors.NewInvalidArgumentError("appClusterPort is required")
	}
	return nil
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Int("port", conf.Port).Msg("gRPC port")
	log.Info().Str("URL", conf.SystemModelAddress).Msg("systemModelAddress")
	log.Info().Str("prefix", conf.AppClusterPrefix).Msg("appClusterPrefix")
	log.Info().Int("port", conf.AppClusterPort).Msg("appClusterPort")
	log.Info().Bool("tls", conf.UseTLS).Bool("insecure", conf.Insecure).Str("cert", conf.CACert).Msg("TLS parameters")
}
