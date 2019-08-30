/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Wrapper for the configuration properties.

package availability_cli

import (
	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/version"
	"github.com/nalej/monitoring/pkg/provider/query/prometheus"
	"github.com/rs/zerolog/log"
)

// Config struct for the API service.
type Config struct {
	// Show verbose status information
	Verbose bool
	// Prometheus configuration
	Prometheus prometheus.PrometheusConfig
}

// Validate the configuration.
func (conf *Config) Validate() derrors.Error {
	return conf.Prometheus.Validate()
}

// Print the current API configuration to the log.
func (conf *Config) Print() {
	log.Info().Str("app", version.AppVersion).Str("commit", version.Commit).Msg("version")
	log.Info().Bool("flag", conf.Verbose).Msg("verbose")
	conf.Prometheus.Print(log.Info())
}
