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
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-edge-inventory-proxy-go"
	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/grpc-utils/pkg/conversions"

	"github.com/rs/zerolog/log"
)

type Manager struct {
	proxyClient grpc_edge_inventory_proxy_go.EdgeControllerProxyClient
	assetsClient grpc_inventory_go.AssetsClient
	controllersClient grpc_inventory_go.ControllersClient
}

func NewManager(proxyClient grpc_edge_inventory_proxy_go.EdgeControllerProxyClient, assetsClient grpc_inventory_go.AssetsClient, controllersClient grpc_inventory_go.ControllersClient) (*Manager, derrors.Error) {
	m := &Manager{
		proxyClient: proxyClient,
		assetsClient: assetsClient,
		controllersClient: controllersClient,
	}

	return m, nil
}

const edgeControllerAliveTimeout = 600

func (m *Manager) ListMetrics(selector *grpc_inventory_go.AssetSelector) (*grpc_monitoring_go.MetricsList, error) {
	log.Debug().Interface("selector", selector).Msg("ListMetrics received")
	// Get a selector for each relevant Edge Controller
	selectorsFactory := NewSelectorMapFactory(m.assetsClient, m.controllersClient)
	selectors, derr := selectorsFactory.SelectorMap(selector)
	if derr != nil {
		return nil, conversions.ToGRPCError(derr)
	}

	metrics := make(map[string]bool)

	// Create a request for each Edge Controller and execute
	for _, proxyRequest := range(selectors) {
		ecId := proxyRequest.GetEdgeControllerId()
		log.Debug().Interface("request", proxyRequest).Msg("proxy request for ListMetrics")
		ctx, cancel := ProxyContext() // Manual calling cancel to avoid big list of defers
		list, err := m.proxyClient.ListMetrics(ctx, proxyRequest)
		cancel()
		if err != nil {
			// We still want to query to working edge controllers
			log.Warn().Str("edge-controller-id", ecId).Err(err).Msg("failed calling ListMetrics")
			continue
		}
		for _, metric := range(list.GetMetrics()) {
			metrics[metric] = true
		}
	}

	// Unify the results
	metricsList := make([]string, 0, len(metrics))
	for metric := range(metrics) {
		metricsList = append(metricsList, metric)
	}

	return &grpc_monitoring_go.MetricsList{
		Metrics: metricsList,
	}, nil
}

func (m *Manager) QueryMetrics(request *grpc_monitoring_go.QueryMetricsRequest) (*grpc_monitoring_go.QueryMetricsResult, error) {
	log.Debug().Interface("request", request).Msg("QueryMetrics received")
	// Get a selector for each relevant Edge Controller
	selectorsFactory := NewSelectorMapFactory(m.assetsClient, m.controllersClient)
	selectors, derr := selectorsFactory.SelectorMap(request.GetAssets())
	if derr != nil {
		return nil, conversions.ToGRPCError(derr)
	}

	aggregationType := request.GetAggregation()
	// If we're going to calculate an average, we actually need the
	// sum. We can recreate the sum by multiplying the average with
	// the number of assets, or we can just ask for the sum. When
	// we process all retrieved metrics, we'll do the division.
	// If we only query a single edge controller we don't need to
	// do post-processing and just return the result, so in that case
	// we do need an average.
	if len(selectors) > 1 && aggregationType == grpc_monitoring_go.AggregationType_AVG {
		request.Aggregation = grpc_monitoring_go.AggregationType_SUM
	}

	// Results is a mapping from metric to values, where values is a mapping
	// from timestamp to value and count. This last mapping is needed for merging
	// results from multiple edge controllers. We will convert to one
	// QueryMetricsResult to return afterwards
	results := NewQueryResults()

	// Request for each Edge Controller and execute
	for _, selector := range(selectors) {
		proxyRequest := &grpc_monitoring_go.QueryMetricsRequest{
			Assets: selector,
			Metrics: request.GetMetrics(),
			TimeRange: request.GetTimeRange(),
			Aggregation: request.GetAggregation(),
		}
		ecId := selector.GetEdgeControllerId()
		log.Debug().Interface("request", proxyRequest).Msg("proxy request for QueryMetrics")
		ctx, cancel := ProxyContext() // Manual calling cancel to avoid big list of defers
		result, err := m.proxyClient.QueryMetrics(ctx, proxyRequest)
		cancel()

		// Optimization when we're querying only a single EC, in which
		// case we can also return any errors
		if len(selectors) == 1 {
			log.Debug().Str("ecid", ecId).Msg("querying single edge controller - skipping merging")
			return result, err
		}

		if err != nil {
			// We still want to query to working edge controllers
			log.Warn().Str("edge-controller-id", ecId).Err(err).Msg("failed calling QueryMetrics")
			continue
		}

		// Add to results
		results.AddResult(ecId, result)
	}

	return results.GetQueryMetricsResult(aggregationType)
}
