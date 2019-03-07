/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Structs for metrics and list of all metrics that a provider can provide

package collector

// List of available metrics
type MetricType string
const (
	MetricServices MetricType = "services"
	MetricVolumes MetricType = "volumes"
	MetricFragments MetricType = "fragments"
	MetricEndpoints MetricType = "endpoints"
)

// Individual metric
type Metric struct {
	Created, Deleted, CurrentRunning, CurrentError int64
}

// Metrics collection
type Metrics map[MetricType]Metric
