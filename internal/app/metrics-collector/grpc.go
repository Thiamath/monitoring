/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package metrics_collector

import (
	"strings"

	"github.com/nalej/monitoring/pkg/metrics"

	grpc "github.com/nalej/grpc-monitoring-go"
)

func GRPCStatsFieldToMetric(g grpc.PlatformStatsField) metrics.MetricType {
	return metrics.MetricType(strings.ToLower(g.String()))
}

func AllGRPCStatsFields() []grpc.PlatformStatsField {
	fields := make([]grpc.PlatformStatsField, 0, len(grpc.PlatformStatsField_name))
	for i, _ := range(grpc.PlatformStatsField_name) {
		fields = append(fields, grpc.PlatformStatsField(i))
	}

	return fields
}
