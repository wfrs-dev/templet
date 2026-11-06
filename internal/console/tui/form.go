package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"charm.land/huh/v2"
)

// monochromeTheme define un tema en blanco y negro puro, compatible con
// cualquier terminal TTY (ver Monochrome).
func monochromeTheme(isDark bool) *huh.Styles {
	return NewMonochrome(isDark).Styles()
}

var theme = huh.ThemeFunc(monochromeTheme)

// run ejecuta un campo dentro de un formulario.
func run(field huh.Field) error {
	return huh.NewForm(huh.NewGroup(field)).WithShowHelp(true).Run()
}

func handleErr(err error) {
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println(Colorize("<red+bold> Operación cancelada.</>"))
		os.Exit(0)
	}
	fmt.Println(err)
}

func Input(label, defval string, validations ...func(string) error) string {
	var input string
	if len(validations) == 0 {
		validations = []func(string) error{func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("no puede estar vacío")
			}
			return nil
		}}
	}

	input = defval
	field := huh.NewInput().
		Title(label).
		Placeholder(defval).
		Value(&input).
		Prompt(Nerdfont("FaPencilSquareO", "=>") + " ").
		Validate(func(s string) error {
			for _, v := range validations {
				if err := v(s); err != nil {
					return err
				}
			}
			return nil
		}).
		WithTheme(theme)
	err := run(field)
	if err != nil {
		handleErr(err)
		return ""
	}
	return strings.TrimSpace(input)
}
