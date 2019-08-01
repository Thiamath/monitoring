/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Collector integration test

// Currently only test queries - not metrics (needs k8s)

package metrics_collector

import (
	"context"

	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/monitoring/internal/pkg/utils"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"github.com/rs/zerolog/log"
)

// NOTE: We don't check exact results as we current do not populate
// Prometheus. We are just validating that the queries are accepted
// and are syntactically valid.

var _ = ginkgo.Describe("integration tests", func() {
	if !utils.RunIntegrationTests() {
		log.Warn().Msg("Integration tests are skipped")
		return
	}

	ginkgo.Context("GetClusterSummary", func(){
		ginkgo.It("should return error on invalid request", func(){
			req := &grpc_monitoring_go.ClusterSummaryRequest{
				// Empty request
			}
			_, err := client.GetClusterSummary(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())

			req = &grpc_monitoring_go.ClusterSummaryRequest{
				OrganizationId: "org-id-1",
			}
			_, err = client.GetClusterSummary(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

		ginkgo.It("should succeed on request without range", func(){
			req := &grpc_monitoring_go.ClusterSummaryRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
			}
			_, err := client.GetClusterSummary(context.Background(), req)
			gomega.Expect(err).To(gomega.Succeed())
		})

		ginkgo.It("should succeed on request with range", func(){
			req := &grpc_monitoring_go.ClusterSummaryRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
				RangeMinutes: 10,
			}
			_, err := client.GetClusterSummary(context.Background(), req)
			gomega.Expect(err).To(gomega.Succeed())
		})
	})

	ginkgo.Context("GetClusterStats", func(){
		ginkgo.It("should return error on invalid request", func(){
			req := &grpc_monitoring_go.ClusterStatsRequest{
				// Empty request
			}
			_, err := client.GetClusterStats(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())

			req = &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: "org-id-1",
			}
			_, err = client.GetClusterStats(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

		ginkgo.It("should succeed on request without range", func(){
			req := &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
			}
			_, err := client.GetClusterStats(context.Background(), req)
			gomega.Expect(err).To(gomega.Succeed())
		})

		ginkgo.It("should succeed on request with range", func(){
			req := &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
				RangeMinutes: 10,
			}
			_, err := client.GetClusterStats(context.Background(), req)
			gomega.Expect(err).To(gomega.Succeed())
		})

		ginkgo.It("should succeed on request with field selector", func(){
			req := &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
				Fields: []grpc_monitoring_go.PlatformStatsField{
					grpc_monitoring_go.PlatformStatsField_VOLUMES,
					grpc_monitoring_go.PlatformStatsField_SERVICES,
				},
			}
			_, err := client.GetClusterStats(context.Background(), req)
			gomega.Expect(err).To(gomega.Succeed())
		})
	})

	ginkgo.Context("Query", func(){
		ginkgo.It("should return error on invalid request", func(){
			req := &grpc_monitoring_go.QueryRequest{
				// Empty request
			}
			_, err := client.Query(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())

			req = &grpc_monitoring_go.QueryRequest{
				OrganizationId: "org-id-1",
			}
			_, err = client.Query(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())

			req = &grpc_monitoring_go.QueryRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
			}
			_, err = client.Query(context.Background(), req)
			gomega.Expect(err).To(gomega.HaveOccurred())
		})

		ginkgo.It("should succeed on valid query", func(){
			req := &grpc_monitoring_go.QueryRequest{
				OrganizationId: "org-id-1",
				ClusterId: "cluster-id-1",
				Type: grpc_monitoring_go.QueryType_PROMETHEUS,
				Query: `{job=~".+"}`,
			}
			_, err := client.Query(context.Background(), req)
			gomega.Expect(err).To(gomega.Succeed())
		})
	})
})
