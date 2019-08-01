/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query provider implementation

package prometheus

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/monitoring/pkg/provider/query"

	"github.com/prometheus/client_golang/api"
	prometheus_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

type PrometheusProvider struct {
	api prometheus_v1.API
	templates query.TemplateMap
}

var PrometheusSupports = query.QueryProviderSupport{
	query.FeaturePlatformStats,
	query.FeatureSystemStats,
}

func NewProvider(config *PrometheusConfig) (*PrometheusProvider, derrors.Error) {
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

	provider := &PrometheusProvider{
		api: prometheus_v1.NewAPI(client),
		templates: templates,
	}

	return provider, nil
}

// Returns the query provider type
func (p *PrometheusProvider) ProviderType() query.QueryProviderType {
	return ProviderType
}

func (p *PrometheusProvider) Supported() query.QueryProviderSupport {
	return PrometheusSupports
}

// Execute query q.
func (p *PrometheusProvider) Query(ctx context.Context, q *query.Query) (query.QueryResult, derrors.Error) {
	var val model.Value
	var err error

	// TODO: validate safe query?

	log.Debug().Str("query", q.QueryString).Msg("executing query")
	// Range or instance query
	if q.Range.End.IsZero() {
		// Instance query
		val, err = p.api.Query(ctx, q.QueryString, q.Range.Start)
	} else {
		val, err = p.api.QueryRange(ctx, q.QueryString, prometheus_v1.Range(q.Range))
	}
	if err != nil {
		return nil, derrors.NewInvalidArgumentError("failed executing query", err)
	}

	return NewPrometheusResult(val), nil
}

func (p *PrometheusProvider) ExecuteTemplate(ctx context.Context, name query.TemplateName, vars *query.TemplateVars) (int64, derrors.Error) {
	q, derr := p.templates.GetTemplateQuery(name, vars)
	if derr != nil {
		return 0, derr
	}

	res, derr := p.Query(ctx, q)
	if derr != nil {
		return 0, derr
	}

	val, derr := res.(*PrometheusResult).GetScalarInt()
	if derr != nil {
		return 0, derr
	}

	return val, nil
}
