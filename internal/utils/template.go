package utils

import (
	"strings"
	"text/template"
)

var funcs = template.FuncMap{
	"slugify": Slugify,
	"ucfirst": func(s string) string {
		if len(s) > 0 {
			return strings.ToUpper(s[0:1]) + s[1:]
		}
		return s
	},
	"md5":       Md5,
	"pluralize": Pluralize,
}

func NewTemplate() *template.Template {
	t := template.New("templet").Delims("{%", "%}").Funcs(funcs)

	return t
}
