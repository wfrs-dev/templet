package template

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
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

func TestGenerateWithDelims(t *testing.T) {
	tests := []struct {
		name           string
		templateDelims []string
		file           File
		variables      map[string]any
		wantPath       string
		wantContent    string
	}{
		{
			name:        "per-file delims render the content",
			file:        File{Path: "out.txt", Content: "Hello [[ .Name ]]", Delims: []string{"[[", "]]"}},
			variables:   map[string]any{"Name": "world"},
			wantPath:    "out.txt",
			wantContent: "Hello world",
		},
		{
			name:        "per-file delims apply to the path",
			file:        File{Path: "[[ .Name ]].txt", Content: "static", Delims: []string{"[[", "]]"}},
			variables:   map[string]any{"Name": "world"},
			wantPath:    "world.txt",
			wantContent: "static",
		},
		{
			name:           "template-level delims apply to a file without override",
			templateDelims: []string{"[[", "]]"},
			file:           File{Path: "out.txt", Content: "Hello [[ .Name ]]"},
			variables:      map[string]any{"Name": "world"},
			wantPath:       "out.txt",
			wantContent:    "Hello world",
		},
		{
			name:           "per-file delims override the template-level ones",
			templateDelims: []string{"[[", "]]"},
			file:           File{Path: "out.txt", Content: "Hello << .Name >>", Delims: []string{"<<", ">>"}},
			variables:      map[string]any{"Name": "world"},
			wantPath:       "out.txt",
			wantContent:    "Hello world",
		},
		{
			name:        "no delims keep the stdlib syntax",
			file:        File{Path: "out.txt", Content: "Hello {{ .Name }}"},
			variables:   map[string]any{"Name": "world"},
			wantPath:    "out.txt",
			wantContent: "Hello world",
		},
		{
			name:        "custom delims emit the stdlib syntax verbatim",
			file:        File{Path: "out.txt", Content: "raw: {{ .Name }} delimited: [[ .Name ]]", Delims: []string{"[[", "]]"}},
			variables:   map[string]any{"Name": "world"},
			wantPath:    "out.txt",
			wantContent: "raw: {{ .Name }} delimited: world",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{Files: []File{tc.file}, Delims: tc.templateDelims}

			if err := NewProject(dir, tpl, tc.variables).Generate(); err != nil {
				t.Fatalf("Generate() error = %v, want nil", err)
			}

			got, err := os.ReadFile(filepath.Join(dir, tc.wantPath))
			if err != nil {
				t.Fatalf("expected rendered file %q: %v", tc.wantPath, err)
			}
			if string(got) != tc.wantContent {
				t.Fatalf("file content = %q, want %q", got, tc.wantContent)
			}
		})
	}
}

func TestGenerateRejectsInvalidDelims(t *testing.T) {
	tests := []struct {
		name           string
		templateDelims []string
		fileDelims     []string
		wantErr        string
	}{
		{
			name:       "one per-file entry",
			fileDelims: []string{"[["},
			wantErr:    "file out.txt: delims must have exactly two entries, got 1",
		},
		{
			name:       "three per-file entries",
			fileDelims: []string{"[[", "]]", "{{"},
			wantErr:    "file out.txt: delims must have exactly two entries, got 3",
		},
		{
			name:           "one template-level entry",
			templateDelims: []string{"[["},
			wantErr:        "file out.txt: delims must have exactly two entries, got 1",
		},
		{
			name:           "three template-level entries",
			templateDelims: []string{"[[", "]]", "{{"},
			wantErr:        "file out.txt: delims must have exactly two entries, got 3",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{
				Delims: tc.templateDelims,
				Files:  []File{{Path: "out.txt", Content: "Hello {{ .Name }}", Delims: tc.fileDelims}},
			}

			err := NewProject(dir, tpl, map[string]any{"Name": "world"}).Generate()
			if err == nil {
				t.Fatal("Generate() error = nil, want an invalid delims error")
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("Generate() error = %q, want %q", err, tc.wantErr)
			}

			entries, readErr := os.ReadDir(dir)
			if readErr != nil {
				t.Fatalf("ReadDir() error = %v", readErr)
			}
			if len(entries) != 0 {
				t.Fatalf("destination directory must stay empty, found %d entries", len(entries))
			}
		})
	}
}

// dirEntries returns the sorted names of everything the generation left in dir.
func dirEntries(t *testing.T, dir string) []string {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir(%q) error = %v", dir, err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		names = append(names, entry.Name())
	}
	sort.Strings(names)

	return names
}

