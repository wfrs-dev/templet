package tui

import "testing"

func TestNerdfont(t *testing.T) {
	glyph, known := fontMap["MdFileCog"]
	if !known {
		t.Fatal("fixture error: MdFileCog is missing from fontMap")
	}

	cases := []struct {
		name string
		icon string
		cdef []string
		want string
	}{
		{"glyph for a known icon name", "MdFileCog", []string{"> "}, glyph},
		{"fallback for an unknown icon name", "MdNotAnIcon", []string{"->"}, "->"},
		{"empty for an unknown icon without a fallback", "MdNotAnIcon", nil, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := Nerdfont(tc.icon, tc.cdef...); got != tc.want {
				t.Errorf("Nerdfont(%q, %q) = %q, want %q", tc.icon, tc.cdef, got, tc.want)
			}
		})
	}
}
