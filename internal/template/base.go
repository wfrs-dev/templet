package template

import (
	"fmt"

	"github.com/BurntSushi/toml"
)

type Replace struct {
	Text    string
	Replace string
}

type File struct {
	Path     string
	Content  string
	Replaces []Replace
	Delims   []string
	When     string
}

type Template struct {
	Files     []File
	Variables []string
	Delims    []string
}

func Load(path string) (*Template, error) {
	var tpl *Template
	_, err := toml.DecodeFile(path, &tpl)
	if err != nil {
		return nil, fmt.Errorf("template.Load: %w", err)
	}

	return tpl, err
}
