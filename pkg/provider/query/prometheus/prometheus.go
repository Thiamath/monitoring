/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query provider implementation

package prometheus

import (
	"context"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/provider/query"

	"github.com/prometheus/client_golang/api"
	prometheus_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/rs/zerolog/log"
)

type PrometheusProvider struct {
	api prometheus_v1.API
	templates query.TemplateMap
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

	provider := &PrometheusProvider{
		api: prometheus_v1.NewAPI(client),
		templates: make(query.TemplateMap, len(queryTemplates)),
	}

	// Pre-parse templates
	for name, tmplStr := range(queryTemplates) {
		parsed, err := template.New(name.String()).Parse(tmplStr)
		if err != nil {
			return nil, derrors.NewInternalError("failed parsing template", err)
		}
		provider.templates[name] = parsed
	}

	return provider, nil
}

// Returns the query provider type
func (p *PrometheusProvider) ProviderType() query.QueryProviderType {
	return ProviderType
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

func (p *PrometheusProvider) ExecuteTemplate(ctx context.Context, name query.TemplateName, avg time.Duration) (int64, derrors.Error) {
	log.Debug().Str("name", name.String()).Str("avg", avg.String()).Msg("executing template query")

	// TODO: make part of this generic
	tmpl, found := p.templates[name]
	if !found {
		return 0, derrors.NewNotFoundError(fmt.Sprintf("template %s not found", name))
	}

	// Averages only make sense for >2m (else there aren't enough data points)
	if avg.Minutes() <= 2 {
		avg = 0
	}

	// Execute template
	vars := &TemplateVars{
		AvgSeconds: int(avg.Seconds()),
	}
	var buf strings.Builder
	err := tmpl.Execute(&buf, vars)
	if err != nil {
		return 0, derrors.NewInternalError("error executing template", err)
	}

	q := &query.Query{
		QueryString: buf.String(),
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
