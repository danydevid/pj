package jump

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

// -----------------------------------------------------------------------------
// helpers
// -----------------------------------------------------------------------------

type harness struct {
	configDir  string
	actionsDir string
}

func newHarness(t *testing.T, projects map[string]config.ProjectItem) *harness {
	t.Helper()
	root := t.TempDir()
	cfgDir := filepath.Join(root, "config")
	actDir := filepath.Join(root, "actions")
	if err := os.MkdirAll(cfgDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(actDir, 0700); err != nil {
		t.Fatal(err)
	}

	reg := config.ProjectsRegistry{
		PJSpecVersion: config.ExpectedSpecVersion,
		Projects:      projects,
	}
	var buf bytes.Buffer
	if err := toml.NewEncoder(&buf).Encode(reg); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, "projects.toml"), buf.Bytes(), 0600); err != nil {
		t.Fatal(err)
	}
	return &harness{configDir: cfgDir, actionsDir: actDir}
}

// env returns a plugin.Env with an actions file present (wrapper case).
func (h *harness) env(shell string) *plugin.Env {
	return &plugin.Env{
		ConfigDir:   h.configDir,
		ActionsFile: filepath.Join(h.actionsDir, "actions"),
		Shell:       shell,
	}
}

// envNoActions returns a plugin.Env without an actions file (direct case).
func (h *harness) envNoActions() *plugin.Env {
	return &plugin.Env{ConfigDir: h.configDir}
}

// -----------------------------------------------------------------------------
// happy path
// -----------------------------------------------------------------------------

func TestRun_Success(t *testing.T) {
	target := t.TempDir()
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: target, Status: "active"},
	})

	var stdout, stderr bytes.Buffer
	rc := run([]string{"foo"}, &stdout, &stderr, h.env("bash"))
	if rc != 0 {
		t.Fatalf("exit = %d, stderr = %q", rc, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Jumping to foo") {
		t.Errorf("stdout = %q", stdout.String())
	}

	data, err := os.ReadFile(h.env("bash").ActionsFile)
	if err != nil {
		t.Fatalf("actions file: %v", err)
	}
	// We only assert the shape here; exact escaping is covered below.
	if !strings.HasPrefix(string(data), "cd ") || !strings.HasSuffix(string(data), "\n") {
		t.Errorf("actions = %q", data)
	}
}

// -----------------------------------------------------------------------------
// missing actions file — direct invocation
// -----------------------------------------------------------------------------

func TestRun_NoActionsFile(t *testing.T) {
	target := t.TempDir()
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: target, Status: "active"},
	})

	var stdout, stderr bytes.Buffer
	rc := run([]string{"foo"}, &stdout, &stderr, h.envNoActions())
	if rc != 0 {
		t.Fatalf("exit = %d, stderr = %q", rc, stderr.String())
	}
	if !strings.Contains(stdout.String(), "Jumping to foo") {
		t.Errorf("stdout = %q", stdout.String())
	}
	// No file should have been created anywhere.
	entries, err := os.ReadDir(h.actionsDir)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 0 {
		t.Errorf("actions dir should be empty, has %d entries", len(entries))
	}
}

// -----------------------------------------------------------------------------
// error paths
// -----------------------------------------------------------------------------

func TestRun_UsageError(t *testing.T) {
	h := newHarness(t, nil)
	for _, args := range [][]string{nil, {}, {"a", "b"}} {
		var out, errBuf bytes.Buffer
		rc := run(args, &out, &errBuf, h.env("bash"))
		if rc != exitUsage {
			t.Errorf("args=%v exit=%d, want %d", args, rc, exitUsage)
		}
		if !strings.Contains(errBuf.String(), "usage:") {
			t.Errorf("args=%v stderr=%q", args, errBuf.String())
		}
	}
}

func TestRun_NotFound(t *testing.T) {
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: t.TempDir(), Status: "active"},
	})
	var out, errBuf bytes.Buffer
	rc := run([]string{"bar"}, &out, &errBuf, h.env("bash"))
	if rc != exitNotFound {
		t.Errorf("exit = %d, want %d", rc, exitNotFound)
	}
	if !strings.Contains(errBuf.String(), "not found") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRun_Archived(t *testing.T) {
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: t.TempDir(), Status: "archived"},
	})
	var out, errBuf bytes.Buffer
	rc := run([]string{"foo"}, &out, &errBuf, h.env("bash"))
	if rc != exitNotFound {
		t.Errorf("exit = %d, want %d", rc, exitNotFound)
	}
	if !strings.Contains(errBuf.String(), "archived") {
		t.Errorf("stderr = %q", errBuf.String())
	}
}

func TestRun_NotADirectory(t *testing.T) {
	// A regular file masquerading as a project path.
	tmp := t.TempDir()
	file := filepath.Join(tmp, "file")
	if err := os.WriteFile(file, []byte("x"), 0600); err != nil {
		t.Fatal(err)
	}
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: file, Status: "active"},
	})
	var out, errBuf bytes.Buffer
	rc := run([]string{"foo"}, &out, &errBuf, h.env("bash"))
	if rc != exitValidation {
		t.Errorf("exit = %d, want %d (stderr=%q)", rc, exitValidation, errBuf.String())
	}
}

