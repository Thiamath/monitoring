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

// Structs for platform metrics for collecting and querying

package query

// String references for the counters in a metric
type MetricCounter string

const (
	MetricCreated MetricCounter = "created"
	MetricDeleted MetricCounter = "deleted"
	MetricErrors  MetricCounter = "errors"
	MetricRunning MetricCounter = "running"
)

func (m MetricCounter) String() string {
	return string(m)
}

// Type of metric values
type ValueType string

const (
	// Monotonic increasing
	ValueCounter ValueType = "counter"
	// Variable
	ValueGauge ValueType = "gauge"
)

var CounterMap = map[MetricCounter]ValueType{
	MetricCreated: ValueCounter,
	MetricDeleted: ValueCounter,
	MetricErrors:  ValueCounter,
	MetricRunning: ValueGauge,
}
