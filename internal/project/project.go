package project

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"gitlab.com/wfrsgo/templet/internal/tui"
	"gitlab.com/wfrsgo/templet/internal/utils"
)

type Project struct {
	location string
	baseDir  string
	dest     string
	dataVars map[string]string
	tpl      *template.Template
}

func New(location, dest string) (*Project, error) {
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
		dest:     dest,
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

func (p *Project) Run(name string, cmds Commands) error {
	err := p.generate(name)
	if err != nil {
		return err
	}

	return p.comandos(name, cmds)
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

func (p *Project) comandos(name string, cmds Commands) error {
	for _, cmd := range cmds {
		fmt.Println(tui.Colorize("<purple>󱆃 Ejecutando comando `%s`</>", cmd))
		if err := utils.Execute(name, "sh", "-c", cmd); err != nil {
			return err
		}
	}

	return nil
}

func (p *Project) Init() error {
	elems := strings.SplitN(p.location, ":", 2)
	if len(elems) != 2 {
		return fmt.Errorf("`%s` no es un repositorio con formato válido (tipo:grupo/nombre)", p.location)
	}

	if elems[0] == "git" && p.dest != "" {
		return p.clone(elems[1])
	} else if elems[0] == "file" {
		return p.copyDir(elems[1], p.baseDir)
	} else {
		return fmt.Errorf("`%s` no es un repositorio válido", p.location)
	}
}

func (p *Project) clone(loc string) error {
	url := fmt.Sprintf("http://%s/%s.git", p.dest, loc)
	if err := utils.Execute("", "git", "clone", url, p.baseDir); err != nil {
		return err
	}

	return nil
}

func (p *Project) copyDir(src, dst string, render ...bool) error {
	if strings.Contains(src, ".git") {
		return nil
	}

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
		if err != nil {
			return err
		}
	} else {
		_, err = io.Copy(dstFile, srcFile)
	}

	if len(render) > 0 && render[0] {
		fmt.Println(tui.Colorize("<blue>󱁻 Archivo `%s` creado</>", mdst))
	}

	return err
}
