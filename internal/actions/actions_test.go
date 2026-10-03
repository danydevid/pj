package actions

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// =============================================================================
// Shell detection
// =============================================================================

func TestParseShell(t *testing.T) {
	cases := []struct {
		in   string
		want Shell
	}{
		{"bash", Bash},
		{"zsh", Zsh},
		{"sh", Sh},
		{"fish", Fish},
		{"nu", Nu},
		{"pwsh", Pwsh},
		{"elvish", Elvish},
		{"", Sh},
		{"unknown", Sh},
		{"BASH", Bash},
		{"  Fish  ", Fish},
	}
	for _, c := range cases {
		if got := ParseShell(c.in); got != c.want {
			t.Errorf("ParseShell(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestDetectShell(t *testing.T) {
	t.Setenv("PJ_SHELL", "zsh")
	if got := DetectShell(); got != Zsh {
		t.Errorf("DetectShell = %q, want zsh", got)
	}

	t.Setenv("PJ_SHELL", "")
	if got := DetectShell(); got != Sh {
		t.Errorf("DetectShell (empty) = %q, want sh", got)
	}
}

func TestShell_IsPOSIX(t *testing.T) {
	for _, s := range []Shell{Sh, Bash, Zsh} {
		if !s.IsPOSIX() {
			t.Errorf("%q should be POSIX", s)
		}
	}
	for _, s := range []Shell{Fish, Nu, Pwsh, Elvish} {
		if s.IsPOSIX() {
			t.Errorf("%q should not be POSIX", s)
		}
	}
}

// =============================================================================
// Escaping
// =============================================================================

func TestEscapePOSIX(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", `''`},
		{"foo", `'foo'`},
		{"my project", `'my project'`},
		{"it's", `'it'\''s'`},
		{`a'b'c`, `'a'\''b'\''c'`},
		{"/tmp/a b", `'/tmp/a b'`},
	}
	for _, c := range cases {
		if got := EscapePOSIX(c.in); got != c.want {
			t.Errorf("EscapePOSIX(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeFish(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", `''`},
		{"foo", `'foo'`},
		{"my project", `'my project'`},
		{"it's", `'it\'s'`},
		{`a\b`, `'a\\b'`},
		{`it's \`, `'it\'s \\'`},
	}
	for _, c := range cases {
		if got := EscapeFish(c.in); got != c.want {
			t.Errorf("EscapeFish(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapeNu(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", `''`},
		{"foo", `'foo'`},
		{"my project", `'my project'`},
		{"it's", `'it''s'`},
		{`a'b'c`, `'a''b''c'`},
	}
	for _, c := range cases {
		if got := EscapeNu(c.in); got != c.want {
			t.Errorf("EscapeNu(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

func TestEscapePwsh(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", `''`},
		{"foo", `'foo'`},
		{"it's", `'it''s'`},
	}
	for _, c := range cases {
		if got := EscapePwsh(c.in); got != c.want {
			t.Errorf("EscapePwsh(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// =============================================================================
// Per-action rendering
// =============================================================================

func TestActions_CD(t *testing.T) {
	a := New().CD("/tmp/foo")
	cases := map[Shell]string{
		Sh:   `cd '/tmp/foo'`,
		Bash: `cd '/tmp/foo'`,
		Zsh:  `cd '/tmp/foo'`,
		Fish: `cd '/tmp/foo'`,
		Nu:   `cd '/tmp/foo'`,
		Pwsh: `Set-Location '/tmp/foo'`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestActions_CD_QuotedPath(t *testing.T) {
	a := New().CD(`/tmp/it's`)
	got, err := a.Render(Bash)
	if err != nil {
		t.Fatal(err)
	}
	want := `cd '/tmp/it'\''s'` + "\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestActions_SetEnv(t *testing.T) {
	a := New().SetEnv("FOO", "bar")
	cases := map[Shell]string{
		Sh:   `export FOO='bar'`,
		Bash: `export FOO='bar'`,
		Zsh:  `export FOO='bar'`,
		Fish: `set -gx FOO 'bar'`,
		Nu:   `$env.FOO = 'bar'`,
		Pwsh: `$env:FOO = 'bar'`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestActions_UnsetEnv(t *testing.T) {
	a := New().UnsetEnv("FOO")
	cases := map[Shell]string{
		Sh:   `unset FOO`,
		Bash: `unset FOO`,
		Zsh:  `unset FOO`,
		Fish: `set -e FOO`,
		Nu:   `hide FOO`,
		Pwsh: `Remove-Item Env:\FOO -ErrorAction SilentlyContinue`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestActions_Alias(t *testing.T) {
	a := New().Alias("gs", "git status")
	cases := map[Shell]string{
		Sh:   `alias gs='git status'`,
		Bash: `alias gs='git status'`,
		Zsh:  `alias gs='git status'`,
		Fish: `alias gs='git status'`,
		Nu:   `alias gs = 'git status'`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}

	if _, err := a.Render(Pwsh); !errors.Is(err, ErrUnsupported) {
		t.Errorf("Render(pwsh) error = %v, want ErrUnsupported", err)
	}
}

func TestActions_Unalias(t *testing.T) {
	a := New().Unalias("gs")
	cases := map[Shell]string{
		Sh:   `unalias gs`,
		Bash: `unalias gs`,
		Zsh:  `unalias gs`,
		Fish: `functions -e gs`,
		Nu:   `hide gs`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestActions_PrependPath(t *testing.T) {
	a := New().PrependPath("/opt/bin")
	cases := map[Shell]string{
		Sh:   `export PATH='/opt/bin':"$PATH"`,
		Bash: `export PATH='/opt/bin':"$PATH"`,
		Zsh:  `export PATH='/opt/bin':"$PATH"`,
		Fish: `set -gx PATH '/opt/bin' $PATH`,
		Nu:   `$env.PATH = ($env.PATH | prepend '/opt/bin')`,
		Pwsh: `$env:PATH = '/opt/bin' + [IO.Path]::PathSeparator + $env:PATH`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestActions_Source(t *testing.T) {
	a := New().Source("/opt/env.sh")
	cases := map[Shell]string{
		Sh:   `. '/opt/env.sh'`,
		Bash: `source '/opt/env.sh'`,
		Zsh:  `source '/opt/env.sh'`,
		Fish: `source '/opt/env.sh'`,
		Nu:   `source '/opt/env.sh'`,
		Pwsh: `. '/opt/env.sh'`,
	}
	for s, want := range cases {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if strings.TrimSpace(got) != want {
			t.Errorf("Render(%s) = %q, want %q", s, got, want)
		}
	}
}

func TestActions_Prompt(t *testing.T) {
	a := New().Prompt("(pj) \\w> ")
	for _, s := range []Shell{Bash, Zsh} {
		got, err := a.Render(s)
		if err != nil {
			t.Fatalf("Render(%s): %v", s, err)
		}
		if !strings.HasPrefix(got, "PS1=") {
			t.Errorf("Render(%s) = %q, want PS1 prefix", s, got)
		}
	}
	for _, s := range []Shell{Sh, Fish, Nu, Pwsh} {
		if _, err := a.Render(s); !errors.Is(err, ErrUnsupported) {
			t.Errorf("Render(%s) error = %v, want ErrUnsupported", s, err)
		}
	}
}

// =============================================================================
// Multi-action rendering
// =============================================================================

func TestActions_RenderOrderAndTrailingNewline(t *testing.T) {
	a := New().
		CD("/tmp/foo").
		SetEnv("FOO", "bar").
		UnsetEnv("OLD").
		PrependPath("/opt/bin")

	got, err := a.Render(Bash)
	if err != nil {
		t.Fatal(err)
	}
	want := "cd '/tmp/foo'\n" +
		"export FOO='bar'\n" +
		"unset OLD\n" +
		"export PATH='/opt/bin':\"$PATH\"\n"
	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestActions_RenderEmpty(t *testing.T) {
	got, err := New().Render(Bash)
	if err != nil {
		t.Fatal(err)
	}
	if got != "" {
		t.Errorf("empty Actions rendered %q, want empty", got)
	}
}

func TestActions_Len(t *testing.T) {
	a := New()
	if a.Len() != 0 || !a.IsEmpty() {
		t.Fatal("new Actions should be empty")
	}
	a.CD("/tmp").SetEnv("K", "V")
	if a.Len() != 2 || a.IsEmpty() {
		t.Errorf("Len = %d, IsEmpty = %v", a.Len(), a.IsEmpty())
	}
}

// =============================================================================
// Validation
// =============================================================================

func TestActions_Validate(t *testing.T) {
	tests := []struct {
		name    string
		build   func() *Actions
		wantErr error
	}{
		{
			name:  "valid cd",
			build: func() *Actions { return New().CD("/tmp") },
		},
		{
			name:    "relative cd path",
			build:   func() *Actions { return New().CD("relative/path") },
			wantErr: ErrInvalidPath,
		},
		{
			name:    "bad env key starting with digit",
			build:   func() *Actions { return New().SetEnv("1BAD", "x") },
			wantErr: ErrInvalidKey,
		},
		{
			name:    "bad env key with dash",
			build:   func() *Actions { return New().SetEnv("A-B", "x") },
			wantErr: ErrInvalidKey,
		},
		{
			name:    "bad unset key",
			build:   func() *Actions { return New().UnsetEnv("bad key") },
			wantErr: ErrInvalidKey,
		},
		{
			name:    "bad alias name",
			build:   func() *Actions { return New().Alias("has space", "ls") },
			wantErr: ErrInvalidAlias,
		},
		{
			name:    "bad unalias name",
			build:   func() *Actions { return New().Unalias("1bad") },
			wantErr: ErrInvalidAlias,
		},
		{
			name:    "relative prepend path",
			build:   func() *Actions { return New().PrependPath("bin") },
			wantErr: ErrInvalidPath,
		},
		{
			name:    "relative source path",
			build:   func() *Actions { return New().Source("env.sh") },
			wantErr: ErrInvalidPath,
		},
		{
			name: "index reported for second action",
			build: func() *Actions {
				return New().CD("/tmp").SetEnv("1BAD", "x")
			},
			wantErr: ErrInvalidKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.build().Validate()
			if tt.wantErr == nil {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("error = %v, want errors.Is %v", err, tt.wantErr)
			}
			if !strings.Contains(err.Error(), "action[") {
				t.Errorf("error should mention action index: %v", err)
			}
		})
	}
}

// =============================================================================
// Write
// =============================================================================

func TestActions_Write(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions")

	a := New().CD("/tmp/foo").SetEnv("FOO", "bar")
	if err := a.Write(path, Bash); err != nil {
		t.Fatal(err)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := "cd '/tmp/foo'\nexport FOO='bar'\n"
	if string(data) != want {
		t.Errorf("file:\n%s\nwant:\n%s", data, want)
	}

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("perm = %o, want 0600", perm)
	}
}

func TestActions_Write_EmptyPathNoop(t *testing.T) {
	a := New().CD("/tmp/foo")
	if err := a.Write("", Bash); err != nil {
		t.Fatalf("Write with empty path should be a no-op, got %v", err)
	}
}

func TestActions_Write_EmptyActionsNoop(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions")

	if err := New().Write(path, Bash); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("empty Actions should not create the file, stat err = %v", err)
	}
}

func TestActions_Write_ValidationFailsBeforeWriting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions")

	a := New().CD("relative")
	if err := a.Write(path, Bash); !errors.Is(err, ErrInvalidPath) {
		t.Fatalf("error = %v, want ErrInvalidPath", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file must not exist after validation failure, stat err = %v", err)
	}
}

func TestActions_Write_Truncates(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions")

	if err := os.WriteFile(path, []byte("stale content\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := New().CD("/tmp").Write(path, Bash); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "cd '/tmp'\n" {
		t.Errorf("file was not truncated: %q", data)
	}
}

func TestActions_Write_UnsupportedShell(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions")

	a := New().Alias("x", "y")
	if err := a.Write(path, Pwsh); !errors.Is(err, ErrUnsupported) {
		t.Fatalf("error = %v, want ErrUnsupported", err)
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("file must not exist after render failure, stat err = %v", err)
	}
}

func TestActions_WriteEnv(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "actions")
	t.Setenv("PJ_ACTIONS_FILE", path)
	t.Setenv("PJ_SHELL", "fish")

	if err := New().CD("/tmp/foo").WriteEnv(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "cd '/tmp/foo'\n" {
		t.Errorf("file content = %q", data)
	}
}

func TestActions_WriteEnv_UnsetFile(t *testing.T) {
	t.Setenv("PJ_ACTIONS_FILE", "")
	t.Setenv("PJ_SHELL", "bash")

	if err := New().CD("/tmp").WriteEnv(); err != nil {
		t.Fatalf("WriteEnv without PJ_ACTIONS_FILE should be a no-op, got %v", err)
	}
}
