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
	delims, err := p.delimsFor(file)
	if err != nil {
		return err
	}

	enabled, err := p.whenEnabled(file, delims)
	if err != nil {
		return err
	}
	if !enabled {
		fmt.Println(tui.Colorize("<dim+italic>%s Skip file `%s`</>", tui.Nerdfont("MdFileCancel", "-"), file.Path))

		return nil
	}

	spath := p.generator.Load(file.Path, delims...).RenderString(p.Variables)
	if err := p.generator.Error(); err != nil {
		return fmt.Errorf("couldn't render path of %s: %w", file.Path, err)
	}

	content, err := p.renderContent(file, delims)
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

// delimsFor resolves the delimiter pair used for a file: the file-level entry
// wins over the template-level default, and an absent pair means the stdlib
// {{ }} syntax. A chosen but malformed pair is rejected before anything is
// rendered or written.
func (p *Project) delimsFor(file File) ([]string, error) {
	delims := file.Delims
	if len(delims) == 0 && p.Template != nil {
		delims = p.Template.Delims
	}

	if len(delims) != 0 && len(delims) != 2 {
		return nil, fmt.Errorf("file %s: delims must have exactly two entries, got %d", file.Path, len(delims))
	}

	return delims, nil
}

// whenEnabled reports whether the file must be generated. The when expression
// is rendered with the file's effective delimiters and the project variables,
// then parsed as a boolean: an empty result keeps the historical behaviour of
// always generating the file, and an unparseable one is an error rather than a
// silent decision.
func (p *Project) whenEnabled(file File, delims []string) (bool, error) {
	rendered := strings.TrimSpace(p.generator.Load(file.When, delims...).RenderString(p.Variables))
	if err := p.generator.Error(); err != nil {
		return false, fmt.Errorf("couldn't render when of %s: %w", file.Path, err)
	}

	switch strings.ToLower(rendered) {
	case "", "true", "t", "1", "yes", "y", "on":
		return true, nil
	case "false", "f", "0", "no", "n", "off":
		return false, nil
	default:
		return false, fmt.Errorf(
			"file %s: when must render to a boolean, got %q (delimiters in effect: %s)",
			file.Path, rendered, describeDelims(delims),
		)
	}
}

// describeDelims nombra el par de delimitadores en uso, para que un `when`
// escrito con los delimitadores equivocados sea evidente en el error.
func describeDelims(delims []string) string {
	if len(delims) == 2 {
		return fmt.Sprintf("%q and %q", delims[0], delims[1])
	}

	return "\"{{\" and \"}}\""
}

// renderContent renders the file content and applies its literal replacements.
// It never returns partially rendered content alongside an error.
func (p *Project) renderContent(file File, delims []string) ([]byte, error) {
	str := p.generator.Load(file.Content, delims...).RenderString(p.Variables)
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
