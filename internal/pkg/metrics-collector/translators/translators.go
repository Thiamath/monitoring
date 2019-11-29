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

// Translators from specific QueryResult implementations to the
// appropriate grpc_monitoring_go.QueryResponse

package translators

import (
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/pkg/provider/query"
)

type TranslatorFunc func(query.QueryResult) (*grpc_monitoring_go.QueryResponse, derrors.Error)

type Translators map[query.QueryProviderType]TranslatorFunc

var DefaultTranslators = Translators{}

func Register(tpe query.QueryProviderType, f TranslatorFunc) {
	DefaultTranslators[tpe] = f
}

func GetTranslator(tpe query.QueryProviderType) (TranslatorFunc, bool) {
	f, found := DefaultTranslators[tpe]
	if !found {
		return nil, false
	}
	return f, true
}
