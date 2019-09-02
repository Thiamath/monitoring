/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

package availability_cli

import (
	"strconv"
)

type Status struct {
	ClusterCount int `json:"cluster_count"`
	HealthyClusterCount int `json:"healthy_cluster_count"`
	Clusters ClusterMap `json:"clusters,omitempty"`
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
	Name string `json:"name"`
	Health float32 `json:"health"`
}

type Cluster struct {
	Nodes NodeMap `json:"nodes"`
	Components []*Component `json:"components"`
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
