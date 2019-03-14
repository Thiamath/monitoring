/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query result implementation

package prometheus

import (
	"strconv"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"

	"github.com/prometheus/common/model"
)

type PrometheusResult struct {
	Type PrometheusResultType
	Values []*PrometheusResultValue
}

type PrometheusResultValue struct {
	Labels map[string]string
	Values []*PrometheusValue
}

type PrometheusValue struct {
	Timestamp time.Time
	Value string
}

type PrometheusResultType string
func (t PrometheusResultType) String() string {
	return string(t)
}

const (
	PrometheusResultScalar PrometheusResultType = "SCALAR"
	PrometheusResultVector PrometheusResultType = "VECTOR"
	PrometheusResultMatrix PrometheusResultType = "MATRIX"
	PrometheusResultString PrometheusResultType = "STRING"
)

func NewPrometheusResult(val model.Value) *PrometheusResult {
	var result *PrometheusResult = nil

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

func (r *PrometheusResult) ResultType() query.QueryProviderType {
	return ProviderType
}

func (r *PrometheusResult) GetScalarInt() (val int64, derr derrors.Error) {
	// We want to catch the panic if some of the arrays below are
	// out of bound
	defer func() {
		if r := recover(); r != nil {
			val = 0
			derr = derrors.NewInternalError("query result empty")
		}
	}()
	if r.Type != PrometheusResultScalar {
		return 0, derrors.NewInternalError("query does not return scalar")
	}

        fval, err := strconv.ParseFloat(r.Values[0].Values[0].Value, 64)
        if err != nil {
                return 0, derrors.NewInternalError("invalid query result", err)
        }

	// TODO: proper float to int

	return int64(fval), nil
}

func scalarResult(val model.Value) *PrometheusResult {
	v := val.(*model.Scalar)

	result := &PrometheusResult{
		Type: PrometheusResultScalar,
		Values: singleValueResult(v.Timestamp.Time(), v.Value.String()),
	}

	return result
}

func vectorResult(val model.Value) *PrometheusResult {
	v := val.(model.Vector)

	resVals := make([]*PrometheusResultValue, 0, v.Len())
	for _, sample := range(([]*model.Sample)(v)) {
		resVal := &PrometheusResultValue{
			Labels: metricToLabel(sample.Metric),
			Values: singleValueList(sample.Timestamp.Time(), sample.Value.String()),
		}
		resVals = append(resVals, resVal)
	}

	result := &PrometheusResult{
		Type: PrometheusResultVector,
		Values: resVals,
	}

	return result
}

func matrixResult(val model.Value) *PrometheusResult {
	v := val.(model.Matrix)

	resVals := make([]*PrometheusResultValue, 0, v.Len())
	for _, sampleStream := range(([]*model.SampleStream)(v)) {
		values := make([]*PrometheusValue, 0, len(sampleStream.Values))
		for _, sample := range(sampleStream.Values) {
			values = append(values, value(sample.Timestamp.Time(), sample.Value.String()))
		}
		resVal := &PrometheusResultValue{
			Labels: metricToLabel(sampleStream.Metric),
			Values: values,
		}
		resVals = append(resVals, resVal)
	}

	result := &PrometheusResult{
		Type: PrometheusResultMatrix,
		Values: resVals,
	}

	return result
}

func stringResult(val model.Value) *PrometheusResult {
	v := val.(*model.String)

	result := &PrometheusResult{
		Type: PrometheusResultString,
		Values: singleValueResult(v.Timestamp.Time(), v.Value),
	}

	return result
}

func value(ts time.Time, s string) *PrometheusValue {
	return &PrometheusValue{
		Timestamp: ts,
		Value: s,
	}
}

func singleValueList(ts time.Time, s string) []*PrometheusValue {
	return []*PrometheusValue{
		value(ts, s),
	}
}

func singleValueResult(ts time.Time, s string) []*PrometheusResultValue {
	return []*PrometheusResultValue{
		&PrometheusResultValue{
			Values: singleValueList(ts, s),
		},
	}
}

func metricToLabel(m model.Metric) map[string]string {
	label := make(map[string]string, len(m))
	for k,v := range(m) {
		label[string(k)] = string(v)
	}

	return label
}
