/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Translator for Prometheus query result

package translators

import (
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-utils/pkg/conversions"

	grpc "github.com/nalej/grpc-infrastructure-monitor-go"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query/prometheus"
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
