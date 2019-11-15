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
