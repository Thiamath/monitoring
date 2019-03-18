/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query provider interface

package query

import (
	"context"
	"time"

	"github.com/nalej/derrors"
)

type QueryProvider interface {
	// Returns the query provider type
	ProviderType() QueryProviderType
	// Returns supported features of provider
	Supported() QueryProviderSupport
	// Execute query q. The response is specific to the query provider
	// but have some common convenience functions to get e.g., the raw
	// values
	Query(ctx context.Context, q *Query) (QueryResult, derrors.Error)
	// We define a number of query templates, referenced by name, that
	// return a single integer and take a single, optional parameter
	// indicating the time range over which the result should be
	// averaged. This function executes such a template using the
	// provider
	ExecuteTemplate(ctx context.Context, name TemplateName, vars *TemplateVars) (int64, derrors.Error)
}

// Types to indicate what a provider supports
type QueryProviderFeature string
const (
	FeaturePlatformStats QueryProviderFeature = "platformstats"
	FeatureSystemStats QueryProviderFeature = "systemstats"
)

type QueryProviderSupport []QueryProviderFeature
func (q QueryProviderSupport) Supports(f QueryProviderFeature) bool {
	for _, supported := range(q) {
		if supported == f {
			return true
		}
	}
	return false
}

// Query descriptor
type Query struct {
	QueryString string
	Range QueryRange
}

// Time range over which to execute the query. If only Start is provided
// (and End is Time.IsZero()), query is executed for a single point in time;
// in that case Step is ignored.
type QueryRange struct {
	Start, End time.Time
	Step time.Duration
}

// Query result interface. Type() can be used to call a handler function
// for the specific type at some later point. For convenience, a function
// is provided to get the basic values as well.
type QueryResult interface {
	// Return type of the query response
	ResultType() QueryProviderType
}