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

package asset

import (
	"github.com/nalej/derrors"

	"github.com/nalej/grpc-inventory-go"
	"github.com/nalej/grpc-monitoring-go"
)

func ValidAssetSelector(selector *grpc_inventory_go.AssetSelector) derrors.Error {
	if selector == nil {
		return derrors.NewInvalidArgumentError("empty asset selector")
	}
	if selector.GetOrganizationId() == "" {
		return derrors.NewInvalidArgumentError("organization_id cannot be empty")
	}
	return nil
}

func ValidTimeRange(timeRange *grpc_monitoring_go.QueryMetricsRequest_TimeRange) derrors.Error {
	if !(timeRange.GetTimestamp() == 0) {
		if timeRange.GetTimeStart() != 0 || timeRange.GetTimeEnd() != 0 || timeRange.GetResolution() != 0 {
			return derrors.NewInvalidArgumentError("timestamp is set; start, end and resolution should be 0").
				WithParams(timeRange.GetTimestamp(), timeRange.GetTimeStart(),
					timeRange.GetTimeEnd(), timeRange.GetResolution())
		}
	} else {
		if timeRange.GetTimeStart() == 0 && timeRange.GetTimeEnd() == 0 {
			return derrors.NewInvalidArgumentError("timestamp is not set; either start, end or both should be set").
				WithParams(timeRange.GetTimestamp(), timeRange.GetTimeStart(),
					timeRange.GetTimeEnd(), timeRange.GetResolution())
		}
	}

	return nil
}

func ValidQueryMetricsRequest(request *grpc_monitoring_go.QueryMetricsRequest) derrors.Error {
	// We check the asset selector so we know we have an organization ID.
	derr := ValidAssetSelector(request.GetAssets())
	if derr != nil {
		return derr
	}

	// Check the time range to either be a point in time or a range
	derr = ValidTimeRange(request.GetTimeRange())
	if derr != nil {
		return derr
	}

	if len(request.GetAssets().GetAssetIds()) != 1 && request.GetAggregation() == grpc_monitoring_go.AggregationType_NONE {
		return derrors.NewInvalidArgumentError("metrics for more than one asset requested without aggregation method")
	}

	return nil
}
