package deactivate

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/danydevid/pj/internal/plugin"
)

func writeActionsPath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "actions")
}

func TestDeactivate_UnsertsKeys(t *testing.T) {
	actionsFile := writeActionsPath(t)
	env := &plugin.Env{
		ActionsFile:   actionsFile,
		Shell:         "bash",
		Project:       "foo",
		ActivatedKeys: "GOFLAGS EDITOR",
	}

	var stdout, stderr bytes.Buffer
	rc := run(nil, &stdout, &stderr, env)
	if rc != 0 {
		t.Fatalf("exit=%d stderr=%q", rc, stderr.String())
	}

	data, err := os.ReadFile(actionsFile)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{
		"unset GOFLAGS\n",
		"unset EDITOR\n",
		"unset PJ_PROJECT\n",
		"unset PJ_PROJECT_PATH\n",
		"unset PJ_ACTIVATED_KEYS\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in:\n%s", want, got)
		}
	}
}

func TestDeactivate_NoActiveProject(t *testing.T) {
	actionsFile := writeActionsPath(t)
	env := &plugin.Env{
		ActionsFile: actionsFile,
		Shell:       "bash",
		// Project empty => no active project
	}

	var stdout, stderr bytes.Buffer
	rc := run(nil, &stdout, &stderr, env)
	if rc != 0 {
		t.Errorf("exit=%d", rc)
	}
	if !strings.Contains(stdout.String(), "No active project") {
		t.Errorf("stdout=%q", stdout.String())
	}
	if _, err := os.Stat(actionsFile); !os.IsNotExist(err) {
		t.Errorf("no actions file should be written, stat err=%v", err)
	}
}

func TestDeactivate_MalformedKeysFiltered(t *testing.T) {
	actionsFile := writeActionsPath(t)
	env := &plugin.Env{
		ActionsFile:   actionsFile,
		Shell:         "bash",
		Project:       "foo",
		ActivatedKeys: "GOOD 1BAD bad-key GOOD2",
	}

	var stdout, stderr bytes.Buffer
	run(nil, &stdout, &stderr, env)

	data, _ := os.ReadFile(actionsFile)
	got := string(data)
	if strings.Contains(got, "unset 1BAD") || strings.Contains(got, "unset bad-key") {
		t.Errorf("invalid keys must not be emitted:\n%s", got)
	}
	if !strings.Contains(got, "unset GOOD\n") || !strings.Contains(got, "unset GOOD2\n") {
		t.Errorf("valid keys must be emitted:\n%s", got)
	}
}

func TestDeactivate_UsageError(t *testing.T) {
	env := &plugin.Env{Shell: "bash", Project: "foo"}
	var stdout, stderr bytes.Buffer
	rc := run([]string{"extra"}, &stdout, &stderr, env)
	if rc != 2 {
		t.Errorf("exit=%d, want 2", rc)
	}
}
