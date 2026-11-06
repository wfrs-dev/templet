package tui

import (
	"os"
	"testing"
)

// unsetColorEnv removes NO_COLOR for the duration of the test.
func unsetColorEnv(t *testing.T) {
	t.Helper()

	old, had := os.LookupEnv("NO_COLOR")
	if err := os.Unsetenv("NO_COLOR"); err != nil {
		t.Fatalf("Unsetenv() error = %v", err)
	}

	t.Cleanup(func() {
		if had {
			_ = os.Setenv("NO_COLOR", old)
		}
	})
}

func TestColorizeEmitsEscapeSequences(t *testing.T) {
	unsetColorEnv(t)

	if got, want := Colorize("<reset>hi</>"), "\033[0mhi\033[0m"; got != want {
		t.Fatalf("Colorize() = %q, want %q", got, want)
	}
}

func TestColorizeStripsTagsWhenNoColor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")

	if got, want := Colorize("<reset>hi</>"), "hi"; got != want {
		t.Fatalf("Colorize() = %q, want %q", got, want)
	}
}
