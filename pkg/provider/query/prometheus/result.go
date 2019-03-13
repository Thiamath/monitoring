/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus query result implementation

package prometheus

import (
	"time"

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
	case model.ValString:
	}

	return result
}

func (r *PrometheusResult) ResultType() query.QueryProviderType {
	return ProviderType
}

func scalarResult(val model.Value) *PrometheusResult {
	v := val.(*model.Scalar)

	result := &PrometheusResult{
		Type: PrometheusResultScalar,
		Values: []*PrometheusResultValue{
			&PrometheusResultValue{
				Values: []*PrometheusValue{
					&PrometheusValue{
						Timestamp: v.Timestamp.Time(),
						Value: v.Value.String(),
					},
				},
			},
		},
	}

	return result
}

func vectorResult(val model.Value) *PrometheusResult {
	v := val.(model.Vector)

	resVals := make([]*PrometheusResultValue, 0, v.Len())
	for _, sample := range(([]*model.Sample)(v)) {
		labels := map[string]string{}
		for k,v := range(sample.Metric) {
			labels[string(k)] = string(v)
		}
		resVal := &PrometheusResultValue{
			Labels: labels,
			Values: []*PrometheusValue{
				&PrometheusValue{
					Timestamp: sample.Timestamp.Time(),
					Value: sample.Value.String(),
				},
			},
		}
		resVals = append(resVals, resVal)
	}

	result := &PrometheusResult{
		Type: PrometheusResultVector,
		Values: resVals,
	}

	return result
}
