package root

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danydevid/pj/internal/plugin"
)

func TestRoot_FindsPjrcUpward(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".pjrc"), []byte(`pj_spec_version = "1"`), 0600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "a", "b", "c")
	if err := os.MkdirAll(nested, 0700); err != nil {
		t.Fatal(err)
	}

	actionsFile := filepath.Join(t.TempDir(), "actions")
	env := &plugin.Env{ActionsFile: actionsFile, Shell: "bash"}

	var stdout, stderr bytes.Buffer
	rc := run(nil, nested, &stdout, &stderr, env)
	if rc != 0 {
		t.Fatalf("exit=%d stderr=%q", rc, stderr.String())
	}

	data, err := os.ReadFile(actionsFile)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(string(data), "cd ") {
		t.Errorf("actions should start with cd: %q", data)
	}
	// The rendered path must resolve to root.
	want, _ := filepath.EvalSymlinks(root)
	got := strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(string(data)), "cd '"), "'")
	gotResolved, _ := filepath.EvalSymlinks(got)
	if gotResolved != want {
		t.Errorf("cd target = %q, want %q", gotResolved, want)
	}
}

func TestRoot_AlreadyAtRoot(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".pjrc"), []byte(`pj_spec_version = "1"`), 0600); err != nil {
		t.Fatal(err)
	}

	actionsFile := filepath.Join(t.TempDir(), "actions")
	env := &plugin.Env{ActionsFile: actionsFile, Shell: "bash"}

	var stdout, stderr bytes.Buffer
	rc := run(nil, root, &stdout, &stderr, env)
	if rc != 0 {
		t.Fatalf("exit=%d", rc)
	}
	if !strings.Contains(stdout.String(), "Already at project root") {
		t.Errorf("stdout=%q", stdout.String())
	}
	if _, err := os.Stat(actionsFile); !os.IsNotExist(err) {
		t.Errorf("no actions file should be written, stat err=%v", err)
	}
}

func TestRoot_NoPjrc(t *testing.T) {
	isolated := t.TempDir()
	env := &plugin.Env{Shell: "bash"}

	var stdout, stderr bytes.Buffer
	rc := run(nil, isolated, &stdout, &stderr, env)
	if rc != 4 {
		t.Errorf("exit=%d, want 4", rc)
	}
}

func TestRoot_NoActionsFile(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, ".pjrc"), []byte(`pj_spec_version = "1"`), 0600); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "sub")
	os.MkdirAll(nested, 0700)

	env := &plugin.Env{Shell: "bash"} // no actions file
	var stdout, stderr bytes.Buffer
	rc := run(nil, nested, &stdout, &stderr, env)
	if rc != 0 {
		t.Errorf("exit=%d", rc)
	}
	if !strings.Contains(stdout.String(), "Project root:") {
		t.Errorf("stdout=%q", stdout.String())
	}
}
