package template

import (
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
	if err := p.generator.Error(); err != nil {
		return fmt.Errorf("couldn't render path of %s: %w", file.Path, err)
	}

	content, err := p.renderContent(file)
	if err != nil {
		return err
	}

	dest := filepath.Join(p.Destination, spath)
	if err := writeFile(dest, content); err != nil {
		return fmt.Errorf("couldn't write file %s: %w", spath, err)
	}

	fmt.Println(tui.Colorize("<blue+italic>%s Generate file `%s`</>", tui.Nerdfont("MdFileCog", "> "), spath))

	return nil
}

// renderContent renders the file content and applies its literal replacements.
// It never returns partially rendered content alongside an error.
func (p *Project) renderContent(file File) ([]byte, error) {
	str := p.generator.Load(file.Content).RenderString(p.Variables)
	if err := p.generator.Error(); err != nil {
		return nil, fmt.Errorf("couldn't render content of %s: %w", file.Path, err)
	}

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

	return []byte(str), nil
}

// writeFile writes content to dest through a temporary file in the same
// directory, so a failing write never leaves a truncated or empty file behind.
func writeFile(dest string, content []byte) error {
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".templet-*")
	if err != nil {
		return err
	}

	tmpName := tmp.Name()
	defer func() { _ = os.Remove(tmpName) }()

	if _, err := tmp.Write(content); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmpName, 0o644); err != nil {
		return err
	}

	return os.Rename(tmpName, dest)
}
