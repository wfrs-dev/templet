package utils

import (
	"text/template"
)

var funcs = template.FuncMap{
	"slugify": Slugify,
}

func NewTemplate() *template.Template {
	t := template.New("templet").Delims("{%", "%}").Funcs(funcs)

	return t
}
