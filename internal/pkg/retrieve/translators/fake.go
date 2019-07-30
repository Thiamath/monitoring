/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translator for Prometheus query result

package translators

import (
	"github.com/nalej/derrors"

	grpc "github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/nalej/monitoring/pkg/provider/query/fake"
)

func FakeTranslator(q query.QueryResult) (*grpc.QueryResponse, derrors.Error) {
	result, ok := q.(fake.FakeResult)
	if !ok || result.ResultType() != fake.FakeProviderType {
		return nil, derrors.NewAbortedError("invalid query result type")
	}
	if string(result) == "" {
		return nil, derrors.NewAbortedError("nil query result")
	}

	grpcResponse := &grpc.QueryResponse{
		Type: grpc.QueryType(-1), // FAKE,
		Result: &QueryResponse_FakeResult{Result: string(result)},
	}

	return grpcResponse, nil
}

type QueryResponse_FakeResult struct {
	Result string
	grpc.QueryResponse_PrometheusResult // to make it a valid response
}

func (q *QueryResponse_FakeResult) String() string {
	return q.Result
}
