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
	// Every load starts a new operation: an error from a previous one must not
	// be reported for this one.
	g.last = nil

	_, err := g.tpl.Parse(content)
	if err != nil {
		g.last = err
		return g
	}

	return g
}

func (g *StringGenerator) Render(data map[string]any, replaces ...Replace) bytes.Buffer {
	var buf bytes.Buffer

	// A failed Load must not be masked by the execution that follows it.
	if g.last != nil {
		return buf
	}

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
