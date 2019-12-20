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
	"context"
	"github.com/nalej/derrors"
	"github.com/nalej/grpc-monitoring-go"
	"html/template"
	"strings"
	"time"
)

const monitoringTimeout = time.Minute

const prometheusExpositionTemplate = `
# {{.Stats.ServiceInstanceName}}
nalej_servinst_cpu_core{appinstid="{{.Stats.AppInstanceId}}",appinstname="{{.Stats.AppInstanceName}}",servgroupinstid="{{.Stats.ServiceGroupInstanceId}}",servgroupinstname="{{.Stats.ServiceGroupInstanceName}}",servinstid="{{.Stats.ServiceInstanceId}}",servinstname="{{.Stats.ServiceInstanceName}}"} {{printf "%f" .Stats.CpuMillicore}} {{.Timestamp}}
nalej_servinst_memory_byte{appinstid="{{.Stats.AppInstanceId}}",appinstname="{{.Stats.AppInstanceName}}",servgroupinstid="{{.Stats.ServiceGroupInstanceId}}",servgroupinstname="{{.Stats.ServiceGroupInstanceName}}",servinstid="{{.Stats.ServiceInstanceId}}",servinstname="{{.Stats.ServiceInstanceName}}"} {{printf "%f" .Stats.MemoryByte}} {{.Timestamp}}
nalej_servinst_storage_byte{appinstid="{{.Stats.AppInstanceId}}",appinstname="{{.Stats.AppInstanceName}}",servgroupinstid="{{.Stats.ServiceGroupInstanceId}}",servgroupinstname="{{.Stats.ServiceGroupInstanceName}}",servinstid="{{.Stats.ServiceInstanceId}}",servinstname="{{.Stats.ServiceInstanceName}}"} {{printf "%f" .Stats.StorageByte}} {{.Timestamp}}
`

type Manager struct {
	monitoringClient *grpc_monitoring_go.MonitoringManagerClient
}

func NewManager(monitoringClient *grpc_monitoring_go.MonitoringManagerClient) (*Manager, derrors.Error) {
	return &Manager{monitoringClient: monitoringClient}, nil
}

func (m *Manager) GetMonitoringClient() grpc_monitoring_go.MonitoringManagerClient {
	return *m.monitoringClient
}

func (m *Manager) Metrics(organizationID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), monitoringTimeout)
	defer cancel()
	stats, err := m.GetMonitoringClient().GetOrganizationApplicationStats(ctx, &grpc_monitoring_go.OrganizationApplicationStatsRequest{OrganizationId: organizationID})
	if err != nil {
		return nil, err
	}
	tmpl := template.Must(template.New("prometheusExpositionTemplate").Parse(prometheusExpositionTemplate))
	response := make([]string, len(stats.ServiceInstanceStats))
	for i, serviceStats := range stats.ServiceInstanceStats {
		var buffer strings.Builder
		err := tmpl.Execute(&buffer, PrometheusExpositionObject{
			Stats:     *serviceStats,
			Timestamp: stats.Timestamp,
		})
		if err != nil {
			return nil, err
		}
		response[i] = buffer.String()
	}

	return []byte(strings.Join(response, "\n")), nil
}

type PrometheusExpositionObject struct {
	Stats     grpc_monitoring_go.OrganizationApplicationStats
	Timestamp int64
}
