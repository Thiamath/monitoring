/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Simple implementation of collector interface

package metrics

import (
	"github.com/nalej/derrors"
)

type countMetric struct {
	m MetricType
	c int64
	e bool
}

// Simple Collector implementation that stores the platform metrics in a
// map - mostly used for simple testing as it might not be the most
// optimal or eveb correct implementation
type SimpleCollector struct {
	metrics Metrics
	countChan chan countMetric
}

func NewSimpleCollector() (*SimpleCollector, derrors.Error) {
	collector := &SimpleCollector{
		metrics: Metrics{},
		countChan: make(chan countMetric, 10),
	}

	// Start serialized metrics collector
	go collector.collect()

	return collector, nil
}

func (c *SimpleCollector) collect() {
	for {
		count := <-c.countChan
		metric, found := c.metrics[count.m]
		if !found {
			metric = &Metric{}
		}

		metric.Running = metric.Running + count.c
		if count.c > 0 {
			if !count.e {
				metric.Created = metric.Created + count.c
			}
		} else {
			metric.Deleted = metric.Deleted - count.c
		}

		c.metrics[count.m] = metric
	}
}

func (c *SimpleCollector) Create(t MetricType) {
	count := countMetric{
		m: t,
		c: 1,
	}
	c.countChan <- count
}

func (c *SimpleCollector) Existing(t MetricType) {
	count := countMetric{
		m: t,
		c: 1,
		e: true,
	}
	c.countChan <- count
}

func (c *SimpleCollector) Delete(t MetricType) {
	count := countMetric{
		m: t,
		c: -1,
	}

	c.countChan <- count
}

func (c *SimpleCollector) GetMetrics(types ...MetricType) (Metrics, derrors.Error) {
	// TODO: filter metrics
	return c.metrics, nil
}
