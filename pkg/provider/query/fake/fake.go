/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Fake implementation of query provider interface

package fake

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
)

const FakeProviderType query.QueryProviderType = "FAKE"

type FakeProvider struct {
	queries map[query.Query]query.QueryResult
	templates map[query.TemplateName]map[query.TemplateVars]int64
}


var FakeProviderSupports = query.QueryProviderSupport{
	query.FeaturePlatformStats,
	query.FeatureSystemStats,
}

func NewFakeProvider(queries map[query.Query]query.QueryResult, templates map[query.TemplateName]map[query.TemplateVars]int64) (*FakeProvider, derrors.Error) {
	p := &FakeProvider{
		queries: queries,
		templates: templates,
	}

	return p, nil
}

// Returns the query provider type
func (p *FakeProvider) ProviderType() query.QueryProviderType {
	return FakeProviderType
}

func (p *FakeProvider) Supported() query.QueryProviderSupport {
	return FakeProviderSupports
}

// Execute query q.
func (p *FakeProvider) Query(ctx context.Context, q *query.Query) (query.QueryResult, derrors.Error) {
	res, found := p.queries[*q]
	if !found {
		return nil, derrors.NewNotFoundError("fake provider received unexpected query").WithParams(q)
	}
	return res, nil
}

func (p *FakeProvider) ExecuteTemplate(ctx context.Context, name query.TemplateName, vars *query.TemplateVars) (int64, derrors.Error) {
	knownvars, found := p.templates[name]
	if !found {
		return 0, derrors.NewNotFoundError("fake provider received unexpected template name").WithParams(name)
	}

	res, found := knownvars[*vars]
	if !found {
		return 0, derrors.NewNotFoundError("fake provider received unexpected template vars").WithParams(name, *vars)
	}

	return res, nil
}
