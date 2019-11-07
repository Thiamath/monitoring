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

package metrics_collector

import (
	"strings"

	grpc "github.com/nalej/grpc-monitoring-go"
)

func GRPCStatsFieldToMetric(g grpc.PlatformStatsField) string {
	return strings.ToLower(g.String())
}

func AllGRPCStatsFields() []grpc.PlatformStatsField {
	fields := make([]grpc.PlatformStatsField, 0, len(grpc.PlatformStatsField_name))
	for i, _ := range(grpc.PlatformStatsField_name) {
		fields = append(fields, grpc.PlatformStatsField(i))
	}

	return fields
}
