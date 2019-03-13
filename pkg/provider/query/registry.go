/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query provider registry

package query

type QueryProviderType string
func (t QueryProviderType) String() string {
	return string(t)
}

/* NOTE

init in provider calls Register

RegistryEntry {
	configFunc Config
	initFunc(Config)
}
*r Config(cobra app) {
	r.Config = configFunc(app)
}

*c Enabled() bool {}

in init() call config and create app config map[type]config ProviderConfigs
in config call print/enabled
in run call init(config) and create map[type]instance ProviderInstances
*/

// A query provider registry translates between a query provider type and
// a QueryProvider implementation instance
type QueryProviderRegistry map[QueryProviderType]QueryProvider

func (r QueryProviderRegistry) Register(p QueryProvider) {
	r[p.ProviderType()] = p
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
