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

package server

import (
	"os"
	"testing"
	"time"

	"github.com/nalej/grpc-monitoring-go"
	"github.com/nalej/grpc-utils/pkg/test"

	"github.com/nalej/monitoring/internal/pkg/metrics-collector/translators"
	"github.com/nalej/monitoring/internal/pkg/utils"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/nalej/monitoring/pkg/provider/query/fake"
	"github.com/nalej/monitoring/pkg/provider/query/prometheus"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"

	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestHandlerPackage(t *testing.T) {
	gomega.RegisterFailHandler(ginkgo.Fail)
	ginkgo.RunSpecs(t, "internal/app/metrics-collector package suite")
}

var listener *bufconn.Listener
var grpcServer *grpc.Server

var client grpc_monitoring_go.MetricsCollectorClient

var manager Manager

var _ = ginkgo.BeforeSuite(func() {
	if utils.RunIntegrationTests() {
		beforeSuiteIntegrationTests()
	}

	beforeSuiteRetrieveManager()
})

var _ = ginkgo.AfterSuite(func() {
	if grpcServer != nil {
		grpcServer.GracefulStop()
	}

	if listener != nil {
		_ = listener.Close()
	}
})

func beforeSuiteIntegrationTests() {
	var prometheusAddress = os.Getenv("IT_PROMETHEUS_ADDRESS")

	if prometheusAddress == "" {
		ginkgo.Fail("missing environment variables")
	}

	prometheusConfig := &prometheus.Config{
		Enable: true,
		Url:    prometheusAddress,
	}

	conf := &Config{
		Port:      8423,
		InCluster: true, // We won't actually connect to K8s, but this passes validation

		QueryProviders: query.ProviderConfigs{
			prometheus.ProviderType: prometheusConfig,
		},
	}

	service, derr := NewService(conf)
	gomega.Expect(derr).To(gomega.Succeed())

	errChan := make(chan error, 1)
	listener = test.GetDefaultListener()
	grpcServer, derr = service.startRetrieve(listener, errChan)
	gomega.Expect(derr).To(gomega.Succeed())

	conn, err := test.GetConn(*listener)
	gomega.Expect(err).To(gomega.Succeed())

	client = grpc_monitoring_go.NewMetricsCollectorClient(conn)
}

func beforeSuiteRetrieveManager() {
	queries := map[query.Query]query.Result{
		query.Query{
			QueryString: "this is a valid fake query",
		}: fake.FakeResult("result 1"),
		query.Query{
			QueryString: "this is a valid fake query",
			Range: query.Range{
				Start: time.Date(2000, time.January, 1, 0, 0, 0, 0, time.UTC),
				End:   time.Date(2000, time.February, 1, 0, 0, 0, 0, time.UTC),
				Step:  time.Duration(10) * time.Second,
			},
		}: fake.FakeResult("result 2"),
		query.Query{
			QueryString: "this is an invalid fake query",
		}: fake.FakeResult(""),
	}

	templates := map[query.TemplateName]map[query.TemplateVars]int64{
		query.TemplateName_CPU + query.TemplateName_Total: {
			query.TemplateVars{AvgSeconds: 0}:   1,
			query.TemplateVars{AvgSeconds: 600}: 2,
		},
		query.TemplateName_CPU + query.TemplateName_Available: {
			query.TemplateVars{AvgSeconds: 0}:   3,
			query.TemplateVars{AvgSeconds: 600}: 4,
		},
		query.TemplateName_Memory + query.TemplateName_Total: {
			query.TemplateVars{AvgSeconds: 0}:   5,
			query.TemplateVars{AvgSeconds: 600}: 6,
		},
		query.TemplateName_Memory + query.TemplateName_Available: {
			query.TemplateVars{AvgSeconds: 0}:   7,
			query.TemplateVars{AvgSeconds: 600}: 8,
		},
		query.TemplateName_Storage + query.TemplateName_Total: {
			query.TemplateVars{AvgSeconds: 0}:   9,
			query.TemplateVars{AvgSeconds: 600}: 10,
		},
		query.TemplateName_Storage + query.TemplateName_Available: {
			query.TemplateVars{AvgSeconds: 0}:   11,
			query.TemplateVars{AvgSeconds: 600}: 12,
		},
		query.TemplateName_UsableStorage + query.TemplateName_Total: {
			query.TemplateVars{AvgSeconds: 0}:   13,
			query.TemplateVars{AvgSeconds: 600}: 14,
		},
		query.TemplateName_UsableStorage + query.TemplateName_Available: {
			query.TemplateVars{AvgSeconds: 0}:   15,
			query.TemplateVars{AvgSeconds: 600}: 16,
		},

		query.TemplateName_PlatformStatsCounter: {
			query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "created"}:    13,
			query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "deleted"}:    14,
			query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "errors"}:     15,
			query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "created"}:  16,
			query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "deleted"}:  17,
			query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "errors"}:   18,
			query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "created"}:     19,
			query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "deleted"}:     20,
			query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "errors"}:      21,
			query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "created"}:   22,
			query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "deleted"}:   23,
			query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "errors"}:    24,
			query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "created"}:   25,
			query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "deleted"}:   26,
			query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "errors"}:    27,
			query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "created"}: 28,
			query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "deleted"}: 29,
			query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "errors"}:  30,
			query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "created"}:   31,
			query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "deleted"}:   32,
			query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "errors"}:    33,
			query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "created"}: 34,
			query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "deleted"}: 35,
			query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "errors"}:  36,
		},

		query.TemplateName_PlatformStatsGauge: {
			query.TemplateVars{AvgSeconds: 0, MetricName: "services", StatName: "running"}:    37,
			query.TemplateVars{AvgSeconds: 600, MetricName: "services", StatName: "running"}:  38,
			query.TemplateVars{AvgSeconds: 0, MetricName: "volumes", StatName: "running"}:     39,
			query.TemplateVars{AvgSeconds: 600, MetricName: "volumes", StatName: "running"}:   40,
			query.TemplateVars{AvgSeconds: 0, MetricName: "fragments", StatName: "running"}:   41,
			query.TemplateVars{AvgSeconds: 600, MetricName: "fragments", StatName: "running"}: 42,
			query.TemplateVars{AvgSeconds: 0, MetricName: "endpoints", StatName: "running"}:   43,
			query.TemplateVars{AvgSeconds: 600, MetricName: "endpoints", StatName: "running"}: 44,
		},
	}

	provider, derr := fake.NewProvider(queries, templates)
	gomega.Expect(derr).To(gomega.Succeed())

	providers := query.Providers{
		provider.ProviderType(): provider,
	}

	manager, derr = NewManager(providers)
	gomega.Expect(derr).To(gomega.Succeed())

	/* Insert fake provider */
	grpc_monitoring_go.QueryType_name[-1] = provider.ProviderType().String()
	grpc_monitoring_go.QueryType_value[provider.ProviderType().String()] = -1

	/* We use the fake translator */
	translators.Register(provider.ProviderType(), translators.FakeTranslator)
}
