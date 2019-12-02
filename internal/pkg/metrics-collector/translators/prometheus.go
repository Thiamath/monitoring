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

	"github.com/nalej/grpc-utils/pkg/conversions"

	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/nalej/monitoring/pkg/provider/query/prometheus"
)

func init() {
	Register(prometheus.ProviderType, PrometheusTranslator)
}

// Converts from the internal PrometheusResult to a grpc_monitoring_go.QueryResponse with a grpc_monitoring_go.PrometheusResponse.
func PrometheusTranslator(q query.Result) (*grpc_monitoring_go.QueryResponse, derrors.Error) {
	promResult, ok := q.(*prometheus.Result)
	if !ok || promResult.ResultType() != prometheus.ProviderType {
		return nil, derrors.NewAbortedError("invalid query result type")
	}
	if promResult == nil {
		return nil, derrors.NewAbortedError("nil query result")
	}

	grpcRes := make([]*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue, 0, len(promResult.Values))
	for _, resVal := range promResult.Values {
		grpcValues := make([]*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue_Value, 0, len(resVal.Values))
		for _, val := range resVal.Values {
			grpcVal := &grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue_Value{
				Timestamp: conversions.GRPCTime(val.Timestamp),
				Value:     val.Value,
			}
			grpcValues = append(grpcValues, grpcVal)
		}
		grpcResVal := &grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue{
			Metric: resVal.Labels,
			Value:  grpcValues,
		}
		grpcRes = append(grpcRes, grpcResVal)
	}

	grpcPromResponse := &grpc_monitoring_go.QueryResponse_PrometheusResponse{
		ResultType: grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultType(grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultType_value[promResult.Type.String()]),
		Result:     grpcRes,
	}

	grpcResponse := &grpc_monitoring_go.QueryResponse{
		Type:   grpc_monitoring_go.QueryType_PROMETHEUS,
		Result: &grpc_monitoring_go.QueryResponse_PrometheusResult{PrometheusResult: grpcPromResponse},
	}

	return grpcResponse, nil
}
