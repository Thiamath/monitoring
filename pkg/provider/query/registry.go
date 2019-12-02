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

// Query provider registry

package query

import (
	"github.com/nalej/derrors"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type ProviderType string

func (t ProviderType) String() string {
	return string(t)
}

type ProviderConfigFunc func(*cobra.Command) ProviderConfig

type ProviderConfig interface {
	Enabled() bool
	Print(*zerolog.Event)
	Validate() derrors.Error
	NewProvider() (Provider, derrors.Error)
}

// A query provider registry translates between a query provider type and
// its configuration function. The returned configuration can be used
// to create a new instance
type ProviderRegistry map[ProviderType]ProviderConfigFunc

// Created during command initialization
type ProviderConfigs map[ProviderType]ProviderConfig

// Created during service start
type Providers map[ProviderType]Provider

func (r ProviderRegistry) Register(tpe ProviderType, f ProviderConfigFunc) {
	r[tpe] = f
}

func (r ProviderRegistry) NumEntries() int {
	return len(r)
}

// Default global query provider registry and convenience functions
var Registry = ProviderRegistry{}

func Register(tpe ProviderType, f ProviderConfigFunc) {
	Registry.Register(tpe, f)
}