func TestGenerateWithWhen(t *testing.T) {
	tests := []struct {
		name           string
		templateDelims []string
		files          []File
		variables      map[string]any
		wantFiles      map[string]string
	}{
		{
			name:      "when that renders true writes the file",
			files:     []File{{Path: "out.txt", Content: "body", When: "true"}},
			variables: map[string]any{},
			wantFiles: map[string]string{"out.txt": "body"},
		},
		{
			name: "when that renders false skips the file and the next one is still generated",
			files: []File{
				{Path: "skipped.txt", Content: "body", When: "false"},
				{Path: "next.txt", Content: "next body"},
			},
			variables: map[string]any{},
			wantFiles: map[string]string{"next.txt": "next body"},
		},
		{
			name:      "absent when keeps generating the file",
			files:     []File{{Path: "out.txt", Content: "body"}},
			variables: map[string]any{},
			wantFiles: map[string]string{"out.txt": "body"},
		},
		{
			name:      "when rendered from a variable with yes writes the file",
			files:     []File{{Path: "docker.txt", Content: "docker", When: "{{ .WithDocker }}"}},
			variables: map[string]any{"WithDocker": "yes"},
			wantFiles: map[string]string{"docker.txt": "docker"},
		},
		{
			name:      "when rendered from a variable with off skips the file",
			files:     []File{{Path: "docker.txt", Content: "docker", When: "{{ .WithDocker }}"}},
			variables: map[string]any{"WithDocker": "off"},
			wantFiles: map[string]string{},
		},
		{
			name:      "when honours per-file delims",
			files:     []File{{Path: "docker.txt", Content: "docker", Delims: []string{"[[", "]]"}, When: "[[ .WithDocker ]]"}},
			variables: map[string]any{"WithDocker": "true"},
			wantFiles: map[string]string{"docker.txt": "docker"},
		},
		{
			name:      "when honours per-file delims when it renders false",
			files:     []File{{Path: "docker.txt", Content: "docker", Delims: []string{"[[", "]]"}, When: "[[ .WithDocker ]]"}},
			variables: map[string]any{"WithDocker": "false"},
			wantFiles: map[string]string{},
		},
		{
			name: "when is evaluated before the path is rendered",
			files: []File{
				{Path: "{{ .Name.Bogus }}.txt", Content: "body", When: "false"},
				{Path: "next.txt", Content: "next body"},
			},
			variables: map[string]any{"Name": "value"},
			wantFiles: map[string]string{"next.txt": "next body"},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{Files: tc.files, Delims: tc.templateDelims}

			if err := NewProject(dir, tpl, tc.variables).Generate(); err != nil {
				t.Fatalf("Generate() error = %v, want nil", err)
			}

			names := dirEntries(t, dir)
			want := make([]string, 0, len(tc.wantFiles))
			for name := range tc.wantFiles {
				want = append(want, name)
			}
			sort.Strings(want)

			if len(names) != len(want) {
				t.Fatalf("destination entries = %v, want %v", names, want)
			}
			for i, name := range names {
				if name != want[i] {
					t.Fatalf("destination entries = %v, want %v", names, want)
				}
			}

			for name, wantContent := range tc.wantFiles {
				got, err := os.ReadFile(filepath.Join(dir, name))
				if err != nil {
					t.Fatalf("expected generated file %q: %v", name, err)
				}
				if string(got) != wantContent {
					t.Fatalf("file %q content = %q, want %q", name, got, wantContent)
				}
			}
		})
	}
}

func TestGenerateWhenAcceptsBooleanAliases(t *testing.T) {
	tests := []struct {
		when          string
		wantGenerated bool
	}{
		{when: "true", wantGenerated: true},
		{when: "TRUE", wantGenerated: true},
		{when: "t", wantGenerated: true},
		{when: "T", wantGenerated: true},
		{when: "1", wantGenerated: true},
		{when: "yes", wantGenerated: true},
		{when: "YES", wantGenerated: true},
		{when: "y", wantGenerated: true},
		{when: "on", wantGenerated: true},
		{when: "false", wantGenerated: false},
		{when: "False", wantGenerated: false},
		{when: "f", wantGenerated: false},
		{when: "0", wantGenerated: false},
		{when: "no", wantGenerated: false},
		{when: "n", wantGenerated: false},
		{when: "off", wantGenerated: false},
		{when: "OFF", wantGenerated: false},
	}

	for _, tc := range tests {
		t.Run(tc.when, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{Files: []File{{Path: "out.txt", Content: "body", When: tc.when}}}

			if err := NewProject(dir, tpl, map[string]any{}).Generate(); err != nil {
				t.Fatalf("Generate() error = %v, want nil", err)
			}

			_, statErr := os.Stat(filepath.Join(dir, "out.txt"))
			if tc.wantGenerated && statErr != nil {
				t.Fatalf("when = %q must generate the file, stat error = %v", tc.when, statErr)
			}
			if !tc.wantGenerated && !os.IsNotExist(statErr) {
				t.Fatalf("when = %q must skip the file, stat error = %v", tc.when, statErr)
			}

			if !tc.wantGenerated {
				if names := dirEntries(t, dir); len(names) != 0 {
					t.Fatalf("a skipped file must leave nothing behind, found %v", names)
				}
			}
		})
	}
}

