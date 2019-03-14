/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query templates for Prometheus

package prometheus

type TemplateVars struct {
	AvgSeconds int
}

var queryTemplates = map[string]string{
	"cpu_available":
`
{{- if (gt .AvgSeconds 0) -}}
scalar(sum (rate(node_cpu{mode='idle'}[{{ .AvgSeconds }}s])) * 1000)
{{- else -}}
scalar(sum (irate(node_cpu{mode='idle'}[5m])) * 1000)
{{- end -}}
`,
}
