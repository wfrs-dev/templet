package project

import (
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/wfrsgo/templet/internal/utils"
)

type Project struct {
	location string
	baseDir  string
	mapDest  map[string]string
	tpl      *template.Template
}

func New(location string, mdest map[string]string) (*Project, error) {
	var dir string
	var err error
	slog.Debug(fmt.Sprintf("Directorio temporal a traves de la variable de entorno TEMPLET_TMP: %q", os.Getenv("TEMPLET_TMP")))
	if d, ok := os.LookupEnv("TEMPLET_TMP"); ok {
		dir = filepath.Join(d, "templet-"+utils.RandString(5))
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			return nil, err
		}
	} else {
		dir, err = os.MkdirTemp("", "templet-")
		if err != nil {
			return nil, err
		}
	}
	slog.Debug(fmt.Sprintf("Creando directorio temporal %q", dir))
	return &Project{
		location: location,
		baseDir:  dir,
		mapDest:  mdest,
		tpl:      utils.NewTemplate(),
	}, nil
}

func (p *Project) Dir() string {
	return p.baseDir
}

func (p *Project) Run(name string) error {
	err := p.generate(name)
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) generate(name string) error {
	cwd, _ := os.Getwd()
	dst := filepath.Join(cwd, name)

	slog.Debug(fmt.Sprintf("Generando proyecto %q, de %q", dst, p.baseDir))

	err := p.copyDir(p.baseDir, dst, true)
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) Init() error {
	slog.Debug(fmt.Sprintf("Clonando proyecto %q", p.location))
	elems := strings.SplitN(p.location, ":", 2)
	if len(elems) != 2 {
		return fmt.Errorf("`%s` no es un repositorio con formato válido (tipo:grupo/nombre)", p.location)
	}
	if d, ok := p.mapDest[elems[0]]; ok {
		return p.clone(fmt.Sprintf(d, elems[1]))
	} else if elems[0] == "file" {
		slog.Info(fmt.Sprintf("Copiando proyecto %q", p.baseDir))
		return p.copyDir(elems[1], p.baseDir)
	} else {
		return fmt.Errorf("`%s` no es un repositorio válido", p.location)
	}
}

func (p *Project) clone(url string) error {
	slog.Info(fmt.Sprintf("Clonando repositorio %q", url))
	cmd := exec.Command("git", "clone", url, p.baseDir)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

func (p *Project) copyDir(src, dst string, render ...bool) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := src + "/" + entry.Name()
		dstPath := dst + "/" + entry.Name()

		if entry.IsDir() {
			if err := p.copyDir(srcPath, dstPath, render...); err != nil {
				return err
			}
		} else {
			if err := p.copyFile(srcPath, dstPath, render...); err != nil {
				return err
			}
		}
	}

	return nil
}

func (p *Project) copyFile(src string, dst string, render ...bool) error {
	canRender := len(render) > 0 && render[0] && strings.HasSuffix(src, ".tmpl")

	var mdst string = dst
	if canRender {
		mdst = replace(strings.Replace(dst, ".tmpl", "", 1), p.mapDest)
		slog.Debug(fmt.Sprintf("> Copiando archivo %q -> %q", src, mdst))
		if strings.HasSuffix(src, "meta.hjson") {
			return nil
		}
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	srcInfo, err := srcFile.Stat()
	if err != nil {
		return err
	}

	dstFile, err := os.OpenFile(mdst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, srcInfo.Mode())
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if canRender {
		st, err := io.ReadAll(srcFile)
		if err != nil {
			return err
		}
		t, err := p.tpl.Parse(string(st))
		if err != nil {
			return err
		}

		err = t.ExecuteTemplate(dstFile, src, p.mapDest)
	} else {
		_, err = io.Copy(dstFile, srcFile)
	}

	return err
}

func replace(s string, r map[string]string) string {
	for k, v := range r {
		s = strings.ReplaceAll(s, "@"+k, v)
	}
	return s
}
