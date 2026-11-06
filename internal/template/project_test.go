package template

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGenerateWritesRenderedContent(t *testing.T) {
	dir := t.TempDir()
	tpl := &Template{Files: []File{{
		Path:    "{{ .Name }}.txt",
		Content: "Hello {{ upper .Name }}",
	}}}

	if err := NewProject(dir, tpl, map[string]any{"Name": "world"}).Generate(); err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "world.txt"))
	if err != nil {
		t.Fatalf("expected rendered file: %v", err)
	}
	if want := "Hello WORLD"; string(got) != want {
		t.Fatalf("file content = %q, want %q", got, want)
	}
}

func TestGenerateAppliesReplaces(t *testing.T) {
	dir := t.TempDir()
	tpl := &Template{Files: []File{{
		Path:    "out.txt",
		Content: "hello world",
		Replaces: []Replace{
			{Text: "hello", Replace: "bye"},
		},
	}}}

	if err := NewProject(dir, tpl, map[string]any{}).Generate(); err != nil {
		t.Fatalf("Generate() error = %v, want nil", err)
	}

	got, err := os.ReadFile(filepath.Join(dir, "out.txt"))
	if err != nil {
		t.Fatalf("expected rendered file: %v", err)
	}
	if want := "bye world"; string(got) != want {
		t.Fatalf("file content = %q, want %q", got, want)
	}
}

func TestGenerateFailsWithoutWritingOnContentParseError(t *testing.T) {
	dir := t.TempDir()
	tpl := &Template{Files: []File{{
		Path:    "out.txt",
		Content: "{{ if }}broken",
	}}}

	err := NewProject(dir, tpl, map[string]any{}).Generate()
	if err == nil {
		t.Fatal("Generate() error = nil, want a content render error")
	}

	if _, statErr := os.Stat(filepath.Join(dir, "out.txt")); !os.IsNotExist(statErr) {
		t.Fatalf("no file may be written when content rendering fails, stat error = %v", statErr)
	}

	entries, readErr := os.ReadDir(dir)
	if readErr != nil {
		t.Fatalf("ReadDir() error = %v", readErr)
	}
	if len(entries) != 0 {
		t.Fatalf("destination directory must stay empty, found %d entries", len(entries))
	}
}

func TestGenerateFailsWithoutTruncatingOnContentExecError(t *testing.T) {
	dir := t.TempDir()
	const existing = "previous content"

	dest := filepath.Join(dir, "out.txt")
	if err := os.WriteFile(dest, []byte(existing), 0o644); err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	tpl := &Template{Files: []File{{
		Path:    "out.txt",
		Content: "{{ .Name.Bogus }}",
	}}}

	err := NewProject(dir, tpl, map[string]any{"Name": "value"}).Generate()
	if err == nil {
		t.Fatal("Generate() error = nil, want a content render error")
	}

	got, readErr := os.ReadFile(dest)
	if readErr != nil {
		t.Fatalf("existing file must survive a failed generation: %v", readErr)
	}
	if string(got) != existing {
		t.Fatalf("existing file content = %q, want untouched %q", got, existing)
	}
}

func TestGeneratorErrorDoesNotLeakAcrossLoads(t *testing.T) {
	gen := NewGenerator()

	gen.Load("{{ if }}")
	if gen.Error() == nil {
		t.Fatal("Error() = nil after a parse error, want the parse error")
	}

	gen.Load("hello {{ .Name }}")
	if err := gen.Error(); err != nil {
		t.Fatalf("Error() = %v after a successful load, want nil (stale error from a previous operation)", err)
	}

	if got, want := gen.RenderString(map[string]any{"Name": "there"}), "hello there"; got != want {
		t.Fatalf("RenderString() = %q, want %q", got, want)
	}
}
