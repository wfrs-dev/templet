package main

import (
	"log/slog"
	"os"

	"github.com/integrii/flaggy"
	"gitlab.com/wfrsgo/templet/internal/project"
	"gitlab.com/wfrsgo/templet/internal/utils"
)

func main() {
	var name string
	var source string

	slog.SetLogLoggerLevel(slog.LevelDebug)

	flaggy.String(&name, "n", "name", "Nombre del proyecto (obligatorio)")

	flaggy.AddPositionalValue(&source, "source", 1, true, "nombre de la plantilla a usar en formato `grupo/nombre`")
	flaggy.ShowHelpOnUnexpectedDisable()

	flaggy.Parse()

	if name == "" {
		exit("El nombre del proyecto es obligatorio (-n, --name)")
	}

	if source == "" {
		exit("La plantilla es obligatoria")
	}

	cfg, err := utils.ReadConfig()
	exitErr(err)

	prj, err := project.New(source, cfg.Routes)
	exitErr(err)

	defer prj.Delete()

	err = prj.Init()
	exitErr(err)

	meta, err := project.ReadMeta(prj.Dir())
	exitErr(err)

	data, err := meta.Variables.Form()
	exitErr(err)
	_ = data

	prj.AddData(data)
	err = prj.Run(name)
	exitErr(err)
}

func exit(msg string) {
	slog.Error(msg)
	os.Exit(1)
}

func exitErr(err error) {
	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}
