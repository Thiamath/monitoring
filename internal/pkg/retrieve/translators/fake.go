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
		Type:   grpc.QueryType(-1), // FAKE,
		Result: &QueryResponse_FakeResult{Result: string(result)},
	}

	return grpcResponse, nil
}

type QueryResponse_FakeResult struct {
	Result                              string
	grpc.QueryResponse_PrometheusResult // to make it a valid response
}

func (q *QueryResponse_FakeResult) String() string {
	return q.Result
}
