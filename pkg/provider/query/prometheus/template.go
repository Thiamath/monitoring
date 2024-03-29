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

// Query templates for Prometheus

package prometheus

import (
	"github.com/nalej/monitoring/pkg/provider/query"
)

// We check if we ask for an average or for the latest number with
// AvgSeconds. If it's lower than 120 (2 minutes), we query for the
// latest instant number, as an average over a short time frame
// does not make sense due to lack of data points.
var queryTemplates = query.TemplateStringMap{
	// node_cpu counts the number of seconds a cpu spends in a certain
	// state. irate gets the rate-of-change per second, based on the last
	// two samples in a vector - so change-of-seconds-per-second is
	// CPU usage. Multiply numbers by 1000 to return millicores.
	// Rate gets rate-of-change per second over complete vector.
	query.TemplateName_CPU + query.TemplateName_Available: `
{{- if (gt .AvgSeconds 120) -}}
scalar(sum (rate(node_cpu_seconds_total{mode='idle'}[{{ .AvgSeconds }}s])) * 1000)
{{- else -}}
scalar(sum (irate(node_cpu_seconds_total{mode='idle'}[2m])) * 1000)
{{- end -}}
`,

	query.TemplateName_CPU + query.TemplateName_Total: `
{{- if (gt .AvgSeconds 120) -}}
scalar(avg_over_time(count(node_cpu_seconds_total{mode='idle'})[{{ .AvgSeconds }}s:60s]) * 1000)
{{- else -}}
scalar(count(node_cpu_seconds_total{mode='idle'}) * 1000)
{{- end -}}
`,

	query.TemplateName_Memory + query.TemplateName_Available: `
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_memory_MemAvailable_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_memory_MemAvailable_bytes))
{{- end -}}
`,

	query.TemplateName_Memory + query.TemplateName_Total: `
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_memory_MemTotal_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_memory_MemTotal_bytes))
{{- end -}}
`,

	query.TemplateName_Storage + query.TemplateName_Available: `
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_filesystem_free_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_filesystem_free_bytes))
{{- end -}}
`,

	query.TemplateName_Storage + query.TemplateName_Total: `
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_filesystem_size_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_filesystem_size_bytes))
{{- end -}}
`,

	query.TemplateName_UsableStorage + query.TemplateName_Available: `
{{- if (gt .AvgSeconds 120) -}}
scalar(max(avg_over_time(node_filesystem_free[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(max(node_filesystem_free))
{{- end -}}
`,

	query.TemplateName_UsableStorage + query.TemplateName_Total: `
{{- if (gt .AvgSeconds 120) -}}
scalar(max(avg_over_time(node_filesystem_size[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(max(node_filesystem_size))
{{- end -}}
`,

	// For counters, we return the increase over the requested period,
	// or the increase over the last minute if no period requested
	// (Alternatively, we could do the average change-per-minute)
	// scalar(rate({{ .StatName }}[{{ .AvgSeconds }}s]) * 60)
	query.TemplateName_PlatformStatsCounter: `
{{- if (gt .AvgSeconds 120) -}}
scalar(increase({{ .MetricName }}_{{ .StatName }}_total[{{ .AvgSeconds }}s]))
{{- else -}}
scalar(irate({{ .MetricName }}_{{ .StatName }}_total[2m]) * 60)
{{- end -}}
`,

	query.TemplateName_PlatformStatsGauge: `
{{- if (gt .AvgSeconds 120) -}}
scalar(avg_over_time({{ .MetricName }}_{{ .StatName }}[{{ .AvgSeconds }}s]))
{{- else -}}
scalar({{ .MetricName }}_{{ .StatName }})
{{- end -}}
`,
}
