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