func TestGenerateRejectsInvalidWhen(t *testing.T) {
	tests := []struct {
		name    string
		when    string
		delims  []string
		wantErr string
	}{
		{
			name:    "unparseable value",
			when:    "maybe",
			wantErr: `file out.txt: when must render to a boolean, got "maybe" (delimiters in effect: "{{" and "}}")`,
		},
		{
			name:    "unparseable value from a variable",
			when:    "{{ .WithDocker }}",
			wantErr: `file out.txt: when must render to a boolean, got "sure" (delimiters in effect: "{{" and "}}")`,
		},
		{
			name:    "custom delims are named in the error",
			when:    "{{ .WithDocker }}",
			delims:  []string{"[[", "]]"},
			wantErr: `file out.txt: when must render to a boolean, got "{{ .WithDocker }}" (delimiters in effect: "[[" and "]]")`,
		},
		{
			name:    "template error",
			when:    "{{ if }}",
			wantErr: "couldn't render when of out.txt:",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{Files: []File{
				{Path: "out.txt", Content: "body", When: tc.when, Delims: tc.delims},
				{Path: "next.txt", Content: "next body"},
			}}

			err := NewProject(dir, tpl, map[string]any{"WithDocker": "sure"}).Generate()
			if err == nil {
				t.Fatal("Generate() error = nil, want an invalid when error")
			}
			if !strings.Contains(err.Error(), tc.wantErr) {
				t.Fatalf("Generate() error = %q, want it to contain %q", err, tc.wantErr)
			}

			if names := dirEntries(t, dir); len(names) != 0 {
				t.Fatalf("an invalid when must write nothing, found %v", names)
			}
		})
	}
}

func TestGenerateWithMode(t *testing.T) {
	tests := []struct {
		name      string
		file      File
		variables map[string]any
		wantPath  string
		wantPerm  os.FileMode
	}{
		{
			name:     "explicit 0755 is executable",
			file:     File{Path: "scripts/build.sh", Content: "#!/bin/sh\n", Mode: "0755"},
			wantPath: "scripts/build.sh",
			wantPerm: 0o755,
		},
		{
			name:     "explicit 0600 is private",
			file:     File{Path: "secret.txt", Content: "token\n", Mode: "0600"},
			wantPath: "secret.txt",
			wantPerm: 0o600,
		},
		{
			name:     "absent mode keeps the 0644 default",
			file:     File{Path: "README.md", Content: "# demo\n"},
			wantPath: "README.md",
			wantPerm: 0o644,
		},
		{
			name: "mode survives the temp file and rename path with delims and when",
			file: File{
				Path:    "scripts/build.sh",
				Content: "#!/bin/sh\necho [[ .Name ]]\n",
				Delims:  []string{"[[", "]]"},
				When:    "[[ .WithScript ]]",
				Mode:    "0755",
			},
			variables: map[string]any{"Name": "demo", "WithScript": "yes"},
			wantPath:  "scripts/build.sh",
			wantPerm:  0o755,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{Files: []File{tc.file}}

			if err := NewProject(dir, tpl, tc.variables).Generate(); err != nil {
				t.Fatalf("Generate() error = %v, want nil", err)
			}

			info, statErr := os.Stat(filepath.Join(dir, tc.wantPath))
			if statErr != nil {
				t.Fatalf("expected generated file %q: %v", tc.wantPath, statErr)
			}
			if got := info.Mode().Perm(); got != tc.wantPerm {
				t.Fatalf("file %q permissions = %04o, want %04o", tc.wantPath, got, tc.wantPerm)
			}
		})
	}
}

func TestGenerateRejectsInvalidMode(t *testing.T) {
	tests := []struct {
		name    string
		mode    string
		wantErr string
	}{
		{
			name:    "trailing non octal digit",
			mode:    "755x",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "755x"`,
		},
		{
			name:    "9 is not an octal digit",
			mode:    "0999",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "0999"`,
		},
		{
			name:    "not a number",
			mode:    "abc",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "abc"`,
		},
		{
			name:    "above 0777",
			mode:    "8888",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "8888"`,
		},
		{
			name:    "setuid bit is not honoured",
			mode:    "4755",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "4755"`,
		},
		{
			name:    "sticky bit is not honoured",
			mode:    "1000",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "1000"`,
		},
		{
			name:    "above 07777",
			mode:    "0177777",
			wantErr: `file scripts/build.sh: mode must be an octal permission literal such as "0755", got "0177777"`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			tpl := &Template{Files: []File{
				{Path: "scripts/build.sh", Content: "#!/bin/sh\n", Mode: tc.mode},
			}}

			err := NewProject(dir, tpl, map[string]any{}).Generate()
			if err == nil {
				t.Fatal("Generate() error = nil, want an invalid mode error")
			}
			if err.Error() != tc.wantErr {
				t.Fatalf("Generate() error = %q, want %q", err, tc.wantErr)
			}

			if names := dirEntries(t, dir); len(names) != 0 {
				t.Fatalf("an invalid mode must write nothing, found %v", names)
			}
		})
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
