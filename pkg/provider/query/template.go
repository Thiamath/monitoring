/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query template constants and generic functions

package query

import (
	"fmt"
	"strings"
	"text/template"

	"github.com/nalej/derrors"
)

type TemplateName string
func (t TemplateName) String() string {
	return string(t)
}

type TemplateVars struct {
	AvgSeconds int32
	MetricName string
	StatName string
}

const (
	TemplateName_Total TemplateName = "_total"
	TemplateName_Available TemplateName = "_available"

	TemplateName_CPU TemplateName = "cpu"
	TemplateName_Memory TemplateName = "memory"
	TemplateName_Storage TemplateName = "storage"

	TemplateName_PlatformStatsCounter TemplateName = "platformcounter"
	TemplateName_PlatformStatsGauge TemplateName = "platformgauge"
)

type TemplateStringMap map[TemplateName]string
type TemplateMap map[TemplateName]*template.Template


func (t TemplateStringMap) ParseTemplates() (TemplateMap, derrors.Error) {
	templates := make(TemplateMap, len(t))

	// Pre-parse templates
	for name, tmplStr := range(t) {
		parsed, err := template.New(name.String()).Parse(tmplStr)
		if err != nil {
			return nil, derrors.NewInternalError("failed parsing template", err)
		}
		templates[name] = parsed
	}

	return templates, nil
}

func (t TemplateMap) GetTemplateQuery(name TemplateName, vars *TemplateVars) (*Query, derrors.Error) {
	tmpl, found := t[name]
	if !found {
		return nil, derrors.NewNotFoundError(fmt.Sprintf("template %s not found", name))
	}

	var buf strings.Builder
	err := tmpl.Execute(&buf, vars)
	if err != nil {
		return nil, derrors.NewInternalError("error executing template", err)
	}

	q := &Query{
		QueryString: buf.String(),
	}

	return q, nil
}
