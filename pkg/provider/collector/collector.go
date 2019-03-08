/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Collect and store metrics in memory

package collector

import (
	"github.com/nalej/derrors"
)

type CountMetric struct {
	m MetricType
	c int64
}

type Collector struct {
	metrics Metrics
	countChan chan CountMetric
}

func NewCollector() (*Collector, derrors.Error) {
	collector := &Collector{
		metrics: Metrics{},
		countChan: make(chan CountMetric, 10),
	}

	// Start serialized metrics collector
	go collector.collect()

	return collector, nil
}

func (c *Collector) collect() {
	for {
		count := <-c.countChan
		metric, found := c.metrics[count.m]
		if !found {
			metric = &Metric{}
		}

		metric.CurrentRunning = metric.CurrentRunning + count.c
		if count.c > 0 {
			metric.Created = metric.Created + count.c
		} else {
			metric.Deleted = metric.Deleted - count.c
		}

		c.metrics[count.m] = metric
	}
}

func (c *Collector) Create(t MetricType) {
	count := CountMetric{
		m: t,
		c: 1,
	}
	c.countChan <- count
}

func (c *Collector) Delete(t MetricType) {
	count := CountMetric{
		m: t,
		c: -1,
	}

	c.countChan <- count
}

func (c *Collector) GetMetrics(metrics ...MetricType) (Metrics, derrors.Error) {
	// TODO: filter metrics
	return c.metrics, nil
}
