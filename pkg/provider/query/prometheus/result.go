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

// Prometheus query result implementation

package prometheus

import (
	"strconv"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/pkg/provider/query"

	"github.com/prometheus/common/model"
)

type Result struct {
	Type   ResultType
	Values []*ResultValue
}

type ResultValue struct {
	Labels map[string]string
	Values []*Value
}

type Value struct {
	Timestamp time.Time
	Value     string
}

type ResultType string

func (t ResultType) String() string {
	return string(t)
}

const (
	ResultScalar ResultType = "SCALAR"
	ResultVector ResultType = "VECTOR"
	ResultMatrix ResultType = "MATRIX"
	ResultString ResultType = "STRING"
)

func NewPrometheusResult(val model.Value) *Result {
	var result *Result = nil

	switch val.Type() {
	case model.ValScalar:
		result = scalarResult(val)
	case model.ValVector:
		result = vectorResult(val)
	case model.ValMatrix:
		result = matrixResult(val)
	case model.ValString:
		result = stringResult(val)
	}

	return result
}

func (r *Result) ResultType() query.ProviderType {
	return ProviderType
}

func (r *Result) GetScalarInt() (val int64, derr derrors.Error) {
	// We want to catch the panic if some of the arrays below are
	// out of bound
	defer func() {
		if r := recover(); r != nil {
			val = 0
			derr = derrors.NewInternalError("query result empty")
		}
	}()
	if r.Type != ResultScalar {
		return 0, derrors.NewInternalError("query does not return scalar")
	}

	fval, err := strconv.ParseFloat(r.Values[0].Values[0].Value, 64)
	if err != nil {
		return 0, derrors.NewInternalError("invalid query result", err)
	}

	ival, err := Ftoi(fval)
	if err != nil {
		return 0, derrors.NewInternalError("error converting query result", err)
	}
	return ival, nil
}

func scalarResult(val model.Value) *Result {
	v := val.(*model.Scalar)

	result := &Result{
		Type:   ResultScalar,
		Values: singleValueResult(v.Timestamp.Time(), v.Value.String()),
	}

	return result
}

func vectorResult(val model.Value) *Result {
	v := val.(model.Vector)

	resVals := make([]*ResultValue, 0, v.Len())
	for _, sample := range ([]*model.Sample)(v) {
		resVal := &ResultValue{
			Labels: metricToLabel(sample.Metric),
			Values: singleValueList(sample.Timestamp.Time(), sample.Value.String()),
		}
		resVals = append(resVals, resVal)
	}

	result := &Result{
		Type:   ResultVector,
		Values: resVals,
	}

	return result
}

func matrixResult(val model.Value) *Result {
	v := val.(model.Matrix)

	resVals := make([]*ResultValue, 0, v.Len())
	for _, sampleStream := range ([]*model.SampleStream)(v) {
		values := make([]*Value, 0, len(sampleStream.Values))
		for _, sample := range sampleStream.Values {
			values = append(values, value(sample.Timestamp.Time(), sample.Value.String()))
		}
		resVal := &ResultValue{
			Labels: metricToLabel(sampleStream.Metric),
			Values: values,
		}
		resVals = append(resVals, resVal)
	}

	result := &Result{
		Type:   ResultMatrix,
		Values: resVals,
	}

	return result
}

func stringResult(val model.Value) *Result {
	v := val.(*model.String)

	result := &Result{
		Type:   ResultString,
		Values: singleValueResult(v.Timestamp.Time(), v.Value),
	}

	return result
}

func value(ts time.Time, s string) *Value {
	return &Value{
		Timestamp: ts.UTC(),
		Value:     s,
	}
}

func singleValueList(ts time.Time, s string) []*Value {
	return []*Value{
		value(ts, s),
	}
}

func singleValueResult(ts time.Time, s string) []*ResultValue {
	return []*ResultValue{
		{
			Values: singleValueList(ts, s),
		},
	}
}

func metricToLabel(m model.Metric) map[string]string {
	label := make(map[string]string, len(m))
	for k, v := range m {
		label[string(k)] = string(v)
	}

	return label
}
