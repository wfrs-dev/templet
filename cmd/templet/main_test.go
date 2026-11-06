package main

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wfrs-dev/templet/internal/console"
)

// TestMain lets the test binary run the CLI itself as a subprocess, so the
// tests can assert the real streams and the real exit code.
func TestMain(m *testing.M) {
	if os.Getenv("TEMPLET_TEST_SUBPROCESS") == "1" {
		main()
		return
	}

	os.Exit(m.Run())
}

func runCLI(t *testing.T, args ...string) (stdout, stderr string, code int) {
	t.Helper()

	cmd := exec.Command(os.Args[0], args...)
	cmd.Env = append(os.Environ(), "TEMPLET_TEST_SUBPROCESS=1")

	var out, errOut bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errOut

	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if !errors.As(err, &exitErr) {
			t.Fatalf("running the CLI: %v", err)
		}
		code = exitErr.ExitCode()
	}

	return out.String(), errOut.String(), code
}

func TestCLIPrintsVersion(t *testing.T) {
	stdout, stderr, code := runCLI(t, "-version")

	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want it empty", stderr)
	}

	trimmed := strings.TrimSpace(stdout)
	if !strings.HasPrefix(trimmed, "templet ") {
		t.Fatalf("stdout = %q, want the format `templet <version>`", stdout)
	}
	if strings.TrimPrefix(trimmed, "templet ") == "" {
		t.Fatalf("stdout = %q, want a non-empty version", stdout)
	}
}

func TestCLIWritesErrorsToStderr(t *testing.T) {
	cases := []struct {
		name string
		args []string
		want string
	}{
		{name: "no file", args: nil, want: "no file specified"},
		{
			name: "missing file",
			args: []string{filepath.Join(t.TempDir(), "nope.toml")},
			want: "couldn't be opened",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdout, stderr, code := runCLI(t, tc.args...)

			if stdout != "" {
				t.Errorf("stdout = %q, want it empty: errors belong on stderr", stdout)
			}
			if !strings.Contains(stderr, tc.want) {
				t.Errorf("stderr = %q, want it to contain %q", stderr, tc.want)
			}
			if code != 1 {
				t.Errorf("exit code = %d, want 1", code)
			}
		})
	}
}

func TestCLIWritesReportsToStdout(t *testing.T) {
	dir := t.TempDir()
	tplPath := filepath.Join(dir, "template.toml")
	content := "variables = [\"Name\"]\n\n[[files]]\npath = \"out.txt\"\ncontent = \"hi {{ .Name }}\"\n"
	if err := os.WriteFile(tplPath, []byte(content), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}

	stdout, stderr, code := runCLI(t, "-dest", dir, "-var", "Name=there", tplPath)
	if code != 0 {
		t.Fatalf("exit code = %d, want 0 (stderr = %q)", code, stderr)
	}
	if stderr != "" {
		t.Errorf("stderr = %q, want it empty on success", stderr)
	}
	if !strings.Contains(stdout, "Generate file `out.txt`") {
		t.Errorf("stdout = %q, want the generation report", stdout)
	}

	got, err := os.ReadFile(filepath.Join(dir, "out.txt"))
	if err != nil {
		t.Fatalf("reading the generated file: %v", err)
	}
	if want := "hi there"; string(got) != want {
		t.Errorf("generated content = %q, want %q", got, want)
	}
}

func collectWithFlags(t *testing.T, flags console.SliceFlag) (map[string]any, error) {
	t.Helper()

	return collectVariables(flags, nil, func(string) string { return "unused" })
}

func TestCollectVariablesKeepsAnsweredValueVerbatim(t *testing.T) {
	got, err := collectVariables(nil, []string{"Name"}, func(string) string {
		return "Mi Proyecto"
	})
	if err != nil {
		t.Fatalf("collectVariables() error = %v", err)
	}

	if want := "Mi Proyecto"; got["Name"] != want {
		t.Fatalf("Name = %q, want %q (the answer must not be slugified)", got["Name"], want)
	}
}

func TestCollectVariablesKeepsInterpolatableValuesIntact(t *testing.T) {
	// Go module paths are case sensitive, so lowercasing one breaks the import.
	const module = "github.com/wfrs-dev/Templet"

	got, err := collectVariables(nil, []string{"ModuleName"}, func(string) string {
		return module
	})
	if err != nil {
		t.Fatalf("collectVariables() error = %v", err)
	}

	if got["ModuleName"] != module {
		t.Fatalf("ModuleName = %q, want %q (slugs drop dots and collapse spaces)", got["ModuleName"], module)
	}
}

func TestCollectVariablesPrefersFlagsOverAsking(t *testing.T) {
	asked := false

	got, err := collectVariables(
		console.SliceFlag{"Name=from-flag"},
		[]string{"Name"},
		func(string) string {
			asked = true
			return "from-prompt"
		},
	)
	if err != nil {
		t.Fatalf("collectVariables() error = %v", err)
	}

	if asked {
		t.Error("the prompt must not run for a variable already given with -var")
	}
	if want := "from-flag"; got["Name"] != want {
		t.Fatalf("Name = %q, want %q", got["Name"], want)
	}
}

func TestCollectVariablesAcceptsValuesContainingEquals(t *testing.T) {
	got, err := collectWithFlags(t, console.SliceFlag{
		"URL=http://host/path?a=b&c=d",
		"Query=a=b=c",
	})
	if err != nil {
		t.Fatalf("collectVariables() error = %v, want nil", err)
	}

	if want := "http://host/path?a=b&c=d"; got["URL"] != want {
		t.Errorf("URL = %v, want %q", got["URL"], want)
	}
	if want := "a=b=c"; got["Query"] != want {
		t.Errorf("Query = %v, want %q", got["Query"], want)
	}
}

func TestCollectVariablesKeepsEmptyValue(t *testing.T) {
	got, err := collectWithFlags(t, console.SliceFlag{"Blank="})
	if err != nil {
		t.Fatalf("collectVariables() error = %v, want nil", err)
	}

	value, ok := got["Blank"]
	if !ok {
		t.Fatal("Blank is missing, want an empty value")
	}
	if value != "" {
		t.Fatalf("Blank = %v, want an empty value", value)
	}
}

func TestCollectVariablesRejectsVariableWithoutSeparator(t *testing.T) {
	_, err := collectWithFlags(t, console.SliceFlag{"Name"})
	if err == nil {
		t.Fatal("error = nil, want an error for a variable without key=value")
	}
}

func TestCollectVariablesRejectsEmptyKey(t *testing.T) {
	_, err := collectWithFlags(t, console.SliceFlag{"=value"})
	if err == nil {
		t.Fatal("error = nil, want an error for an empty key")
	}
}
