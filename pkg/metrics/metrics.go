/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Structs for platform metrics for collecting and querying

package metrics

// List of available metrics
type MetricType string
const (
	MetricServices MetricType = "services"
	MetricVolumes MetricType = "volumes"
	MetricFragments MetricType = "fragments"
	MetricEndpoints MetricType = "endpoints"
)

func (m MetricType) String() string {
	return string(m)
}

// String references for the counters in a metric
type MetricCounter string
const (
	MetricCreated MetricCounter = "created"
	MetricDeleted MetricCounter = "deleted"
	MetricErrors MetricCounter = "errors"
	MetricRunning MetricCounter = "running"
)

func (m MetricCounter) String() string {
	return string(m)
}

// Type of metric values
type ValueType string
const (
	// Monotonic increasing
	ValueCounter ValueType = "couter"
	// Variable
	ValueGauge ValueType = "gauge"
)

var CounterMap = map[MetricCounter]ValueType{
	MetricCreated: ValueCounter,
	MetricDeleted: ValueCounter,
	MetricErrors: ValueCounter,
	MetricRunning: ValueGauge,
}

// Individual metric
type Metric struct {
	Created, Deleted, Running, Errors int64
}

// Metrics collection
type Metrics map[MetricType]*Metric
