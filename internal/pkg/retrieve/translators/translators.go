/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translators from specific QueryResult implementations to the
// appropriate grpc.QueryResponse

package translators

import (
	"github.com/nalej/derrors"

	grpc "github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/pkg/provider/query"
)

type TranslatorFunc func(query.QueryResult) (*grpc.QueryResponse, derrors.Error)

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
