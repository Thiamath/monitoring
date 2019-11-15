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

// Prometheus query provider config

package prometheus

import (
	"net/url"

	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

const ProviderType query.QueryProviderType = "PROMETHEUS"

type PrometheusConfig struct {
	Enable bool
	Url    string
}

func NewPrometheusConfig(cmd *cobra.Command) query.QueryProviderConfig {
	c := &PrometheusConfig{}

	cmd.Flags().BoolVar(&c.Enable, "retrieve.prometheus.enabled", false, "Enable Prometheus retrieval backend")
	cmd.Flags().StringVar(&c.Url, "retrieve.prometheus.url", "http://localhost:9090", "Prometheus retrieval backend URL")

	return c
}

func (c *PrometheusConfig) Enabled() bool {
	return c.Enable
}

func (c *PrometheusConfig) Print(log *zerolog.Event) {
	log.Bool("enabled", c.Enable).Str("url", c.Url).Msg("prometheus retrieval backend")
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
