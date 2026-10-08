package tui

import (
	"image/color"
	"sync/atomic"

	"charm.land/huh/v2"
	"charm.land/lipgloss/v2"
)

// ttyMode indica si se fuerza el modo TTY (perfil ANSI de 4 bits).
var ttyMode atomic.Bool

// SetTTYMode activa o desactiva el modo TTY. En modo TTY los formularios se
// renderizan forzando el perfil ANSI de 4 bits, evitando secuencias de color
// extendido que una terminal sin entorno grafico no puede representar.
func SetTTYMode(enabled bool) {
	ttyMode.Store(enabled)
}

// TTYMode informa si el modo TTY esta activo.
func TTYMode() bool {
	return ttyMode.Load()
}

// Monochrome es la paleta monocromatica de goretro.
//
// Solo emplea los 16 colores ANSI basicos (4 bits) para que pueda
// representarse correctamente en cualquier terminal TTY, incluso sin
// soporte de color verdadero (truecolor) ni de 256 colores.
type Monochrome struct {
	FG   color.Color // texto principal
	BG   color.Color // fondo
	Dim  color.Color // texto atenuado y bordes
	Dark bool
}

// NewMonochrome construye la paleta monocromatica segun el fondo del
// terminal. En un fondo oscuro el texto principal es blanco brillante y el
// atenuado blanco, para mantener el contraste sobre negro; en un fondo claro
// se invierten fg/bg y el atenuado pasa a gris.
func NewMonochrome(isDark bool) Monochrome {
	m := Monochrome{
		FG:   lipgloss.BrightWhite,
		BG:   lipgloss.Black,
		Dim:  lipgloss.White,
		Dark: isDark,
	}
	if !isDark {
		m.FG, m.BG = m.BG, m.FG
		m.Dim = lipgloss.BrightBlack
	}

	return m
}

// Styles traduce la paleta monocromatica a los estilos de huh.
func (m Monochrome) Styles() *huh.Styles {
	t := huh.ThemeBase(m.Dark)

	fg, bg, dim := m.FG, m.BG, m.Dim

	t.Focused.Base = t.Focused.Base.BorderForeground(dim)
	t.Focused.Card = t.Focused.Base
	t.Focused.Title = t.Focused.Title.Foreground(fg).Bold(true)
	t.Focused.NoteTitle = t.Focused.NoteTitle.Foreground(fg).Bold(true)
	t.Focused.Directory = t.Focused.Directory.Foreground(fg)
	t.Focused.Description = t.Focused.Description.Foreground(dim)
	t.Focused.ErrorIndicator = t.Focused.ErrorIndicator.Foreground(fg).Bold(true)
	t.Focused.ErrorMessage = t.Focused.ErrorMessage.Foreground(fg).Bold(true)
	t.Focused.SelectSelector = t.Focused.SelectSelector.Foreground(fg)
	t.Focused.NextIndicator = t.Focused.NextIndicator.Foreground(fg)
	t.Focused.PrevIndicator = t.Focused.PrevIndicator.Foreground(fg)
	t.Focused.Option = t.Focused.Option.Foreground(fg)
	t.Focused.MultiSelectSelector = t.Focused.MultiSelectSelector.Foreground(fg)
	t.Focused.SelectedOption = t.Focused.SelectedOption.Foreground(fg).Bold(true)
	t.Focused.SelectedPrefix = t.Focused.SelectedPrefix.Foreground(fg).Bold(true)
	t.Focused.UnselectedOption = t.Focused.UnselectedOption.Foreground(fg)
	t.Focused.UnselectedPrefix = t.Focused.UnselectedPrefix.Foreground(dim)
	t.Focused.FocusedButton = t.Focused.FocusedButton.Foreground(bg).Background(fg)
	t.Focused.BlurredButton = t.Focused.BlurredButton.Foreground(fg).Background(bg)

	t.Focused.TextInput.Cursor = t.Focused.TextInput.Cursor.Foreground(fg)
	t.Focused.TextInput.Placeholder = t.Focused.TextInput.Placeholder.Foreground(dim)
	t.Focused.TextInput.Prompt = t.Focused.TextInput.Prompt.Foreground(fg)

	t.Blurred = t.Focused
	t.Blurred.Base = t.Blurred.Base.BorderStyle(lipgloss.HiddenBorder())
	t.Blurred.Card = t.Blurred.Base
	t.Blurred.Title = t.Blurred.Title.Foreground(dim)
	t.Blurred.NoteTitle = t.Blurred.NoteTitle.Foreground(dim)
	t.Blurred.TextInput.Prompt = t.Blurred.TextInput.Prompt.Foreground(dim)
	t.Blurred.TextInput.Text = t.Blurred.TextInput.Text.Foreground(fg)

	t.Blurred.NextIndicator = lipgloss.NewStyle()
	t.Blurred.PrevIndicator = lipgloss.NewStyle()

	t.Group.Title = t.Focused.Title
	t.Group.Description = t.Focused.Description

	return t
}
