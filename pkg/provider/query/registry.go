/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query provider registry

package query

import (
	"github.com/nalej/derrors"
	"github.com/spf13/cobra"
	"github.com/rs/zerolog"
)

type QueryProviderType string
func (t QueryProviderType) String() string {
	return string(t)
}

type QueryProviderConfigFunc func(*cobra.Command) QueryProviderConfig

type QueryProviderConfig interface {
	Enabled() bool
	Print(*zerolog.Event)
	Validate() derrors.Error
	NewProvider() (QueryProvider, derrors.Error)
}

// A query provider registry translates between a query provider type and
// its configuration function. The returned configuration can be used
// to create a new instance
type QueryProviderRegistry map[QueryProviderType]QueryProviderConfigFunc

// Created during command initialization
type QueryProviderConfigs map[QueryProviderType]QueryProviderConfig

// Created during service start
type QueryProviders map[QueryProviderType]QueryProvider

func (r QueryProviderRegistry) Register(tpe QueryProviderType, f QueryProviderConfigFunc) {
	r[tpe] = f
}

func (r QueryProviderRegistry) NumEntries() int {
	return len(r)
}

// Default global query provider registry and convenience functions
var Registry = QueryProviderRegistry{}

func Register(tpe QueryProviderType, f QueryProviderConfigFunc) {
	Registry.Register(tpe, f)
}
