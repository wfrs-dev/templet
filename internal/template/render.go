package template

import (
	"bytes"
	"text/template"
)

type StringGenerator struct {
	tpl  *template.Template
	last error
}

func NewGenerator() *StringGenerator {
	tpl := template.New("templet")
	tpl.Funcs(tplFunctions())

	return &StringGenerator{
		tpl: tpl,
	}
}

func (g *StringGenerator) Load(content string) *StringGenerator {
	_, err := g.tpl.Parse(content)
	if err != nil {
		g.last = err
		return g
	}

	return g
}

func (g *StringGenerator) Render(data map[string]any, replaces ...Replace) bytes.Buffer {
	var buf bytes.Buffer

	err := g.tpl.Execute(&buf, data)
	if err != nil {
		g.last = err
	}

	return buf
}

func (g *StringGenerator) RenderString(data map[string]any, replaces ...Replace) string {
	buf := g.Render(data, replaces...)

	return buf.String()
}

func (g *StringGenerator) Error() error {
	return g.last
}
