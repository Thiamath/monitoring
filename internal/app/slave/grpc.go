/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Slave implementation for RetrieveManager

package slave

import (
	"strings"

	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	grpc "github.com/nalej/grpc-infrastructure-monitor-go"
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
