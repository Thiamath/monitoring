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
	for k, v := range(p.queries) {
		if k == *q {
			return v, nil
		}
	}

	return nil, derrors.NewNotFoundError("fake provider received unexpected query").WithParams(q)
}

func (p *FakeProvider) ExecuteTemplate(ctx context.Context, name query.TemplateName, vars *query.TemplateVars) (int64, derrors.Error) {
	for k, v := range(p.templates) {
		if k == name {
			for kv, res := range(v) {
				if kv == *vars {
					return res, nil
				}
			}
		}
	}

	return 0, derrors.NewNotFoundError("fake provider received unexpected template request").WithParams(name, *vars)
}
