package activate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

type harness struct {
	projectPath string
	configDir   string
	actionsFile string
}

func newHarness(t *testing.T, name, pjrc string) *harness {
	t.Helper()
	root := t.TempDir()

	projectPath := filepath.Join(root, "proj")
	if err := os.MkdirAll(projectPath, 0700); err != nil {
		t.Fatal(err)
	}
	if pjrc != "" {
		if err := os.WriteFile(filepath.Join(projectPath, ".pjrc"), []byte(pjrc), 0600); err != nil {
			t.Fatal(err)
		}
	}

	configDir := filepath.Join(root, "config")
	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatal(err)
	}

	if name != "" {
		reg := config.ProjectsRegistry{
			PJSpecVersion: config.ExpectedSpecVersion,
			Projects: map[string]config.ProjectItem{
				name: {Path: projectPath, Status: "active"},
			},
		}
		var buf bytes.Buffer
		if err := toml.NewEncoder(&buf).Encode(reg); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(configDir, "projects.toml"), buf.Bytes(), 0600); err != nil {
			t.Fatal(err)
		}
	}

	return &harness{
		projectPath: projectPath,
		configDir:   configDir,
		actionsFile: filepath.Join(root, "actions"),
	}
}

func (h *harness) env() *plugin.Env {
	return &plugin.Env{
		ConfigDir:   h.configDir,
		ActionsFile: h.actionsFile,
		Shell:       "bash",
	}
}

func readActions(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read actions: %v", err)
	}
	return string(data)
}

func TestActivate_ByName_SetsEnvAndCD(t *testing.T) {
	h := newHarness(t, "foo", `
pj_spec_version = "1"

[project]
name = "foo"

[env]
GOFLAGS = "-mod=vendor"
EDITOR = "nvim"
`)
	var stdout, stderr bytes.Buffer
	rc := run([]string{"foo"}, "/", &stdout, &stderr, h.env())
	if rc != 0 {
		t.Fatalf("exit=%d stderr=%q", rc, stderr.String())
	}

	got := readActions(t, h.actionsFile)
	for _, want := range []string{
		"cd '" + h.projectPath + "'\n",
		"export GOFLAGS='-mod=vendor'\n",
		"export EDITOR='nvim'\n",
		"export PJ_PROJECT='foo'\n",
		"export PJ_PROJECT_PATH='" + h.projectPath + "'\n",
		"export PJ_ACTIVATED_KEYS='EDITOR GOFLAGS'\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("actions missing %q\ngot:\n%s", want, got)
		}
	}
}

func TestActivate_InPlace_NoCD(t *testing.T) {
	h := newHarness(t, "", `
pj_spec_version = "1"
[env]
FOO = "bar"
`)
	var stdout, stderr bytes.Buffer
	rc := run(nil, h.projectPath, &stdout, &stderr, h.env())
	if rc != 0 {
		t.Fatalf("exit=%d stderr=%q", rc, stderr.String())
	}
	got := readActions(t, h.actionsFile)
	if strings.Contains(got, "cd ") {
		t.Errorf("in-place activation must not emit cd:\n%s", got)
	}
	if !strings.Contains(got, "export FOO='bar'") {
		t.Errorf("actions missing FOO export:\n%s", got)
	}
}

func TestActivate_Substitution(t *testing.T) {
	h := newHarness(t, "foo", `
pj_spec_version = "1"
[env]
ROOT = "${PJ_PROJECT_PATH}"
NAME = "${PJ_PROJECT_NAME}"
`)
	var stdout, stderr bytes.Buffer
	run([]string{"foo"}, "/", &stdout, &stderr, h.env())

	got := readActions(t, h.actionsFile)
	if !strings.Contains(got, "export ROOT='"+h.projectPath+"'") {
		t.Errorf("${PJ_PROJECT_PATH} not substituted:\n%s", got)
	}
	if !strings.Contains(got, "export NAME='foo'") {
		t.Errorf("${PJ_PROJECT_NAME} not substituted:\n%s", got)
	}
}

func TestActivate_Reactivation_UnsetsPreviousKeys(t *testing.T) {
	h := newHarness(t, "foo", `
pj_spec_version = "1"
[env]
NEW = "x"
`)
	env := h.env()
	env.ActivatedKeys = "OLD1 OLD2"

	var stdout, stderr bytes.Buffer
	run([]string{"foo"}, "/", &stdout, &stderr, env)

	got := readActions(t, h.actionsFile)
	for _, want := range []string{"unset OLD1\n", "unset OLD2\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestActivate_NoActionsFile(t *testing.T) {
	h := newHarness(t, "foo", `pj_spec_version = "1"`)
	env := h.env()
	env.ActionsFile = ""

	var stdout, stderr bytes.Buffer
	rc := run([]string{"foo"}, "/", &stdout, &stderr, env)
	if rc != 0 {
		t.Fatalf("exit=%d", rc)
	}
	if !strings.Contains(stdout.String(), "Activating foo") {
		t.Errorf("stdout = %q", stdout.String())
	}
	if _, err := os.Stat(h.actionsFile); !os.IsNotExist(err) {
		t.Errorf("actions file should not exist, stat err = %v", err)
	}
}

func TestActivate_NotFound(t *testing.T) {
	h := newHarness(t, "foo", `pj_spec_version = "1"`)
	var stdout, stderr bytes.Buffer
	rc := run([]string{"bar"}, "/", &stdout, &stderr, h.env())
	if rc != 4 {
		t.Errorf("exit=%d, want 4", rc)
	}
}

func TestActivate_InPlace_NoPjrcAnywhere(t *testing.T) {
	h := newHarness(t, "", "")
	// Point cwd at an isolated empty directory tree.
	isolated := t.TempDir()
	var stdout, stderr bytes.Buffer
	rc := run(nil, isolated, &stdout, &stderr, h.env())
	if rc != 4 {
		t.Errorf("exit=%d, want 4 (stderr=%q)", rc, stderr.String())
	}
}
