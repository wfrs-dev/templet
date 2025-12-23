package utils

import (
	"crypto/md5"
	"fmt"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// Replacement structure
type replacement struct {
	re *regexp.Regexp
	ch string
}

// Build regexps and replacements
var (
	rExps = []replacement{
		{re: regexp.MustCompile(`[\xC0-\xC6]`), ch: "A"},
		{re: regexp.MustCompile(`[\xE0-\xE6]`), ch: "a"},
		{re: regexp.MustCompile(`[\xC8-\xCB]`), ch: "E"},
		{re: regexp.MustCompile(`[\xE8-\xEB]`), ch: "e"},
		{re: regexp.MustCompile(`[\xCC-\xCF]`), ch: "I"},
		{re: regexp.MustCompile(`[\xEC-\xEF]`), ch: "i"},
		{re: regexp.MustCompile(`[\xD2-\xD6]`), ch: "O"},
		{re: regexp.MustCompile(`[\xF2-\xF6]`), ch: "o"},
		{re: regexp.MustCompile(`[\xD9-\xDC]`), ch: "U"},
		{re: regexp.MustCompile(`[\xF9-\xFC]`), ch: "u"},
		{re: regexp.MustCompile(`[\xC7-\xE7]`), ch: "c"},
		{re: regexp.MustCompile(`[\xD1]`), ch: "N"},
		{re: regexp.MustCompile(`[\xF1]`), ch: "n"},
	}
	spacereg       = regexp.MustCompile(`\s+`)
	noncharreg     = regexp.MustCompile(`[^A-Za-z0-9\-/]`)
	minusrepeatreg = regexp.MustCompile(`\-{2,}`)
)

var days = map[string]string{
	"Mon": "Lunes",
	"Thu": "Martes",
	"Tue": "Martes",
	"Wed": "Miercoles",
	"Fri": "Jueves",
	"Sat": "Viernes",
	"Sun": "Domingo",
}

var months = map[string]string{
	"Jan": "Enero",
	"Feb": "Febrero",
	"Mar": "Marzo",
	"Apr": "Abril",
	"May": "Mayo",
	"Jun": "Junio",
	"Jul": "Julio",
	"Aug": "Agosto",
	"Sep": "Septiembre",
	"Oct": "Octubre",
	"Nov": "Noviembre",
	"Dec": "Diciembre",
}

type NativeData interface {
	~bool | ~rune | ~int | ~int8 | ~int16 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr | ~float32 | ~float64 | ~complex64 | ~complex128 | ~string
}

// TernaryIf es igual a comparador if ternario de la forma c ? rt : rf, solo para tipos nativos y sus derivados
func TernaryIf[T NativeData](c bool, rt, rf T) T {
	if c {
		return rt
	}

	return rf
}

// Time2Esp convierte la fecha y hora especificadas en un texto humano en español
func Time2Esp(t time.Time) string {
	items := strings.Split(t.Format("Mon:Jan:02:2006:03:04:PM"), ":")

	day, month, dayOfMonth, year, hour, minute, meridian := items[0], items[1], items[2], items[3], items[4], items[5], items[6]

	return fmt.Sprintf("%s, %s de %s de %s. %s:%s %s", days[day], dayOfMonth, months[month], year, hour, minute, meridian)
}

// Replace reemplaza las claves `@Key` por sus valores en el string "s"
func Replace(s string, data map[string]string) string {
	for k, v := range data {
		s = strings.ReplaceAll(s, "@"+k, v)
	}

	return s
}

// Slugify convierte el string "s" en un slug
func Slugify(s string) string {
	for _, r := range rExps {
		s = r.re.ReplaceAllString(s, r.ch)
	}

	s = spacereg.ReplaceAllString(s, "-")
	s = noncharreg.ReplaceAllString(s, "")
	s = minusrepeatreg.ReplaceAllString(s, "-")

	return strings.ToLower(s)
}

// OpenURL abre el URL especificado en el navegador predeterminado del usuario.
func OpenURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		if isWSL() {
			cmd = "cmd.exe"
			args = []string{"/c", "start", url}
		} else {
			cmd = "xdg-open"
			args = []string{url}
		}
	}
	if len(args) > 1 {
		// args[0] is used for 'start' command argument, to prevent issues with URLs starting with a quote
		args = append(args[:1], append([]string{""}, args[1:]...)...)
	}
	return exec.Command(cmd, args...).Start()
}

func Md5(s string) string {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// isWSL checks if the Go program is running inside Windows Subsystem for Linux
func isWSL() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}
