package main

import (
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/wfrs-dev/templet/internal/console"
	"github.com/wfrs-dev/templet/internal/console/tui"
	"github.com/wfrs-dev/templet/internal/template"
)

var dest string
var vars console.SliceFlag

func exit(msg string) {
	fmt.Println(msg)
	os.Exit(1)
}

func setLog(lvl slog.Level) {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: lvl,
	})

	logger := slog.New(handler)

	slog.SetDefault(logger)
}

func main() {
	setLog(slog.LevelDebug)

	flag.StringVar(&dest, "dest", ".", "destination directory")
	flag.Var(&vars, "var", "variables to be replaced in the template (key=value)")
	flag.Usage = flag.PrintDefaults
	flag.Parse()

	filename := flag.Arg(0)

	if filename == "" {
		exit("no file specified")
	}
	if _, err := os.Stat(filename); err != nil || errors.Is(err, os.ErrNotExist) {
		exit(fmt.Sprintf("file `%s` couldn't be opened: %s", filename, err.Error()))
	}

	tpl, err := template.Load(filename)
	if err != nil {
		exit(fmt.Sprintf("couldn't load template: %s", err.Error()))
	}

	// project := template.NewProject(dest, tpl, nil)
	//
	// err = project.Generate()
	// if err != nil {
	// 	exit(fmt.Sprintf("couldn't generate project: %s", err.Error()))
	// }

	mvars, err := loadVariables(vars, tpl.Variables)
	if err != nil {
		exit(fmt.Sprintf("couldn't generate project vars: %s", err.Error()))
	}
	project := template.NewProject(dest, tpl, mvars)

	err = project.Generate()
	if err != nil {
		exit(fmt.Sprintf("couldn't generate project: %s", err.Error()))
	}

	// for _, file := range tpl.Files {
	// 	fmt.Println(">", file.Path)
	// }
}

func loadVariables(vars console.SliceFlag, variables []string) (map[string]any, error) {
	varMap := make(map[string]any)

	for _, v := range vars {
		parts := strings.Split(v, "=")
		if len(parts) != 2 {
			continue
		}
		varMap[parts[0]] = parts[1]
	}

	vinput := make([]string, 0)
	for _, v := range variables {
		if _, ok := varMap[v]; !ok {
			vinput = append(vinput, v)
		}
	}

	if len(vinput) > 0 {
		for _, v := range vinput {
			varMap[v] = template.Slugify(tui.Input(fmt.Sprintf("Value for `%s`:", v), "", func(s string) error {
				if s == "" {
					return errors.New("empty value")
				}
				return nil
			}))

			fmt.Println(tui.Colorize("<bold>%s:</> %s", v, varMap[v]))
		}
	}

	return varMap, nil
}
