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

// Manager tests

package server

import (
	"context"

	"github.com/nalej/monitoring/internal/pkg/metrics-collector/translators"

	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/nalej/grpc-monitoring-go"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

const (
	OrganizationId = "77b5425b-4276-45b8-85f4-c01f74bbc376"
	ClusterId      = "e98efd7d-166e-4419-ae71-4c81cff9442c"
)

var _ = ginkgo.Describe("retrieve_manager", func() {

	ginkgo.Context("GetClusterSummary", func() {
		ginkgo.It("should return cluster summary without range", func() {
			request := &grpc_monitoring_go.ClusterSummaryRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
			}

			result := &grpc_monitoring_go.ClusterSummary{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				CpuMillicores: &grpc_monitoring_go.ClusterStat{
					Total:     1,
					Available: 3,
				},
				MemoryBytes: &grpc_monitoring_go.ClusterStat{
					Total:     5,
					Available: 7,
				},
				StorageBytes: &grpc_monitoring_go.ClusterStat{
					Total:     9,
					Available: 11,
				},
				UsableStorageBytes: &grpc_monitoring_go.ClusterStat{
					Total:     13,
					Available: 15,
				},
			}
			gomega.Expect(manager.GetClusterSummary(context.Background(), request)).To(gomega.Equal(result))
		})

		ginkgo.It("should return cluster summary with range", func() {
			request := &grpc_monitoring_go.ClusterSummaryRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				RangeMinutes:   10,
			}

			result := &grpc_monitoring_go.ClusterSummary{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				CpuMillicores: &grpc_monitoring_go.ClusterStat{
					Total:     2,
					Available: 4,
				},
				MemoryBytes: &grpc_monitoring_go.ClusterStat{
					Total:     6,
					Available: 8,
				},
				StorageBytes: &grpc_monitoring_go.ClusterStat{
					Total:     10,
					Available: 12,
				},
				UsableStorageBytes: &grpc_monitoring_go.ClusterStat{
					Total:     14,
					Available: 16,
				},
			}
			gomega.Expect(manager.GetClusterSummary(context.Background(), request)).To(gomega.Equal(result))
		})
	})

	ginkgo.Context("GetClusterStats", func() {
		ginkgo.It("should return cluster stats for single metric", func() {
			request := &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Fields:         []grpc_monitoring_go.PlatformStatsField{grpc_monitoring_go.PlatformStatsField_VOLUMES},
			}

			result := &grpc_monitoring_go.ClusterStats{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Stats: map[int32]*grpc_monitoring_go.PlatformStat{
					int32(grpc_monitoring_go.PlatformStatsField_VOLUMES): {
						Created: 19,
						Deleted: 20,
						Running: 39,
						Errors:  21,
					},
				},
			}
			gomega.Expect(manager.GetClusterStats(context.Background(), request)).To(gomega.Equal(result))
		})

		ginkgo.It("should return cluster stats for all metrics without range", func() {
			request := &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
			}

			result := &grpc_monitoring_go.ClusterStats{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Stats: map[int32]*grpc_monitoring_go.PlatformStat{
					int32(grpc_monitoring_go.PlatformStatsField_SERVICES): {
						Created: 13,
						Deleted: 14,
						Running: 37,
						Errors:  15,
					},
					int32(grpc_monitoring_go.PlatformStatsField_VOLUMES): {
						Created: 19,
						Deleted: 20,
						Running: 39,
						Errors:  21,
					},
					int32(grpc_monitoring_go.PlatformStatsField_FRAGMENTS): {
						Created: 25,
						Deleted: 26,
						Running: 41,
						Errors:  27,
					},
					int32(grpc_monitoring_go.PlatformStatsField_ENDPOINTS): {
						Created: 31,
						Deleted: 32,
						Running: 43,
						Errors:  33,
					},
				},
			}
			gomega.Expect(manager.GetClusterStats(context.Background(), request)).To(gomega.Equal(result))
		})

		ginkgo.It("should return cluster stats for all metrics with range", func() {
			request := &grpc_monitoring_go.ClusterStatsRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				RangeMinutes:   10,
			}

			result := &grpc_monitoring_go.ClusterStats{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Stats: map[int32]*grpc_monitoring_go.PlatformStat{
					int32(grpc_monitoring_go.PlatformStatsField_SERVICES): {
						Created: 16,
						Deleted: 17,
						Running: 38,
						Errors:  18,
					},
					int32(grpc_monitoring_go.PlatformStatsField_VOLUMES): {
						Created: 22,
						Deleted: 23,
						Running: 40,
						Errors:  24,
					},
					int32(grpc_monitoring_go.PlatformStatsField_FRAGMENTS): {
						Created: 28,
						Deleted: 29,
						Running: 42,
						Errors:  30,
					},
					int32(grpc_monitoring_go.PlatformStatsField_ENDPOINTS): {
						Created: 34,
						Deleted: 35,
						Running: 44,
						Errors:  36,
					},
				},
			}
			gomega.Expect(manager.GetClusterStats(context.Background(), request)).To(gomega.Equal(result))
		})
	})

	ginkgo.Context("Query", func() {
		ginkgo.It("should accept a valid query without range", func() {
			request := &grpc_monitoring_go.QueryRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Type:           grpc_monitoring_go.QueryType(-1), // FAKE
				Query:          "this is a valid fake query",
			}

			response := &grpc_monitoring_go.QueryResponse{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Type:           grpc_monitoring_go.QueryType(-1), // FAKE
				Result:         &translators.QueryResponseFakeResult{Result: "result 1"},
			}

			gomega.Expect(manager.Query(context.Background(), request)).To(gomega.Equal(response))
		})

		ginkgo.It("should accept a valid query with range", func() {
			request := &grpc_monitoring_go.QueryRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Type:           grpc_monitoring_go.QueryType(-1), // FAKE
				Range: &grpc_monitoring_go.QueryRequest_QueryRange{
					Start: &timestamp.Timestamp{Seconds: 946684800},
					End:   &timestamp.Timestamp{Seconds: 949363200},
					Step:  10.0,
				},
				Query: "this is a valid fake query",
			}

			response := &grpc_monitoring_go.QueryResponse{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Type:           grpc_monitoring_go.QueryType(-1), // FAKE
				Result:         &translators.QueryResponseFakeResult{Result: "result 2"},
			}

			gomega.Expect(manager.Query(context.Background(), request)).To(gomega.Equal(response))
		})

		ginkgo.It("should handle an invalid query", func() {
			request := &grpc_monitoring_go.QueryRequest{
				OrganizationId: OrganizationId,
				ClusterId:      ClusterId,
				Type:           grpc_monitoring_go.QueryType(-1), // FAKE
				Query:          "this is an invalid fake query",
			}

			result, err := manager.Query(context.Background(), request)
			gomega.Expect(result).To(gomega.BeNil())
			gomega.Expect(err).To(gomega.HaveOccurred())
		})
	})
})
