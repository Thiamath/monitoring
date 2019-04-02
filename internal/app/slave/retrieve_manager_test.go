/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// RetrieveManager tests

package slave

import (
	"context"
	"time"

	"github.com/nalej/infrastructure-monitor/internal/pkg/retrieve/translators"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query"
	"github.com/nalej/infrastructure-monitor/pkg/provider/query/fake"

	grpc "github.com/nalej/grpc-infrastructure-monitor-go"
	"github.com/golang/protobuf/ptypes/timestamp"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	OrganizationId = "77b5425b-4276-45b8-85f4-c01f74bbc376"
	ClusterId = "e98efd7d-166e-4419-ae71-4c81cff9442c"
)

var _ = ginkgo.Describe("retrieve_manager", func() {

	var manager *RetrieveManager

	ginkgo.BeforeSuite(func() {
		queries := map[query.Query]query.QueryResult{
			query.Query{
				QueryString: "this is a valid fake query",
			}: fake.FakeResult("result 1"),
			query.Query{
				QueryString: "this is a valid fake query",
				Range: query.QueryRange{
					Start: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
					End: time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC),
					Step: time.Duration(10) * time.Second,
				},
			}: fake.FakeResult("result 2"),
			query.Query{
				QueryString: "this is an invalid fake query",
			}: fake.FakeResult(""),
		}

		templates := map[query.TemplateName]map[query.TemplateVars]int64{
			query.TemplateName_CPU + query.TemplateName_Total: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0}: 1,
				query.TemplateVars{AvgSeconds: 600}: 2,
			},
			query.TemplateName_CPU + query.TemplateName_Available: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0}: 3,
				query.TemplateVars{AvgSeconds: 600}: 4,
			},
			query.TemplateName_Memory + query.TemplateName_Total: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0}: 5,
				query.TemplateVars{AvgSeconds: 600}: 6,
			},
			query.TemplateName_Memory + query.TemplateName_Available: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0}: 7,
				query.TemplateVars{AvgSeconds: 600}: 8,
			},
			query.TemplateName_Storage + query.TemplateName_Total: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0}: 9,
				query.TemplateVars{AvgSeconds: 600}: 10,
			},
			query.TemplateName_Storage + query.TemplateName_Available: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0}: 11,
				query.TemplateVars{AvgSeconds: 600}: 12,
			},

			query.TemplateName_PlatformStatsCounter: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "created"}: 13,
				query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "deleted"}: 14,
				query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "errors"}: 15,
				query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "created"}: 16,
				query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "deleted"}: 17,
				query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "errors"}: 18,
				query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "created"}: 19,
				query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "deleted"}: 20,
				query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "errors"}: 21,
				query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "created"}: 22,
				query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "deleted"}: 23,
				query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "errors"}: 24,
				query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "created"}: 25,
				query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "deleted"}: 26,
				query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "errors"}: 27,
				query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "created"}: 28,
				query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "deleted"}: 29,
				query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "errors"}: 30,
				query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "created"}: 31,
				query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "deleted"}: 32,
				query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "errors"}: 33,
				query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "created"}: 34,
				query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "deleted"}: 35,
				query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "errors"}: 36,
			},

			query.TemplateName_PlatformStatsGauge: map[query.TemplateVars]int64{
				query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "running"}: 37,
				query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "running"}: 38,
				query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "running"}: 39,
				query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "running"}: 40,
				query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "running"}: 41,
				query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "running"}: 42,
				query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "running"}: 43,
				query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "running"}: 44,
			},
		}

		provider, derr := fake.NewFakeProvider(queries, templates)
		gomega.Expect(derr).To(gomega.Succeed())

		providers := query.QueryProviders{
			provider.ProviderType(): provider,
		}

		manager, derr = NewRetrieveManager(providers)
		gomega.Expect(derr).To(gomega.Succeed())

		/* Insert fake provider */
		grpc.QueryType_name[-1] = provider.ProviderType().String()
		grpc.QueryType_value[provider.ProviderType().String()] = -1

		/* We use the Prometheus translator */
		translators.Register(provider.ProviderType(), translators.FakeTranslator)
	})

	ginkgo.Context("GetClusterSummary", func() {
		ginkgo.It("should return cluster summary without range", func() {
			request := &grpc.ClusterSummaryRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
			}

			result := &grpc.ClusterSummary{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				CpuMillicores: &grpc.ClusterStat{
					Total: 1,
					Available: 3,
				},
				MemoryBytes: &grpc.ClusterStat{
					Total: 5,
					Available: 7,
				},
				StorageBytes: &grpc.ClusterStat{
					Total: 9,
					Available: 11,
				},
			}
			gomega.Expect(manager.GetClusterSummary(context.Background(), request)).To(gomega.Equal(result))
		})

		ginkgo.It("should return cluster summary with range", func() {
			request := &grpc.ClusterSummaryRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				RangeMinutes: 10,
			}

			result := &grpc.ClusterSummary{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				CpuMillicores: &grpc.ClusterStat{
					Total: 2,
					Available: 4,
				},
				MemoryBytes: &grpc.ClusterStat{
					Total: 6,
					Available: 8,
				},
				StorageBytes: &grpc.ClusterStat{
					Total: 10,
					Available: 12,
				},
			}
			gomega.Expect(manager.GetClusterSummary(context.Background(), request)).To(gomega.Equal(result))
		})
	})

	ginkgo.Context("GetClusterStats", func() {
		ginkgo.It("should return cluster stats for single metric", func() {
			request := &grpc.ClusterStatsRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Fields: []grpc.PlatformStatsField{grpc.PlatformStatsField_VOLUMES},
			}

			result := &grpc.ClusterStats{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Stats: map[int32]*grpc.PlatformStat{
					int32(grpc.PlatformStatsField_VOLUMES): &grpc.PlatformStat{
						Created: 19,
						Deleted: 20,
						Running: 39,
						Errors: 21,
					},
				},
			}
			gomega.Expect(manager.GetClusterStats(context.Background(), request)).To(gomega.Equal(result))
		})

		ginkgo.It("should return cluster stats for all metrics without range", func() {
			request := &grpc.ClusterStatsRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
			}

			result := &grpc.ClusterStats{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Stats: map[int32]*grpc.PlatformStat{
					int32(grpc.PlatformStatsField_SERVICES): &grpc.PlatformStat{
						Created: 13,
						Deleted: 14,
						Running: 37,
						Errors: 15,
					},
					int32(grpc.PlatformStatsField_VOLUMES): &grpc.PlatformStat{
						Created: 19,
						Deleted: 20,
						Running: 39,
						Errors: 21,
					},
					int32(grpc.PlatformStatsField_FRAGMENTS): &grpc.PlatformStat{
						Created: 25,
						Deleted: 26,
						Running: 41,
						Errors: 27,
					},
					int32(grpc.PlatformStatsField_ENDPOINTS): &grpc.PlatformStat{
						Created: 31,
						Deleted: 32,
						Running: 43,
						Errors: 33,
					},
				},
			}
			gomega.Expect(manager.GetClusterStats(context.Background(), request)).To(gomega.Equal(result))
		})

		ginkgo.It("should return cluster stats for all metrics with range", func() {
			request := &grpc.ClusterStatsRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				RangeMinutes: 10,
			}

			result := &grpc.ClusterStats{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Stats: map[int32]*grpc.PlatformStat{
					int32(grpc.PlatformStatsField_SERVICES): &grpc.PlatformStat{
						Created: 16,
						Deleted: 17,
						Running: 38,
						Errors: 18,
					},
					int32(grpc.PlatformStatsField_VOLUMES): &grpc.PlatformStat{
						Created: 22,
						Deleted: 23,
						Running: 40,
						Errors: 24,
					},
					int32(grpc.PlatformStatsField_FRAGMENTS): &grpc.PlatformStat{
						Created: 28,
						Deleted: 29,
						Running: 42,
						Errors: 30,
					},
					int32(grpc.PlatformStatsField_ENDPOINTS): &grpc.PlatformStat{
						Created: 34,
						Deleted: 35,
						Running: 44,
						Errors: 36,
					},
				},
			}
			gomega.Expect(manager.GetClusterStats(context.Background(), request)).To(gomega.Equal(result))
		})
	})

	ginkgo.Context("Query", func() {
		ginkgo.It("should accept a valid query without range", func() {
			request := &grpc.QueryRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Type: grpc.QueryType(-1), // FAKE
				Query: "this is a valid fake query",
			}

			response := &grpc.QueryResponse{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Type: grpc.QueryType(-1), // FAKE
				Result: &translators.QueryResponse_FakeResult{Result: "result 1"},
			}

			gomega.Expect(manager.Query(context.Background(), request)).To(gomega.Equal(response))
		})

		ginkgo.It("should accept a valid query with range", func() {
			request := &grpc.QueryRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Type: grpc.QueryType(-1), // FAKE
				Range: &grpc.QueryRequest_QueryRange{
					Start: &timestamp.Timestamp{Seconds: 946684800},
					End: &timestamp.Timestamp{Seconds: 949363200},
					Step: 10.0,
				},
				Query: "this is a valid fake query",
			}

			response := &grpc.QueryResponse{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Type: grpc.QueryType(-1), // FAKE
				Result: &translators.QueryResponse_FakeResult{Result: "result 2"},
			}

			gomega.Expect(manager.Query(context.Background(), request)).To(gomega.Equal(response))
		})

		ginkgo.It("should handle an invalid query", func() {
			request := &grpc.QueryRequest{
				OrganizationId: OrganizationId,
				ClusterId: ClusterId,
				Type: grpc.QueryType(-1), // FAKE
				Query: "this is an invalid fake query",
			}

			result, err := manager.Query(context.Background(), request)
			gomega.Expect(result).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})

/*
Create stub provider
 have fixed replies for fixed requests
 check response on requests
 tests query translator
  stub translator?
*/
