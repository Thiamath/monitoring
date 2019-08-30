/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package availability_cli

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/nalej/derrors"
	"github.com/nalej/monitoring/pkg/provider/query"
	"github.com/nalej/monitoring/pkg/provider/query/prometheus"
	"github.com/rs/zerolog/log"
)

// Status creates a connection with Prometheus and retrieves and outputs
// platform status
type Status struct {
	Configuration *Config
	Provider query.QueryProvider
}

func NewStatus(conf *Config) (*Status, derrors.Error) {
	derr := conf.Validate()
	if derr != nil {
		return nil, derr
	}

	p, derr := conf.Prometheus.NewProvider()
	if derr != nil {
		return nil, derr
	}

	s := &Status{
		Configuration: conf,
		Provider: p,
	}

	return s, nil
}

const (
	QueryNodesReady = "node:kube_node_status_condition:selectready > 0"
	QueryNodesMemoryPressure = "node:kube_node_status_condition:selectmemorypressure > 0"
	QueryNodesDiskPressure = "node:kube_node_status_condition:selectdiskpressure > 0"

	QueryDegradedComponents = `sum by (cluster_id, organization_id, component) (label_join(((deployment:kube_deployment_status_replicas_available_per_spec:union_ratio < 1) or (daemonset:kube_daemonset_status_number_available_per_desired:union_ratio < 1) or (statefulset:kube_statefulset_status_replicas_ready:union_ratio < 1)),"component","","deployment","daemonset","statefulset"))`
	QueryComponents = `sum by (cluster_id, organization_id, component) (label_join(((deployment:kube_deployment_status_replicas_available_per_spec:union_ratio) or (daemonset:kube_daemonset_status_number_available_per_desired:union_ratio) or (statefulset:kube_statefulset_status_replicas_ready:union_ratio)),"component","","deployment","daemonset","statefulset"))`
)

// Get and print current platform status
func (s *Status) GetStatus() derrors.Error {
	// TODO: If we really want to do this right, we need to query the
	// dashboards in Grafana to get list of clusters. If we set some variables
	// in the dashboards, we can also get cluster name and address.
	// If we don't do this, we won't show clusters that went down completely.

	ctx := context.Background()

	// Get number of clusters and total capacity
	clusterCount, derr := s.Provider.ExecuteTemplate(ctx, query.TemplateName_Clusters + query.TemplateName_Total, nil)
	if derr != nil {
		return derr
	}

	fmt.Printf("Clusters: %d\n", clusterCount)

	// Get healthy clusters
	healthyCount, derr := s.Provider.ExecuteTemplate(ctx, query.TemplateName_Clusters + query.TemplateName_Healthy, nil)
	if derr != nil {
		return derr
	}

	fmt.Printf("Healthy Clusters: %d\n", healthyCount)

	// Get node degredation conditions, store in map
	clusters := ClusterMap{}
	derr = s.queryNodeStatus(ctx, clusters, QueryNodesMemoryPressure, ConditionMemoryPressure)
	if derr != nil {
		return derr
	}
	derr = s.queryNodeStatus(ctx, clusters, QueryNodesDiskPressure, ConditionDiskPressure)
	if derr != nil {
		return derr
	}

	// Get healthy clusters if we want verbose results
	if s.Configuration.Verbose {
		derr = s.queryNodeStatus(ctx, clusters, QueryNodesReady, ConditionReady)
		if derr != nil {
			return derr
		}
	}

	// Get component issues
	componentsQuery := QueryDegradedComponents
	if s.Configuration.Verbose {
		componentsQuery = QueryComponents
	}
	result, derr := s.execQuery(ctx, componentsQuery)
	if derr != nil {
		return derr
	}
	for _, value := range(result) {
		if len(value.Values) != 1 {
			return derrors.NewInvalidArgumentError("received unexpected result with more than one value")
		}
		clusters.AddComponent(value.Labels["cluster_id"], value.Labels["component"], value.Values[0].Value)
	}

	log.Info().Interface("clusters", clusters).Msg("")
	return nil
}

type NodeCondition string
const (
	ConditionReady NodeCondition = "Ready"
	ConditionMemoryPressure NodeCondition = "Memory pressure"
	ConditionDiskPressure NodeCondition = "Disk pressure"
)

type NodeMap map[string][]NodeCondition
func (n NodeMap) AddNode(nodeName string, condition NodeCondition) {
	conditions, found := n[nodeName]
	if !found {
		conditions = []NodeCondition{}
	}

	conditions = append(conditions, condition)
	n[nodeName] = conditions
}

type Component struct {
	Name string
	Health float32
}

type Cluster struct {
	Nodes NodeMap
	Components []*Component
}


type ClusterMap map[string]*Cluster
func (c ClusterMap) GetCluster(clusterId string) *Cluster {
	cluster, found := c[clusterId]
	if !found {
		cluster = &Cluster{
			Nodes: NodeMap{},
			Components: []*Component{},
		}
		c[clusterId] = cluster
	}

	return cluster
}

func (c ClusterMap) AddNode(clusterId string, nodeName string, condition NodeCondition) {
	c.GetCluster(clusterId).Nodes.AddNode(nodeName, condition)
}

func (c ClusterMap) AddComponent(clusterId string, componentName string, healthStr string) {
	health, err := strconv.ParseFloat(healthStr, 32)
	if err != nil {
		health = 0
	}

	cluster := c.GetCluster(clusterId)
	component := &Component{
		Name: componentName,
		Health: float32(health),
	}

	cluster.Components = append(cluster.Components, component)
}

func (s *Status) execQuery(ctx context.Context, queryString string) ([]*prometheus.PrometheusResultValue, derrors.Error) {
	q := &query.Query{
		QueryString: queryString,
		Range: query.QueryRange{Start: time.Now()},
	}

	result, derr := s.Provider.Query(ctx, q)
	if derr != nil {
		return nil, derr
	}
	pResult, ok := result.(*prometheus.PrometheusResult)
	if !ok {
		return nil, derrors.NewInvalidArgumentError("invalid result received from Prometheus")
	}

	if pResult.Type != prometheus.PrometheusResultVector {
		return nil, derrors.NewInvalidArgumentError("invalid result type received from Prometheus")
	}

	return pResult.Values, nil
}

func (s *Status) queryNodeStatus(ctx context.Context, clusters ClusterMap, queryString string, condition NodeCondition) derrors.Error {
	result, derr := s.execQuery(ctx, queryString)
	if derr != nil {
		return derr
	}
	for _, value := range(result) {
		if len(value.Values) != 1 {
			return derrors.NewInvalidArgumentError("received unexpected result with more than one value")
		}
		clusters.AddNode(value.Labels["cluster_id"], value.Labels["node"], condition)
	}

	return nil
}
