/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query templates for Prometheus

package prometheus

import (
	. "github.com/nalej/infrastructure-monitor/pkg/provider/query"
)
type TemplateVars struct {
	AvgSeconds int
}

var queryTemplates = TemplateStringMap{
	TemplateName_CPU + TemplateName_Available:
`
{{- if (gt .AvgSeconds 0) -}}
scalar(sum (rate(node_cpu{mode='idle'}[{{ .AvgSeconds }}s])) * 1000)
{{- else -}}
scalar(sum (irate(node_cpu{mode='idle'}[5m])) * 1000)
{{- end -}}
`,

	TemplateName_CPU + TemplateName_Total:
`
{{- if (gt .AvgSeconds 0) -}}
scalar(avg_over_time(count(node_cpu{mode='idle'})[{{ .AvgSeconds }}s:60s]) * 1000)
{{- else -}}
scalar(count(node_cpu{mode='idle'}) * 1000)
{{- end -}}
`,

	TemplateName_Memory + TemplateName_Available:
`
{{- if (gt .AvgSeconds 0) -}}
scalar(sum(avg_over_time(node_memory_MemAvailable[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_memory_MemAvailable))
{{- end -}}
`,

	TemplateName_Memory + TemplateName_Total:
`
{{- if (gt .AvgSeconds 0) -}}
scalar(sum(avg_over_time(node_memory_MemTotal[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_memory_MemTotal))
{{- end -}}
`,

	TemplateName_Storage + TemplateName_Available:
`
{{- if (gt .AvgSeconds 0) -}}
scalar(sum(avg_over_time(node_filesystem_free[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_filesystem_free))
{{- end -}}
`,

	TemplateName_Storage + TemplateName_Total:
`
{{- if (gt .AvgSeconds 0) -}}
scalar(sum(avg_over_time(node_filesystem_size[{{ .AvgSeconds }}s])))
{{- else -}}
scalar(sum(node_filesystem_size))
{{- end -}}
`,
}
