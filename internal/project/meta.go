package project

import (
	"os"
	"path/filepath"

	"github.com/hjson/hjson-go/v4"
	"gitlab.com/wfrsgo/templet/internal/tui"
)

type Variable struct {
	Nombre      string   `json:"nombre"`
	Descripcion string   `json:"descripcion"`
	Opciones    []string `json:"opciones,omitempty"`
}

type Variables []Variable

type Meta struct {
	Nombre      string    `json:"nombre"`
	Descripcion string    `json:"descripcion"`
	Variables   Variables `json:"variables"`
}

func ReadMeta(path string) (*Meta, error) {
	var meta Meta
	metaDest := filepath.Join(path, "meta.hjson")

	data, err := os.ReadFile(metaDest)
	if err != nil {
		return nil, err
	}

	err = hjson.Unmarshal(data, &meta)
	if err != nil {
		return nil, err
	}

	return &meta, nil
}

func (v Variables) Form() (map[string]string, error) {
	form := make(map[string]string)
	var err error
	for _, variable := range v {
		err = nil
		if len(variable.Opciones) > 0 {
			form[variable.Nombre] = tui.Choose(variable.Descripcion, variable.Opciones, variable.Opciones[0])
			continue
		} else {
			form[variable.Nombre] = tui.Prompt(variable.Descripcion, "myapp")
			if err != nil {
				return form, err
			}
		}
	}

	return form, nil
}
