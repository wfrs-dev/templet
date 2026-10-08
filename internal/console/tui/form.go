package tui

import (
	"errors"
	"fmt"
	"os"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
	"github.com/charmbracelet/colorprofile"
)

type Item struct {
	Label  string
	Value  string
	Active bool
}

type Items []Item

// monochromeTheme define un tema en blanco y negro puro, compatible con
// cualquier terminal TTY (ver Monochrome).
func monochromeTheme(isDark bool) *huh.Styles {
	return NewMonochrome(isDark || TTYMode()).Styles()
}

var theme = huh.ThemeFunc(monochromeTheme)

// run ejecuta un campo dentro de un formulario. En modo TTY fuerza el perfil
// de color ANSI (4 bits) para que la salida sea valida en terminales sin
// entorno grafico.
func run(field huh.Field) error {
	form := huh.NewForm(huh.NewGroup(field)).WithShowHelp(true)

	if TTYMode() {
		form = form.WithProgramOptions(
			tea.WithOutput(os.Stderr),
			tea.WithColorProfile(colorprofile.ANSI),
		)
	}

	return form.Run()
}

func handleErr(err error) {
	if errors.Is(err, huh.ErrUserAborted) {
		fmt.Println(Colorize("<red+bold> Operación cancelada.</>"))
		os.Exit(0)
	}
	fmt.Println(err)
}

func Select(slabel string, items Items) string {
	var input string

	hopts := make([]huh.Option[string], len(items))
	for i, opt := range items {
		hopts[i] = huh.NewOption(opt.Label, opt.Value).Selected(opt.Active)
	}
	field := huh.NewSelect[string]().
		Title(slabel).
		Value(&input).
		Options(hopts...).
		Validate(func(s string) error {
			if strings.TrimSpace(s) == "" {
				return errors.New("no option selected")
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

func MultiSelect(slabel string, items Items) []string {
	var input []string

	hopts := make([]huh.Option[string], len(items))
	for i, opt := range items {
		hopts[i] = huh.NewOption(opt.Label, opt.Value).Selected(opt.Active)
	}

	field := huh.NewMultiSelect[string]().
		Title(slabel).
		Description("↑ Up 🞄 ↓ Down 🞄 / Search mode 🞄 ⏎  Show Search items/Submit 🞄 ␣ Select item 🞄 Esc Exit search mode").
		Value(&input).
		Options(hopts...).
		WithHeight(10).
		WithTheme(theme)
	err := run(field)
	if err != nil {
		handleErr(err)
		return []string{}
	}

	return input
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

func Confirm(msg string) bool {
	fmt.Print(Colorize("<bold>%s</>(y/N): ", msg))
	var char rune
	fmt.Scanf("%c", &char)

	return char == 'y' || char == 'Y' || char == 's' || char == 'S'
}
