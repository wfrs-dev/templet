package template

import (
	"strings"
	systpl "text/template"
	"unicode"
)

// slugTranslit maps a rune to the ASCII replacement a slug uses for it. Runes
// absent from the table are dropped, so a mapping to "" (the Cyrillic hard and
// soft signs) is equivalent to dropping the rune. Accented Greek carried by
// everyday names (ά, ή, ώ, ...) folds to the base letter so that names like
// Αθήνα survive precomposed input.
var slugTranslit = newSlugTranslit()

// newSlugTranslit builds slugTranslit from output-grouped rune sets, which
// keeps the table readable and free of duplicated mappings.
func newSlugTranslit() map[rune]string {
	groups := []struct {
		out   string
		runes string
	}{
		// Latin-1 and Latin Extended-A, uppercase and lowercase.
		{out: "a", runes: "ÀÁÂÃÄÅàáâãäåĀĂĄāăą"},
		{out: "ae", runes: "Ææ"},
		{out: "c", runes: "ÇçĆĈĊČćĉċč"},
		{out: "d", runes: "ÐðĎĐďđ"},
		{out: "e", runes: "ÈÉÊËèéêëĒĔĖĘĚēĕėęě"},
		{out: "g", runes: "ĞĜĠĢğĝġģ"},
		{out: "h", runes: "ĤĦĥħ"},
		{out: "i", runes: "ÌÍÎÏìíîïĨĪĬĮİıĩīĭį"},
		{out: "j", runes: "Ĵĵ"},
		{out: "k", runes: "Ķķ"},
		{out: "l", runes: "ĹĻĽĿŁĺļľŀł"},
		{out: "n", runes: "ÑñŃŅŇŊńņňŋ"},
		{out: "o", runes: "ÒÓÔÕÖØòóôõöøŌŎŐōŏő"},
		{out: "oe", runes: "Œœ"},
		{out: "r", runes: "ŔŖŘŕŗř"},
		{out: "s", runes: "ŚŜŞŠśŝşš"},
		{out: "ss", runes: "ß"},
		{out: "t", runes: "ŢŤŦţťŧ"},
		{out: "th", runes: "Þþ"},
		{out: "u", runes: "ÙÚÛÜùúûüŨŪŬŮŰŲũūŭůűų"},
		{out: "w", runes: "Ŵŵ"},
		{out: "y", runes: "ÝýŶŸŷÿ"},
		{out: "z", runes: "ŹŻŽźżž"},

		// Greek, including the tonos and dialytika forms.
		{out: "a", runes: "αάΑΆ"},
		{out: "b", runes: "βΒ"},
		{out: "g", runes: "γΓ"},
		{out: "d", runes: "δΔ"},
		{out: "e", runes: "εέΕΈ"},
		{out: "z", runes: "ζΖ"},
		{out: "i", runes: "ηήιίϊΐΗΉΙΊΪ"},
		{out: "th", runes: "θΘ"},
		{out: "k", runes: "κΚ"},
		{out: "l", runes: "λΛ"},
		{out: "m", runes: "μΜ"},
		{out: "n", runes: "νΝ"},
		{out: "x", runes: "ξΞ"},
		{out: "o", runes: "οόωώΟΌΩΏ"},
		{out: "p", runes: "πΠ"},
		{out: "r", runes: "ρΡ"},
		{out: "s", runes: "σςΣ"},
		{out: "t", runes: "τΤ"},
		{out: "y", runes: "υύϋΰΥΎΫ"},
		{out: "f", runes: "φΦ"},
		{out: "ch", runes: "χΧ"},
		{out: "ps", runes: "ψΨ"},

		// Cyrillic, including the Ukrainian letters.
		{out: "a", runes: "аА"},
		{out: "b", runes: "бБ"},
		{out: "v", runes: "вВ"},
		{out: "g", runes: "гГґҐ"},
		{out: "d", runes: "дД"},
		{out: "e", runes: "еЕёЁэЭєЄ"},
		{out: "zh", runes: "жЖ"},
		{out: "z", runes: "зЗ"},
		{out: "i", runes: "иИйЙіІїЇ"},
		{out: "k", runes: "кК"},
		{out: "l", runes: "лЛ"},
		{out: "m", runes: "мМ"},
		{out: "n", runes: "нН"},
		{out: "o", runes: "оО"},
		{out: "p", runes: "пП"},
		{out: "r", runes: "рР"},
		{out: "s", runes: "сС"},
		{out: "t", runes: "тТ"},
		{out: "u", runes: "уУ"},
		{out: "f", runes: "фФ"},
		{out: "kh", runes: "хХ"},
		{out: "ts", runes: "цЦ"},
		{out: "ch", runes: "чЧ"},
		{out: "sh", runes: "шШ"},
		{out: "shch", runes: "щЩ"},
		{out: "", runes: "ъЪьЬ"},
		{out: "y", runes: "ыЫ"},
		{out: "yu", runes: "юЮ"},
		{out: "ya", runes: "яЯ"},
	}

	translit := make(map[rune]string, 256)
	for _, g := range groups {
		for _, r := range g.runes {
			translit[r] = g.out
		}
	}

	return translit
}

// Slugify convierte s en un slug apto para nombres de archivo y directorio.
//
// Recorre los runes de s en orden: las letras ASCII se pasan a minusculas, los
// espacios se vuelven "-", los runes con equivalente ASCII se transliteran
// (acentos latinos, griego y cirilico) y todo lo demas se descarta. Luego las
// corridas de "-" se colapsan en una sola y se recortan los extremos. El
// resultado puede ser vacio cuando no sobrevive nada ASCII, como ocurre con un
// nombre solo en CJK, asi que las plantillas no deberian depender de slug para
// esos nombres.
func Slugify(s string) string {
	// pendingDash recuerda un separador pendiente: se escribe recien cuando
	// aparece algo despues, lo que colapsa corridas y recorta los extremos.
	pendingDash := false

	var b strings.Builder
	b.Grow(len(s))

	for _, r := range s {
		var tok string

		switch {
		case 'a' <= r && r <= 'z', '0' <= r && r <= '9', r == '/', r == '.':
			tok = string(r)
		case r == '-' || unicode.IsSpace(r):
			pendingDash = b.Len() > 0

			continue
		case 'A' <= r && r <= 'Z':
			tok = string(r + ('a' - 'A'))
		default:
			tok = slugTranslit[r]
		}

		if tok == "" {
			continue
		}

		if pendingDash {
			b.WriteByte('-')
			pendingDash = false
		}

		b.WriteString(tok)
	}

	return b.String()
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
