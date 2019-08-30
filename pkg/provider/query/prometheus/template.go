/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query templates for Prometheus

package prometheus

import (
	. "github.com/nalej/monitoring/pkg/provider/query"
)

// We check if we ask for an average or for the latest number with
// AvgSeconds. If it's lower than 120 (2 minutes), we query for the
// latest instant number, as an average over a short time frame
// does not make sense due to lack of data points.
var queryTemplates = TemplateStringMap{
	// node_cpu counts the number of seconds a cpu spends in a certain
	// state. irate gets the rate-of-change per second, based on the last
	// two samples in a vector - so change-of-seconds-per-second is
	// CPU usage. Multiply numbers by 1000 to return millicores.
	// Rate gets rate-of-change per second over complete vector.
	TemplateName_CPU + TemplateName_Available:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(sum (rate(node_cpu_seconds_total{mode='idle'}[{{ .AvgSeconds }}s])) * 1000)
{{- else -}}
scalar(sum (irate(node_cpu_seconds_total{mode='idle'}[2m])) * 1000)
{{- end -}}
`,

	TemplateName_CPU + TemplateName_Total:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(avg_over_time(count(node_cpu_seconds_total{mode='idle'})[{{ .AvgSeconds }}s:60s]) * 1000)
{{- else -}}
scalar(count(node_cpu_seconds_total{mode='idle'}) * 1000)
{{- end -}}
`,

	TemplateName_Memory + TemplateName_Available:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_memory_MemAvailable_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_memory_MemAvailable_bytes))
{{- end -}}
`,

	TemplateName_Memory + TemplateName_Total:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_memory_MemTotal_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_memory_MemTotal_bytes))
{{- end -}}
`,

	TemplateName_Storage + TemplateName_Available:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_filesystem_free_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_filesystem_free_bytes))
{{- end -}}
`,

	TemplateName_Storage + TemplateName_Total:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(sum(avg_over_time(node_filesystem_size_bytes[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_filesystem_size_bytes))
{{- end -}}
`,

	TemplateName_UsableStorage + TemplateName_Available:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(max(avg_over_time(node_filesystem_free[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(max(node_filesystem_free))
{{- end -}}
`,

	TemplateName_UsableStorage + TemplateName_Total:
`
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
	TemplateName_PlatformStatsCounter:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(increase({{ .MetricName }}_{{ .StatName }}_total[{{ .AvgSeconds }}s]))
{{- else -}}
scalar(irate({{ .MetricName }}_{{ .StatName }}_total[2m]) * 60)
{{- end -}}
`,

	TemplateName_PlatformStatsGauge:
`
{{- if (gt .AvgSeconds 120) -}}
scalar(avg_over_time({{ .MetricName }}_{{ .StatName }}[{{ .AvgSeconds }}s]))
{{- else -}}
scalar({{ .MetricName }}_{{ .StatName }})
{{- end -}}
`,

	TemplateName_Clusters + TemplateName_Total:
`
scalar(count(sum by (cluster_id) (node:kube_node_status_condition:selectready)))
`,

	TemplateName_Clusters + TemplateName_Healthy:
`
scalar(count((job:kube_node_status_healthy:intersect == 1) and on(cluster_id) (job:nalej_components_healthy:intersect == 1)) or vector(0))
`,

}
