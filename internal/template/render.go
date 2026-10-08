package template

import (
	"bytes"
	"strings"
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

	if len(replaces) > 0 {
		str := buf.String()
		for _, r := range replaces {
			str = strings.ReplaceAll(str, r.Text, r.Replacer)
		}
		buf = *bytes.NewBufferString(str)
	}

	return buf
}

func (g *StringGenerator) Error() error {
	return g.last
}
