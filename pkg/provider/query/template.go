/*
 * Copyright (C) 2019 Nalej - All Rights Reserved
 */

// Query template constants and generic functions

package query

import (
	"text/template"
)

type TemplateName string
func (t TemplateName) String() string {
	return string(t)
}

const (
	TemplateName_Total TemplateName = "_total"
	TemplateName_Available TemplateName = "_available"

	TemplateName_CPU TemplateName = "cpu"
	TemplateName_Memory TemplateName = "memory"
	TemplateName_Storage TemplateName = "storage"
)

type TemplateStringMap map[TemplateName]string
type TemplateMap map[TemplateName]*template.Template
