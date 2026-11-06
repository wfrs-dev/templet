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

// Version se define en tiempo de compilacion con `-ldflags "-X main.Version=<v>"`
var Version = "development"

// exit informa el error por stderr y termina con codigo 1
func exit(msg string) {
	fmt.Fprintln(os.Stderr, msg)
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

	showVersion := false
	flag.BoolVar(&showVersion, "version", false, "print the version and exit")
	flag.StringVar(&dest, "dest", ".", "destination directory")
	flag.Var(&vars, "var", "variables to be replaced in the template (key=value)")
	flag.Usage = flag.PrintDefaults
	flag.Parse()

	if showVersion {
		fmt.Printf("templet %s\n", Version)
		return
	}

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
	return collectVariables(vars, variables, askValue)
}

// askValue pide por consola el valor de una variable no recibida por `-var`
func askValue(name string) string {
	return tui.Input(fmt.Sprintf("Value for `%s`:", name), "", func(s string) error {
		if s == "" {
			return errors.New("empty value")
		}
		return nil
	})
}

// collectVariables resuelve los valores de las variables de la plantilla: los
// recibidos con `-var` y, para el resto, los que devuelve `ask`
func collectVariables(vars console.SliceFlag, variables []string, ask func(name string) string) (map[string]any, error) {
	varMap := make(map[string]any)

	for _, v := range vars {
		key, value, ok := strings.Cut(v, "=")
		if !ok {
			return nil, fmt.Errorf("invalid variable %q: expected key=value", v)
		}
		if key == "" {
			return nil, fmt.Errorf("invalid variable %q: empty key", v)
		}

		varMap[key] = value
	}

	vinput := make([]string, 0)
	for _, v := range variables {
		if _, ok := varMap[v]; !ok {
			vinput = append(vinput, v)
		}
	}

	if len(vinput) > 0 {
		for _, v := range vinput {
			// Valores verbatim: no se slugifican, el slug es responsabilidad de
			// la plantilla (`{{ slug .Name }}`).
			varMap[v] = ask(v)

			fmt.Println(tui.Colorize("<bold>%s:</> %s", v, varMap[v]))
		}
	}

	return varMap, nil
}
