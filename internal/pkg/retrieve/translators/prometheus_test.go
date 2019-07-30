/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Tests from Prometheus query result translation

package translators

import (
	"time"

	grpc "github.com/nalej/grpc-monitoring-go"
	"github.com/golang/protobuf/ptypes/timestamp"

	. "github.com/nalej/monitoring/pkg/provider/query/prometheus"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("prometheus", func() {

	ginkgo.BeforeSuite(func() {

	})

	ginkgo.Context("PrometheusTranslator", func() {
		ginkgo.It("should translate a query result", func() {
			qres := PrometheusResult{
				Type: PrometheusResultMatrix,
				Values: []*PrometheusResultValue{
					&PrometheusResultValue{
						Labels: map[string]string{
							"job": "example",
							"instance": "node1",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: time.Unix(1435781430, 0).UTC(),
								Value: "1",
							},
							&PrometheusValue{
								Timestamp: time.Unix(1435781445, 0).UTC(),
								Value: "2",
							},
						},
					},
					&PrometheusResultValue{
						Labels: map[string]string{
							"job": "example",
							"instance": "node2",
						},
						Values: []*PrometheusValue{
							&PrometheusValue{
								Timestamp: time.Unix(1435781430, 0).UTC(),
								Value: "3",
							},
							&PrometheusValue{
								Timestamp: time.Unix(1435781445, 0).UTC(),
								Value: "4",
							},
						},
					},
				},
			}

			pres := grpc.QueryResponse{
				Type: grpc.QueryType_PROMETHEUS,
				Result: &grpc.QueryResponse_PrometheusResult{
					&grpc.QueryResponse_PrometheusResponse{
						ResultType: grpc.QueryResponse_PrometheusResponse_MATRIX,
						Result: []*grpc.QueryResponse_PrometheusResponse_ResultValue{
							&grpc.QueryResponse_PrometheusResponse_ResultValue{
								Metric: map[string]string{
									"job": "example",
									"instance": "node1",
								},
								Value: []*grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
									&grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781430},
										Value: "1",
									},
									&grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781445},
										Value: "2",
									},
								},
							},
							&grpc.QueryResponse_PrometheusResponse_ResultValue{
								Metric: map[string]string{
									"job": "example",
									"instance": "node2",
								},
								Value: []*grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
									&grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781430},
										Value: "3",
									},
									&grpc.QueryResponse_PrometheusResponse_ResultValue_Value{
										Timestamp: &timestamp.Timestamp{Seconds: 1435781445},
										Value: "4",
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