func TestRun_PathDoesNotExist(t *testing.T) {
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: "/nonexistent/pj-test-path", Status: "active"},
	})
	var out, errBuf bytes.Buffer
	rc := run([]string{"foo"}, &out, &errBuf, h.env("bash"))
	if rc != exitAccess {
		t.Errorf("exit = %d, want %d (stderr=%q)", rc, exitAccess, errBuf.String())
	}
}

// -----------------------------------------------------------------------------
// escaping — the real test
//
// For each tricky directory name, we:
//   1. run the plugin to produce an actions file,
//   2. source that file inside a real shell,
//   3. assert $PWD matches the intended directory.
//
// If any escape were wrong, the shell would land somewhere else, error
// out, or execute injected content.
// -----------------------------------------------------------------------------

func TestRun_Escaping_Bash(t *testing.T) {
	if _, err := exec.LookPath("bash"); err != nil {
		t.Skip("bash not available")
	}
	runEscapingSuite(t, "bash")
}

func TestRun_Escaping_Zsh(t *testing.T) {
	if _, err := exec.LookPath("zsh"); err != nil {
		t.Skip("zsh not available")
	}
	runEscapingSuite(t, "zsh")
}

func TestRun_Escaping_Fish(t *testing.T) {
	if _, err := exec.LookPath("fish"); err != nil {
		t.Skip("fish not available")
	}
	runEscapingSuite(t, "fish")
}

func runEscapingSuite(t *testing.T, shell string) {
	t.Helper()

	cases := []struct {
		label string
		dir   string
	}{
		{"plain", "plain"},
		{"spaces", "my project"},
		{"single quote", "it's"},
		{"double quote", `say "hi"`},
		{"dollar", "cost$5"},
		{"backtick", "run`ls`"},
		{"semicolon", "a;echo pwned"},
		{"ampersand", "a && echo pwned"},
		{"backslash", `back\slash`},
		{"leading dash in segment", "sub/-weird"},
	}

	for _, c := range cases {
		t.Run(shell+"/"+c.label, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, c.dir)
			if err := os.MkdirAll(target, 0700); err != nil {
				t.Fatalf("mkdir %q: %v", target, err)
			}

			h := newHarness(t, map[string]config.ProjectItem{
				"foo": {Path: target, Status: "active"},
			})
			env := h.env(shell)

			var stdout, stderr bytes.Buffer
			rc := run([]string{"foo"}, &stdout, &stderr, env)
			if rc != 0 {
				t.Fatalf("exit=%d stderr=%q", rc, stderr.String())
			}

			// Confirm no injection: the marker file must not exist.
			pwned := filepath.Join(root, "pwned")
			if _, err := os.Stat(pwned); err == nil {
				t.Fatalf("path escaped into shell: %s created", pwned)
			}

			got := sourceAndPwd(t, shell, env.ActionsFile)
			want := resolve(t, target)
			if got != want {
				data, _ := os.ReadFile(env.ActionsFile)
				t.Errorf("pwd = %q, want %q\nrendered actions:\n%s",
					got, want, data)
			}
		})
	}
}

// sourceAndPwd sources actionsFile in a fresh shell and prints $PWD.
// The shell is given the path as a positional argument so we never have
// to quote it ourselves inside the test — the shell's own argument
// handling is trusted here.
func sourceAndPwd(t *testing.T, shell, actionsFile string) string {
	t.Helper()

	var script string
	var args []string
	switch shell {
	case "bash", "zsh":
		// $1 is the actions file path.
		script = `. "$1" && pwd -P`
		args = []string{"-c", script, shell, actionsFile}
	case "fish":
		// $argv[1] is the actions file path.
		script = `source $argv[1]; and pwd -P`
		args = []string{"-c", script, actionsFile}
	default:
		t.Fatalf("unsupported shell %q", shell)
	}

	cmd := exec.Command(shell, args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s failed: %v\noutput:\n%s", shell, err, out)
	}
	return strings.TrimSpace(string(out))
}

// resolve returns the canonical (symlink-free) form of path.
// On macOS, t.TempDir() lives under /var which is a symlink to /private/var.
func resolve(t *testing.T, path string) string {
	t.Helper()
	r, err := filepath.EvalSymlinks(path)
	if err != nil {
		t.Fatalf("EvalSymlinks(%q): %v", path, err)
	}
	return r
}

// -----------------------------------------------------------------------------
// explicit action-file shape checks
//
// The shell tests above prove correctness in practice. These pin the
// exact bytes so a future refactor cannot silently change the format.
// -----------------------------------------------------------------------------

func TestRun_ActionsFileShape(t *testing.T) {
	target := t.TempDir()
	h := newHarness(t, map[string]config.ProjectItem{
		"foo": {Path: target, Status: "active"},
	})

	var out, errBuf bytes.Buffer
	if rc := run([]string{"foo"}, &out, &errBuf, h.env("bash")); rc != 0 {
		t.Fatalf("exit=%d", rc)
	}

	data, err := os.ReadFile(h.env("bash").ActionsFile)
	if err != nil {
		t.Fatal(err)
	}
	want := "cd " + "'" + target + "'" + "\n"
	if string(data) != want {
		t.Errorf("actions = %q\nwant     %q", data, want)
	}
}
