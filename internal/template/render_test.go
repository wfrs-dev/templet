package template

import "testing"

func TestLoadDelims(t *testing.T) {
	tests := []struct {
		name    string
		delims  []string
		content string
		want    string
	}{
		{
			name:    "no delims keeps the stdlib syntax",
			content: "Hi {{ .Name }}",
			want:    "Hi world",
		},
		{
			name:    "two delims replace the stdlib syntax",
			delims:  []string{"[[", "]]"},
			content: "Hi [[ .Name ]]",
			want:    "Hi world",
		},
		{
			name:    "two delims leave the stdlib syntax literal",
			delims:  []string{"[[", "]]"},
			content: "Hi {{ .Name }}",
			want:    "Hi {{ .Name }}",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewGenerator()

			got := gen.Load(tc.content, tc.delims...).RenderString(map[string]any{"Name": "world"})
			if err := gen.Error(); err != nil {
				t.Fatalf("Error() = %v, want nil", err)
			}
			if got != tc.want {
				t.Fatalf("RenderString() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLoadRejectsInvalidDelimCount(t *testing.T) {
	tests := []struct {
		name   string
		delims []string
		want   string
	}{
		{
			name:   "one delim",
			delims: []string{"[["},
			want:   "delims must have zero or two entries, got 1",
		},
		{
			name:   "three delims",
			delims: []string{"[[", "]]", "{{"},
			want:   "delims must have zero or two entries, got 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewGenerator()

			got := gen.Load("Hi {{ .Name }}", tc.delims...).RenderString(map[string]any{"Name": "world"})
			if err := gen.Error(); err == nil {
				t.Fatal("Error() = nil, want an invalid delims error")
			} else if err.Error() != tc.want {
				t.Fatalf("Error() = %q, want %q", err, tc.want)
			}
			if got != "" {
				t.Fatalf("RenderString() = %q, want an empty result when the load failed", got)
			}
		})
	}
}

func TestLoadParsesIntoAFreshTemplate(t *testing.T) {
	// A delimiter pair must never leak from one load into the next one, in
	// either direction.
	tests := []struct {
		name     string
		first    []string
		firstTpl string
		second   []string
		secondTp string
	}{
		{
			name:     "custom delims do not leak into the default syntax",
			first:    []string{"[[", "]]"},
			firstTpl: "Hi [[ .Name ]]",
			secondTp: "Hi {{ .Name }}",
		},
		{
			name:     "the default syntax does not leak into custom delims",
			firstTpl: "Hi {{ .Name }}",
			second:   []string{"[[", "]]"},
			secondTp: "Hi [[ .Name ]]",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gen := NewGenerator()

			if got, want := gen.Load(tc.firstTpl, tc.first...).RenderString(map[string]any{"Name": "world"}), "Hi world"; got != want {
				t.Fatalf("first RenderString() = %q, want %q (error = %v)", got, want, gen.Error())
			}
			if err := gen.Error(); err != nil {
				t.Fatalf("Error() = %v after the first load, want nil", err)
			}

			got := gen.Load(tc.secondTp, tc.second...).RenderString(map[string]any{"Name": "world"})
			if err := gen.Error(); err != nil {
				t.Fatalf("Error() = %v after the second load, want nil", err)
			}
			if want := "Hi world"; got != want {
				t.Fatalf("second RenderString() = %q, want %q", got, want)
			}
		})
	}
}
