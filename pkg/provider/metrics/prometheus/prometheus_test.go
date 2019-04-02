/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Prometheus metrics collector test

package prometheus

import (
	"github.com/nalej/derrors"

	"github.com/nalej/infrastructure-monitor/pkg/metrics"

	"github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
)

var _ = ginkgo.Describe("prometheus", func() {

	var provider *MetricsProvider

	ginkgo.BeforeSuite(func() {
		var derr derrors.Error
		provider, derr = NewMetricsProvider()
		gomega.Expect(derr).To(gomega.Succeed())
	})

	ginkgo.It("should count metrics correctly", func() {
		provider.Existing(metrics.MetricVolumes)
		gomega.Expect(provider.GetMetrics()).To(gomega.Equal(metrics.Metrics{
			metrics.MetricVolumes: &metrics.Metric{
				Created: 0,
				Deleted: 0,
				Errors: 0,
				Running: 1,
			},
		}))

		provider.Create(metrics.MetricVolumes)
		gomega.Expect(provider.GetMetrics()).To(gomega.Equal(metrics.Metrics{
			metrics.MetricVolumes: &metrics.Metric{
				Created: 1,
				Deleted: 0,
				Errors: 0,
				Running: 2,
			},
		}))

		provider.Delete(metrics.MetricVolumes)
		gomega.Expect(provider.GetMetrics()).To(gomega.Equal(metrics.Metrics{
			metrics.MetricVolumes: &metrics.Metric{
				Created: 1,
				Deleted: 1,
				Errors: 0,
				Running: 1,
			},
		}))

		provider.Error(metrics.MetricVolumes)
		gomega.Expect(provider.GetMetrics()).To(gomega.Equal(metrics.Metrics{
			metrics.MetricVolumes: &metrics.Metric{
				Created: 1,
				Deleted: 1,
				Errors: 1,
				Running: 1,
			},
		}))
	})
})

/*
count a few things, get metrics
*/
