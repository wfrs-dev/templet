package template

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/wfrs-dev/templet/internal/console/tui"
)

type Project struct {
	Destination string
	Template    *Template
	Variables   map[string]any
	generator   *StringGenerator
}

func NewProject(destination string, tpl *Template, variables map[string]any) *Project {
	return &Project{
		Destination: destination,
		Template:    tpl,
		Variables:   variables,
		generator:   NewGenerator(),
	}
}

func (p *Project) Generate() error {
	for _, file := range p.Template.Files {
		if err := p.GenerateFile(file); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) GenerateFile(file File) error {
	spath := p.generator.Load(file.Path).RenderString(p.Variables)
	fmt.Println(tui.Colorize("<blue+italic>%s Generate file `%s`</>", tui.Nerdfont("MdFileCog", "> "), spath))
	if p.generator.Error() != nil {
		return fmt.Errorf("couldn't generate file %s: %w", file.Path, p.generator.Error())
	}

	destDir := filepath.Join(p.Destination, filepath.Dir(spath))
	os.MkdirAll(destDir, 0755)
	var buf bytes.Buffer
	if len(file.Replaces) > 0 {
		str := p.generator.Load(file.Content).RenderString(p.Variables)
		for _, replace := range file.Replaces {
			fmt.Println(
				tui.Colorize(
					"    <purple>%s Reemplazo:</> <yellow>%s</> <purple>%s</> <yellow>%s</>",
					tui.Nerdfont("CodReplace", "-"),
					replace.Replace,
					tui.Nerdfont("MdArrowRight", "->"),
					replace.Text,
				),
			)
			str = strings.ReplaceAll(str, replace.Text, replace.Replace)
		}
		buf.WriteString(str)
	} else {
		buf = p.generator.Load(file.Content).Render(p.Variables)
	}

	dest := filepath.Join(p.Destination, spath)
	err := os.WriteFile(dest, buf.Bytes(), 0644)
	if err != nil {
		return fmt.Errorf("couldn't write file %s: %w", spath, err)
	}

	return nil
}
