/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query provider registry

package query

type QueryProviderType string

// A query provider registry translates between a query provider type and
// a QueryProvider implementation instance
type QueryProviderRegistry map[QueryProviderType]QueryProvider

func (r QueryProviderRegistry) Register(p QueryProvider) {
	r[p.Type()] = p
}

func (r QueryProviderRegistry) GetProvider(t QueryProviderType) QueryProvider {
	p, found := r[t]
	if !found {
		return nil
	}
	return p
}

// Default global query provider registry and convenience functions
var DefaultRegistry = QueryProviderRegistry{}

func Register(p QueryProvider) {
	DefaultRegistry.Register(p)
}

func GetProvider(t QueryProviderType) QueryProvider {
	return DefaultRegistry.GetProvider(t)
}
