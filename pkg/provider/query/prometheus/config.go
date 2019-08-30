/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query provider config

package prometheus

import (
	"net/url"

	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/spf13/cobra"
	"github.com/rs/zerolog"
)

const ProviderType query.QueryProviderType = "PROMETHEUS"

type PrometheusConfig struct {
	Enable bool
	Url string
	Username string
	Password string
}

func NewPrometheusConfig(cmd *cobra.Command) query.QueryProviderConfig {
	c := &PrometheusConfig{}

	cmd.Flags().BoolVar(&c.Enable, "retrieve.prometheus.enabled", false, "Enable Prometheus retrieval backend")
	cmd.Flags().StringVar(&c.Url, "retrieve.prometheus.url", "http://localhost:9090", "Prometheus retrieval backend URL")
	cmd.Flags().StringVar(&c.Url, "retrieve.prometheus.username", "", "Prometheus retrieval backend username")
	cmd.Flags().StringVar(&c.Url, "retrieve.prometheus.password", "", "Prometheus retrieval backend password")

	return c
}

func (c *PrometheusConfig) Enabled() bool {
	return c.Enable
}

func (c *PrometheusConfig) Print(log *zerolog.Event) {
	l := log.Bool("enabled", c.Enable).Str("url", c.Url)
	if c.Username != "" {
		l = l.Str("username", c.Username)
	}
	if c.Password != "" {
		l = l.Str("password", "********")
	}
	l.Msg("prometheus retrieval backend")
}

func (c *PrometheusConfig) Validate() derrors.Error {
	// Disabled is always ok
	if !c.Enabled() {
		return nil
	}

	_, err := url.ParseRequestURI(c.Url)
	if err != nil {
		return derrors.NewInvalidArgumentError("invalid url", err)
	}

	if c.Username != "" && c.Password == "" {
		return derrors.NewInvalidArgumentError("no password provided")
	}
	if c.Username == "" && c.Password != "" {
		return derrors.NewInvalidArgumentError("no username provided")
	}

	return nil
}

func (c *PrometheusConfig) NewProvider() (query.QueryProvider, derrors.Error) {
	if !c.Enabled() {
		return nil, derrors.NewInternalError("cannot create a disabled query provider")
	}
	return NewProvider(c)
}

func init() {
	query.Register(ProviderType, NewPrometheusConfig)
}
