package template

import "testing"

func TestTitleFunction(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		{"hello world", "Hello World"},
		{"hello_world", "Hello_world"},
		{"hello-world", "Hello-World"},
		{"it's", "It'S"},
		{"x1y2", "X1y2"},
		{"café olé", "Café Olé"},
		{"1abc 2def", "1abc 2def"},
	}

	for _, tc := range cases {
		gen := NewGenerator().Load("{{ title .In }}")
		if err := gen.Error(); err != nil {
			t.Fatalf("load error = %v", err)
		}

		if got := gen.RenderString(map[string]any{"In": tc.in}); got != tc.want {
			t.Errorf("title(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSlugFunction(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"", ""},
		// ASCII passthrough and lowercasing.
		{"Mi Proyecto", "mi-proyecto"},
		{"MiProyecto2024", "miproyecto2024"},
		// Accented Latin folds to its base letter.
		{"Caf\u00e9 Ol\u00e9", "cafe-ole"},
		{"\u00d1and\u00fa", "nandu"},
		// Special Latin forms and ligatures.
		{"Stra\u00dfe", "strasse"},
		{"\u00c6on Flux", "aeon-flux"},
		{"\u00d8resund", "oresund"},
		{"\u0141\u00f3d\u017a", "lodz"},
		{"\u00de\u00f3r", "thor"},
		// Latin Extended-A.
		{"\u015etefan \u010cupek", "stefan-cupek"},
		{"\u011e\u00f6k\u00e7eada", "gokceada"},
		// The multiplication sign must not fold to c.
		{"a\u00d7b", "ab"},
		// Greek.
		{"\u0391\u03b8\u03ae\u03bd\u03b1", "athina"},
		// Cyrillic.
		{"\u041f\u0440\u0438\u0432\u0435\u0442 \u043c\u0438\u0440", "privet-mir"},
		{"\u0429\u0443\u043a\u0430", "shchuka"},
		{"\u042e\u043b\u0456\u044f \u0404\u0432\u0430", "yuliya-eva"},
		{"\u042a\u044c\u0431\u044b\u0442\u044c", "byt"},
		// Mixed scripts and digits.
		{"Proyecto \u041f\u0440\u0438\u0432\u0435\u0442 2024", "proyecto-privet-2024"},
		// Scripts without an ASCII equivalent vanish.
		{"\u65e5\u672c\u8a9e", ""},
		{"emoji \U0001f680 name", "emoji-name"},
		// Separators that survive and characters that do not.
		{"api/v1.2", "api/v1.2"},
		{"mi_proyecto", "miproyecto"},
		// Whitespace runs collapse and edge dashes are trimmed.
		{"  hola   mundo  ", "hola-mundo"},
		{"hola --- mundo", "hola-mundo"},
		{"---hola---", "hola"},
	}

	for _, tc := range cases {
		gen := NewGenerator().Load("{{ slug .In }}")
		if err := gen.Error(); err != nil {
			t.Fatalf("load error = %v", err)
		}

		if got := gen.RenderString(map[string]any{"In": tc.in}); got != tc.want {
			t.Errorf("slug(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestSlugFunctionDropsMultiplicationSign(t *testing.T) {
	gen := NewGenerator().Load("{{ slug .In }}")
	if err := gen.Error(); err != nil {
		t.Fatalf("load error = %v", err)
	}

	got := gen.RenderString(map[string]any{"In": "a\u00d7b"})
	if got == "acb" {
		t.Fatalf("slug(%q) = %q, want the multiplication sign dropped, not folded to c", "a\u00d7b", got)
	}

	if want := "ab"; got != want {
		t.Fatalf("slug(%q) = %q, want %q", "a\u00d7b", got, want)
	}
}
