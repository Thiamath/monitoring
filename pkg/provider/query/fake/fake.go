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

// Fake implementation of query provider interface

package fake

import (
	"context"

	"github.com/nalej/derrors"

	"github.com/nalej/monitoring/pkg/provider/query"
)

const ProviderType query.ProviderType = "FAKE"

type Provider struct {
	queries   map[query.Query]query.Result
	templates map[query.TemplateName]map[query.TemplateVars]int64
}

var Supports = query.ProviderSupport{
	query.FeaturePlatformStats,
	query.FeatureSystemStats,
}

func NewProvider(queries map[query.Query]query.Result, templates map[query.TemplateName]map[query.TemplateVars]int64) (*Provider, derrors.Error) {
	p := &Provider{
		queries:   queries,
		templates: templates,
	}

	return p, nil
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
	res, found := p.queries[*q]
	if !found {
		return nil, derrors.NewNotFoundError("fake provider received unexpected query").WithParams(q)
	}
	return res, nil
}

func (p *Provider) ExecuteTemplate(ctx context.Context, name query.TemplateName, vars *query.TemplateVars) (int64, derrors.Error) {
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
