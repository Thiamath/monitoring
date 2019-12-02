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

// Prometheus query provider implementation

package prometheus

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/monitoring/pkg/provider/query"

	"github.com/prometheus/client_golang/api"
	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

type Provider struct {
	api       v1.API
	templates query.TemplateMap
}

var Supports = query.ProviderSupport{
	query.FeaturePlatformStats,
	query.FeatureSystemStats,
}

func NewProvider(config *Config) (*Provider, derrors.Error) {
	log.Debug().Str("url", config.Url).Str("type", string(ProviderType)).Msg("creating query provider")
	// Create API client
	client, err := api.NewClient(api.Config{
		Address: config.Url,
	})
	if err != nil {
		return nil, derrors.NewUnavailableError("failed creating prometheus client", err)
	}

	templates, derr := queryTemplates.ParseTemplates()
	if derr != nil {
		return nil, derr
	}

	provider := &Provider{
		api:       v1.NewAPI(client),
		templates: templates,
	}

	return provider, nil
}

// Returns the query provider type
func (p *Provider) ProviderType() query.ProviderType {
	return ProviderType
}

func (p *Provider) Supported() query.ProviderSupport {
	return Supports
}

// Execute query q.
func (p *Provider) Query(ctx context.Context, q *query.Query) (query.Result, derrors.Error) {
	var val model.Value
	var err error

	// TODO: validate safe query?

	log.Debug().Str("query", q.QueryString).Msg("executing query")
	// Range or instance query
	if q.Range.End.IsZero() {
		// Instance query
		val, err = p.api.Query(ctx, q.QueryString, q.Range.Start)
	} else {
		val, err = p.api.QueryRange(ctx, q.QueryString, v1.Range(q.Range))
	}
	if err != nil {
		return nil, derrors.NewInvalidArgumentError("failed executing query", err)
	}

	return NewPrometheusResult(val), nil
}

func (p *Provider) ExecuteTemplate(ctx context.Context, name query.TemplateName, vars *query.TemplateVars) (int64, derrors.Error) {
	q, derr := p.templates.GetTemplateQuery(name, vars)
	if derr != nil {
		return 0, derr
	}

	res, derr := p.Query(ctx, q)
	if derr != nil {
		return 0, derr
	}

	val, derr := res.(*Result).GetScalarInt()
	if derr != nil {
		return 0, derr
	}

	return val, nil
}
