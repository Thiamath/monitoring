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

// Tests from Prometheus query result translation

package translators

import (
	"time"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/nalej/grpc-monitoring-go"

	"github.com/nalej/monitoring/pkg/provider/query/prometheus"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("prometheus", func() {

	ginkgo.BeforeSuite(func() {

	})

	ginkgo.Context("PrometheusTranslator", func() {
		ginkgo.It("should translate a query result", func() {
			qres := prometheus.Result{
				Type: prometheus.ResultMatrix,
				Values: []*prometheus.ResultValue{
					{
						Labels: map[string]string{
							"job":      "example",
							"instance": "node1",
						},
						Values: []*prometheus.Value{
							{
								Timestamp: time.Unix(1435781430, 0).UTC(),
								Value:     "1",
							},
							{
								Timestamp: time.Unix(1435781445, 0).UTC(),
								Value:     "2",
							},
						},
					},
					{
						Labels: map[string]string{
							"job":      "example",
							"instance": "node2",
						},
						Values: []*prometheus.Value{
							{
								Timestamp: time.Unix(1435781430, 0).UTC(),
								Value:     "3",
							},
							{
								Timestamp: time.Unix(1435781445, 0).UTC(),
								Value:     "4",
							},
						},
					},
				},
			}

			pres := grpc_monitoring_go.QueryResponse{
				Type: grpc_monitoring_go.QueryType_PROMETHEUS,
				Result: &grpc_monitoring_go.QueryResponse_PrometheusResult{
					PrometheusResult: &grpc_monitoring_go.QueryResponse_PrometheusResponse{
						ResultType: grpc_monitoring_go.QueryResponse_PrometheusResponse_MATRIX,
						Result: []*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue{
							{
								Metric: map[string]string{
									"job":      "example",
									"instance": "node1",
								},
								Value: []*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue_Value{
									{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781430},
										Value:     "1",
									},
									{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781445},
										Value:     "2",
									},
								},
							},
							{
								Metric: map[string]string{
									"job":      "example",
									"instance": "node2",
								},
								Value: []*grpc_monitoring_go.QueryResponse_PrometheusResponse_ResultValue_Value{
									{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781430},
										Value:     "3",
									},
									{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781445},
										Value:     "4",
									},
								},
							},
						},
					},
				},
			}

			gomega.Expect(PrometheusTranslator(&qres)).To(gomega.Equal(&pres))
		})

		ginkgo.It("should handle error result", func() {
			res, derr := PrometheusTranslator(nil)
			gomega.Expect(res).To(gomega.BeNil())
			gomega.Expect(derr).To(gomega.HaveOccurred())
		})
	})
})
