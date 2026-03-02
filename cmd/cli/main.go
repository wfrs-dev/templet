package main

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/integrii/flaggy"
	"gitlab.com/wfrsgo/templet/internal/project"
	"gitlab.com/wfrsgo/templet/internal/tui"
)

var Version = "devel"

func main() {
	var name string
	var source string
	var repo string

	slog.SetLogLoggerLevel(slog.LevelDebug)

	flaggy.SetName(nombreEjecutable())
	flaggy.SetDescription("Herramienta para la creación de estructuras de archivos a partir de plantillas")

	flaggy.String(&name, "n", "name", "Nombre del proyecto (obligatorio)")
	flaggy.String(&repo, "r", "repo", "Ruta del repositorio git donde descargar la plantilla (host[:puerto], opcional)")

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

	prj, err := project.New(source, repo)
	exitErr(err)

	defer prj.Delete()

	err = prj.Init()
	exitErr(err)

	meta, err := project.ReadMeta(prj.Dir())
	exitErr(err)

	title := fmt.Sprintf("Asistente para la creación del proyecto %s:", name)
	fmt.Println(tui.Colorize("<red><bold>%s\n%s</>", title, strings.Repeat("=", utf8.RuneCountInString(title))))
	fmt.Println()

	data, err := meta.Variables.Form()
	exitErr(err)
	_ = data

	prj.AddData(data)
	err = prj.Run(name, meta.Comandos)
	exitErr(err)

	fmt.Println(
		tui.Colorize(
			fmt.Sprintf(
				"<green>✔ Proyecto <bold>%s</><green> creado con éxito a las %s</>\n",
				name,
				time.Now().Format("01/02/2006 03:04:05 PM"),
			),
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

// nombreEjecutable devuelve el nombre del archivo ejecutable del programa sin extensión.
func nombreEjecutable() string {
	var name string
	exePath, err := os.Executable()
	if err != nil {
		// En caso de error, recurrimos a os.Args[0]
		name = filepath.Base(os.Args[0])
	} else {
		name = filepath.Base(exePath)
	}
	return strings.TrimSuffix(name, filepath.Ext(name))
}
