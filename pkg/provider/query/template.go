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
 *
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
	TemplateName_UsableStorage TemplateName = "usablestorage"

	TemplateName_PlatformStatsCounter TemplateName = "platformcounter"
	TemplateName_PlatformStatsGauge TemplateName = "platformgauge"
)

func GetPlatformTemplateName(m MetricCounter) (TemplateName, derrors.Error) {
	// Determine template based on value type (counter, gauge)
	var templateName TemplateName
	valType, found := CounterMap[m]
	if !found {
		return "", derrors.NewUnavailableError("no appropriate statistic available")
	}

	switch valType {
	case ValueCounter:
		templateName = TemplateName_PlatformStatsCounter
	case ValueGauge:
		templateName = TemplateName_PlatformStatsGauge
	default:
		return "", derrors.NewUnavailableError("no appropriate query template available")
	}

	return templateName, nil
}

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

	if vars == nil {
		vars = &TemplateVars{}
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
