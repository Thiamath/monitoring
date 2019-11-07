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
 *
 */

// Translator for Prometheus query result

package translators

import (
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-utils/pkg/conversions"

	grpc "github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/nalej/monitoring/pkg/provider/query/prometheus"
)

func init() {
	Register(prometheus.ProviderType, PrometheusTranslator)
}

// Converts from the internal PrometheusResult to a grpc.QueryResponse with a grpc.PrometheusResponse.
func PrometheusTranslator(q query.QueryResult) (*grpc.QueryResponse, derrors.Error) {
	promResult, ok := q.(*prometheus.PrometheusResult)
	if !ok || promResult.ResultType() != prometheus.ProviderType {
		return nil, derrors.NewAbortedError("invalid query result type")
	}
	if promResult == nil {
		return nil, derrors.NewAbortedError("nil query result")
	}

	grpcRes := make([]*grpc.QueryResponse_PrometheusResponse_ResultValue, 0, len(promResult.Values))
	for _, resVal := range(promResult.Values) {
		grpcValues := make([]*grpc.QueryResponse_PrometheusResponse_ResultValue_Value, 0, len(resVal.Values))
		for _, val := range(resVal.Values) {
			grpcVal := &grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
				Timestamp: conversions.GRPCTime(val.Timestamp),
				Value: val.Value,
			}
			grpcValues = append(grpcValues, grpcVal)
		}
		grpcResVal := &grpc.QueryResponse_PrometheusResponse_ResultValue{
			Metric: resVal.Labels,
			Value: grpcValues,
		}
		grpcRes = append(grpcRes, grpcResVal)
	}

	grpcPromResponse := &grpc.QueryResponse_PrometheusResponse{
		ResultType: grpc.QueryResponse_PrometheusResponse_ResultType(grpc.QueryResponse_PrometheusResponse_ResultType_value[promResult.Type.String()]),
		Result: grpcRes,
	}

	grpcResponse := &grpc.QueryResponse{
		Type: grpc.QueryType_PROMETHEUS,
		Result: &grpc.QueryResponse_PrometheusResult{grpcPromResponse},
	}

	return grpcResponse, nil
}
