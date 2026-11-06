package template

import (
	"bytes"
	"fmt"
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

func (g *StringGenerator) Load(content string, delims ...string) *StringGenerator {
	// Every load starts a new operation: an error from a previous one must not
	// be reported for this one.
	g.last = nil

	// The delimiter pair is optional, but a partial or oversized pair is a
	// caller error rather than something the stdlib can recover from.
	if len(delims) != 0 && len(delims) != 2 {
		g.last = fmt.Errorf("delims must have zero or two entries, got %d", len(delims))

		return g
	}

	// Each load parses into a fresh template so that neither the delimiter pair
	// nor the definitions of a previous load can leak into this one. The current
	// template is only replaced once the new one parses successfully.
	tpl := template.New("templet").Funcs(tplFunctions())
	if len(delims) == 2 {
		tpl.Delims(delims[0], delims[1])
	}

	parsed, err := tpl.Parse(content)
	if err != nil {
		g.last = err
		return g
	}

	g.tpl = parsed

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
