package project

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/wfrsgo/templet/internal/tui"
	"gitlab.com/wfrsgo/templet/internal/utils"
)

type Project struct {
	location string
	baseDir  string
	mapDest  map[string]string
	dataVars map[string]string
	tpl      *template.Template
}

func New(location string, mdest map[string]string) (*Project, error) {
	var dir string
	var err error
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

func (p *Project) Delete() {
	os.RemoveAll(p.baseDir)
}

func (p *Project) AddData(m map[string]string) {
	p.dataVars = m
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

	err := p.copyDir(p.baseDir, dst, true)
	if err != nil {
		return err
	}

	return nil
}

func (p *Project) Init() error {
	elems := strings.SplitN(p.location, ":", 2)
	if len(elems) != 2 {
		return fmt.Errorf("`%s` no es un repositorio con formato válido (tipo:grupo/nombre)", p.location)
	}

	if d, ok := p.mapDest[elems[0]]; ok {
		return p.clone(fmt.Sprintf(d, elems[1]))
	} else if elems[0] == "file" {
		return p.copyDir(elems[1], p.baseDir)
	} else {
		return fmt.Errorf("`%s` no es un repositorio válido", p.location)
	}
}

func (p *Project) clone(url string) error {
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

	var mdst string = dst
	if len(render) > 0 && render[0] {
		mdst = utils.Replace(dst, p.dataVars)
	}

	if err := os.MkdirAll(mdst, srcInfo.Mode()); err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := src + "/" + entry.Name()
		dstPath := mdst + "/" + entry.Name()

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
	if len(render) > 0 && render[0] {
		mdst = utils.Replace(dst, p.dataVars)
	}

	if len(render) > 0 && render[0] && strings.HasSuffix(src, "meta.hjson") {
		return nil
	}

	if canRender {
		mdst = strings.Replace(mdst, ".tmpl", "", 1)
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

		err = t.Execute(dstFile, p.dataVars)
	} else {
		_, err = io.Copy(dstFile, srcFile)
	}

	if len(render) > 0 && render[0] {
		fmt.Println(tui.Colorize("󱁻 Archivo `%s` creado", mdst))
	}

	return err
}
