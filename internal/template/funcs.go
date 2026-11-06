package template

import (
	"regexp"
	"strings"
	systpl "text/template"
	"unicode"
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
	noncharreg     = regexp.MustCompile(`[^A-Za-z0-9\-/\.]`)
	minusrepeatreg = regexp.MustCompile(`\-{2,}`)
)

func Slugify(s string) string {
	for _, r := range rExps {
		s = r.re.ReplaceAllString(s, r.ch)
	}

	s = spacereg.ReplaceAllString(s, "-")
	s = noncharreg.ReplaceAllString(s, "")
	s = minusrepeatreg.ReplaceAllString(s, "-")

	return strings.ToLower(s)
}

// title retorna una copia de s con la inicial de cada palabra en mayuscula.
// Reemplaza a strings.Title, deprecado desde Go 1.18, preservando su semantica.
func title(s string) string {
	prev := ' '

	return strings.Map(func(r rune) rune {
		if isSeparator(prev) {
			prev = r
			return unicode.ToTitle(r)
		}
		prev = r

		return r
	}, s)
}

// isSeparator indica si r separa palabras para el formateo de titulos
func isSeparator(r rune) bool {
	if r <= 0x7F {
		switch {
		case '0' <= r && r <= '9':
			return false
		case 'a' <= r && r <= 'z':
			return false
		case 'A' <= r && r <= 'Z':
			return false
		case r == '_':
			return false
		}

		return true
	}

	if unicode.IsLetter(r) || unicode.IsDigit(r) {
		return false
	}

	return unicode.IsSpace(r)
}

func tplFunctions() systpl.FuncMap {
	return systpl.FuncMap{
		"upper":       strings.ToUpper,
		"lower":       strings.ToLower,
		"title":       title,
		"starts":      strings.HasPrefix,
		"ends":        strings.HasSuffix,
		"contains":    strings.Contains,
		"containsAny": strings.ContainsAny,
		"repeat":      strings.Repeat,
		"replace":     strings.ReplaceAll,

		"slug": Slugify,
	}
}
