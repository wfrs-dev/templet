package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/integrii/flaggy"
	"gitlab.com/wfrsgo/templet/internal/project"
	"gitlab.com/wfrsgo/templet/internal/tui"
	"gitlab.com/wfrsgo/templet/internal/utils"
)

var Version = "devel"

func main() {
	var name string
	var source string

	slog.SetLogLoggerLevel(slog.LevelDebug)

	flaggy.String(&name, "n", "name", "Nombre del proyecto (obligatorio)")

	flaggy.AddPositionalValue(&source, "source", 1, true, "nombre de la plantilla a usar en formato `grupo/nombre`")
	flaggy.ShowHelpOnUnexpectedDisable()
	flaggy.SetVersion(Version)

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
	err = prj.Run(name, meta.Comandos)
	exitErr(err)

	fmt.Println(
		tui.Colorize(
			fmt.Sprintf("<green>✔ Proyecto creado con éxito a las %s</>\n", time.Now().Format("01/02/2006 03:04:05 PM")),
		),
	)
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
