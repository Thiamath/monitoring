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

// Query provider interface

package query

import (
	"context"
	"time"

	"github.com/nalej/derrors"
)

type Provider interface {
	// Returns the query provider type
	ProviderType() ProviderType
	// Returns supported features of provider
	Supported() ProviderSupport
	// Execute query q. The response is specific to the query provider
	// but have some common convenience functions to get e.g., the raw
	// values
	Query(ctx context.Context, q *Query) (Result, derrors.Error)
	// We define a number of query templates, referenced by name, that
	// return a single integer and take a single, optional parameter
	// indicating the time range over which the result should be
	// averaged. This function executes such a template using the
	// provider
	ExecuteTemplate(ctx context.Context, name TemplateName, vars *TemplateVars) (int64, derrors.Error)
}

// Types to indicate what a provider supports
type ProviderFeature string

const (
	FeaturePlatformStats ProviderFeature = "platformstats"
	FeatureSystemStats   ProviderFeature = "systemstats"
)

type ProviderSupport []ProviderFeature

func (q ProviderSupport) Supports(f ProviderFeature) bool {
	for _, supported := range q {
		if supported == f {
			return true
		}
	}
	return false
}

// Query descriptor
type Query struct {
	QueryString string
	Range       Range
}

// Time range over which to execute the query. If only Start is provided
// (and End is Time.IsZero()), query is executed for a single point in time;
// in that case Step is ignored.
type Range struct {
	Start, End time.Time
	Step       time.Duration
}

// Query result interface. Type() can be used to call a handler function
// for the specific type at some later point. For convenience, a function
// is provided to get the basic values as well.
type Result interface {
	// Return type of the query response
	ResultType() ProviderType
}
