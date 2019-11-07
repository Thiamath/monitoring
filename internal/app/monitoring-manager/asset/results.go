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

package asset

import (
	"sort"

	"github.com/nalej/derrors"

	"github.com/nalej/grpc-monitoring-go"

	"github.com/rs/zerolog/log"
)

// QueryResults is a mapping from metric to values, where values is a mapping
// from timestamp to value and count. This last mapping is needed for merging
// results from multiple edge controllers. We will convert to one
// QueryMetricsResult to return when we've collected all metrics.
type QueryResults map[string]map[int64]*grpc_monitoring_go.QueryMetricsResult_Value

func NewQueryResults() QueryResults {
	return make(QueryResults)
}

func (r QueryResults) AddResult(ecId string, result *grpc_monitoring_go.QueryMetricsResult) {
	// Loop over all returned metrics
	for metric, assetMetrics := range(result.GetMetrics()) {
		// Currently, we always have a single assetMetric
		assetMetricList := assetMetrics.GetMetrics()
		if len(assetMetricList) == 0 {
			continue
		}
		if len(assetMetricList) > 1 {
			log.Warn().Msg("received query result for more than one individual asset - not supported")
		}

		assetMetric := assetMetricList[0]
		assetMetricValues := assetMetric.GetValues()

		// Create or retrieve map for this metric
		values, found := r[metric]
		if !found {
			values = make(map[int64]*grpc_monitoring_go.QueryMetricsResult_Value, len(assetMetricValues))
			r[metric] = values
		}

		log.Debug().Str("ecid", ecId).Str("metric", metric).Int("count", len(assetMetricValues)).Msg("storing metrics")
		for _, assetMetricValue := range assetMetricValues {
			// Create or add to value for this timestamp
			timestamp := assetMetricValue.GetTimestamp()
			value, found := values[timestamp]
			if !found {
				values[timestamp] = assetMetricValue
			} else {
				value.Value += assetMetricValue.Value
				value.AssetCount += assetMetricValue.AssetCount
			}
		}
	}
}

// Unify the summed results with the timestamp-to-value map into the gRPC result.
func (r QueryResults) GetQueryMetricsResult(aggregationType grpc_monitoring_go.AggregationType) (*grpc_monitoring_go.QueryMetricsResult, derrors.Error) {
	metricResults := make(map[string]*grpc_monitoring_go.QueryMetricsResult_AssetMetrics, len(r))
	for metric, valueMap := range(r) {
		// Make a list sorted timestamps
		keys := make([]int64, 0, len(valueMap))
		for key := range(valueMap) {
			keys = append(keys, key)
		}
		// Sort int64
		sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })

		log.Debug().Str("metric", metric).Int("count", len(keys)).Msg("aggregating metrics")
		// Create final value list, applying aggregation if needed
		values := make([]*grpc_monitoring_go.QueryMetricsResult_Value, 0, len(keys))
		for _, key := range(keys) {
			value := valueMap[key]
			switch aggregationType {
			case grpc_monitoring_go.AggregationType_SUM:
				// Nothing to do - it's already a sum
			case grpc_monitoring_go.AggregationType_AVG:
				value.Value = value.Value / value.AssetCount
			default:
				return nil, derrors.NewInvalidArgumentError("unknown aggregation type").WithParams(aggregationType)
			}
			values = append(values, value)
		}

		metricResults[metric] = &grpc_monitoring_go.QueryMetricsResult_AssetMetrics{
			Metrics: []*grpc_monitoring_go.QueryMetricsResult_AssetMetricValues{
				&grpc_monitoring_go.QueryMetricsResult_AssetMetricValues{
					// We don't have an asset id - if we had a single
					// asset, we would have had a single EC, and we
					// would have returned already.
					Values: values,
					Aggregation: aggregationType,
				},
			},
		}
	}

	result := &grpc_monitoring_go.QueryMetricsResult{
		Metrics: metricResults,
	}
	return result, nil
}
