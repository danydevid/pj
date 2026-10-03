package config

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// ============================================================
// GlobalConfig
// ============================================================

func TestGlobalConfig_ApplyDefaults(t *testing.T) {
	t.Setenv("USER", "alice")
	t.Setenv("EDITOR", "vim")
	t.Setenv("SHELL", "/bin/zsh")

	c := &GlobalConfig{}
	c.ApplyDefaults()

	if c.PJSpecVersion != ExpectedSpecVersion {
		t.Errorf("PJSpecVersion = %q, want %q", c.PJSpecVersion, ExpectedSpecVersion)
	}
	if c.User.Name != "alice" {
		t.Errorf("User.Name = %q, want %q", c.User.Name, "alice")
	}
	if c.Defaults.Editor != "vim" {
		t.Errorf("Defaults.Editor = %q, want %q", c.Defaults.Editor, "vim")
	}
	if c.Defaults.Shell != "/bin/zsh" {
		t.Errorf("Defaults.Shell = %q, want %q", c.Defaults.Shell, "/bin/zsh")
	}
	if c.Behavior.AutoCD == nil || !*c.Behavior.AutoCD {
		t.Error("Behavior.AutoCD should default to true")
	}
	if c.Behavior.AutoActivate == nil || *c.Behavior.AutoActivate {
		t.Error("Behavior.AutoActivate should default to false")
	}
	if c.Behavior.TrackUsage == nil || !*c.Behavior.TrackUsage {
		t.Error("Behavior.TrackUsage should default to true")
	}
	if c.Logging.Level != "info" {
		t.Errorf("Logging.Level = %q, want info", c.Logging.Level)
	}
	if c.Logging.Format != "text" {
		t.Errorf("Logging.Format = %q, want text", c.Logging.Format)
	}
}

func TestGlobalConfig_ApplyDefaults_PreservesExplicitValues(t *testing.T) {
	falseVal := false
	c := &GlobalConfig{
		PJSpecVersion: "1",
		User:          UserConfig{Name: "bob"},
		Behavior:      BehaviorConfig{AutoCD: &falseVal},
		Logging:       LoggingConfig{Level: "debug", Format: "json"},
	}
	c.ApplyDefaults()

	if c.User.Name != "bob" {
		t.Errorf("User.Name = %q, want bob", c.User.Name)
	}
	if c.Behavior.AutoCD == nil || *c.Behavior.AutoCD {
		t.Error("Behavior.AutoCD should remain false")
	}
	if c.Logging.Level != "debug" {
		t.Errorf("Logging.Level = %q, want debug", c.Logging.Level)
	}
	if c.Logging.Format != "json" {
		t.Errorf("Logging.Format = %q, want json", c.Logging.Format)
	}
}

func TestGlobalConfig_Validate(t *testing.T) {
	tests := []struct {
		name    string
		cfg     GlobalConfig
		wantErr bool
	}{
		{
			name: "valid",
			cfg: GlobalConfig{
				PJSpecVersion: "1",
				Logging:       LoggingConfig{Level: "info", Format: "text"},
			},
		},
		{
			name: "wrong spec version",
			cfg: GlobalConfig{
				PJSpecVersion: "99",
				Logging:       LoggingConfig{Level: "info", Format: "text"},
			},
			wantErr: true,
		},
		{
			name: "invalid log level",
			cfg: GlobalConfig{
				PJSpecVersion: "1",
				Logging:       LoggingConfig{Level: "verbose", Format: "text"},
			},
			wantErr: true,
		},
		{
			name: "invalid log format",
			cfg: GlobalConfig{
				PJSpecVersion: "1",
				Logging:       LoggingConfig{Level: "info", Format: "xml"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if err != nil && !errors.Is(err, ErrConfigInvalid) {
				t.Errorf("expected error to wrap ErrConfigInvalid, got %v", err)
			}
		})
	}
}

// ============================================================
// ProjectsRegistry
// ============================================================

func TestProjectsRegistry_Validate(t *testing.T) {
	tests := []struct {
		name    string
		reg     ProjectsRegistry
		wantErr bool
	}{
		{
			name: "valid",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"foo": {Path: "/home/user/foo", Status: "active"},
				},
			},
		},
		{
			name: "empty status defaults to active",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"foo": {Path: "/home/user/foo"},
				},
			},
		},
		{
			name: "invalid spec version",
			reg: ProjectsRegistry{
				PJSpecVersion: "0",
				Projects:      map[string]ProjectItem{},
			},
			wantErr: true,
		},
		{
			name: "invalid project name (space)",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"has space": {Path: "/home/user/foo"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid project name (too long)",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa": {
						Path: "/home/user/foo",
					},
				},
			},
			wantErr: true,
		},
		{
			name: "non-absolute path",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"foo": {Path: "relative/path"},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid status",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"foo": {Path: "/home/user/foo", Status: "deleted"},
				},
			},
			wantErr: true,
		},
		{
			name: "archived status is valid",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"foo": {Path: "/home/user/foo", Status: "archived"},
				},
			},
		},
		{
			name: "template status is valid",
			reg: ProjectsRegistry{
				PJSpecVersion: "1",
				Projects: map[string]ProjectItem{
					"foo": {Path: "/home/user/foo", Status: "template"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.reg.Validate()
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

// ============================================================
// PJRC (in-memory)
// ============================================================

func TestPJRC_ApplyDefaults(t *testing.T) {
	p := &PJRC{}
	p.ApplyDefaults()
	if p.PJSpecVersion != ExpectedSpecVersion {
		t.Errorf("PJSpecVersion = %q, want %q", p.PJSpecVersion, ExpectedSpecVersion)
	}

	p2 := &PJRC{PJSpecVersion: "1"}
	p2.ApplyDefaults()
	if p2.PJSpecVersion != "1" {
		t.Errorf("ApplyDefaults overrode explicit spec version: %q", p2.PJSpecVersion)
	}
}

func TestPJRC_Validate(t *testing.T) {
	if err := (&PJRC{PJSpecVersion: "1"}).Validate(); err != nil {
		t.Errorf("valid PJRC failed validation: %v", err)
	}
	if err := (&PJRC{PJSpecVersion: "2"}).Validate(); err == nil {
		t.Error("expected error for unsupported spec version")
	}
}

// ============================================================
// LoadPJRC / LoadPJRCFile
// ============================================================

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func TestLoadPJRC_HappyPath(t *testing.T) {
	tmp := t.TempDir()
	writeFile(t, filepath.Join(tmp, PJRCFilename), `
pj_spec_version = "1"

[project]
name = "test"
`)

	p, err := LoadPJRC(tmp)
	if err != nil {
		t.Fatalf("LoadPJRC: %v", err)
	}
	if p.Project.Name != "test" {
		t.Errorf("Project.Name = %q, want test", p.Project.Name)
	}
}

func TestLoadPJRC_RelativePathRejected(t *testing.T) {
	_, err := LoadPJRC("relative/path")
	if err == nil {
		t.Fatal("expected error for relative path")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Errorf("expected ErrConfigInvalid, got %v", err)
	}
}

func TestLoadPJRC_NotADirectory(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "file")
	writeFile(t, f, "x")

	if _, err := LoadPJRC(f); err == nil {
		t.Fatal("expected error for non-directory project path")
	}
}

func TestLoadPJRC_MissingFile(t *testing.T) {
	tmp := t.TempDir()
	_, err := LoadPJRC(tmp)
	if err == nil {
		t.Fatal("expected error for missing .pjrc")
	}
	if !errors.Is(err, ErrPJRCNotFound) {
		t.Errorf("expected ErrPJRCNotFound, got %v", err)
	}
}

func TestLoadPJRCFile_FullContent(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, PJRCFilename)
	writeFile(t, path, `
pj_spec_version = "1"

[project]
name = "foo"
type = "go-cli"
description = "Foo CLI tool"

[env]
GOFLAGS = "-mod=vendor"
EDITOR = "nvim"

[tasks]
build = "go build ./..."
test = "go test ./..."

[hooks]
after_jump = "git status --short"

[shell]
init = "source ./venv/bin/activate"

[template]
source = "go-cli"
version = "1.0.0"
`)

	p, err := LoadPJRCFile(path)
	if err != nil {
		t.Fatalf("LoadPJRCFile: %v", err)
	}
	if p.Project.Name != "foo" {
		t.Errorf("Project.Name = %q", p.Project.Name)
	}
	if p.Project.Type != "go-cli" {
		t.Errorf("Project.Type = %q", p.Project.Type)
	}
	if got := p.Env["GOFLAGS"]; got != "-mod=vendor" {
		t.Errorf("Env[GOFLAGS] = %q", got)
	}
	if got := p.Tasks["build"]; got != "go build ./..." {
		t.Errorf("Tasks[build] = %q", got)
	}
	if p.Hooks.AfterJump != "git status --short" {
		t.Errorf("Hooks.AfterJump = %q", p.Hooks.AfterJump)
	}
	if p.Shell.Init != "source ./venv/bin/activate" {
		t.Errorf("Shell.Init = %q", p.Shell.Init)
	}
	if p.Template.Source != "go-cli" {
		t.Errorf("Template.Source = %q", p.Template.Source)
	}
	if p.Template.Version != "1.0.0" {
		t.Errorf("Template.Version = %q", p.Template.Version)
	}
}

func TestLoadPJRCFile_EmptyDefaultsSpecVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, PJRCFilename)
	writeFile(t, path, ``)

	p, err := LoadPJRCFile(path)
	if err != nil {
		t.Fatalf("LoadPJRCFile: %v", err)
	}
	if p.PJSpecVersion != ExpectedSpecVersion {
		t.Errorf("PJSpecVersion = %q, want %q", p.PJSpecVersion, ExpectedSpecVersion)
	}
}

func TestLoadPJRCFile_InvalidTOML(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, PJRCFilename)
	writeFile(t, path, `this is not = valid = toml`)

	_, err := LoadPJRCFile(path)
	if err == nil {
		t.Fatal("expected error for invalid TOML")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Errorf("expected ErrConfigInvalid, got %v", err)
	}
}

func TestLoadPJRCFile_UnsupportedSpecVersion(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, PJRCFilename)
	writeFile(t, path, `pj_spec_version = "99"`)

	_, err := LoadPJRCFile(path)
	if err == nil {
		t.Fatal("expected error for unsupported spec version")
	}
	if !errors.Is(err, ErrConfigInvalid) {
		t.Errorf("expected ErrConfigInvalid, got %v", err)
	}
}

func TestLoadPJRCFile_NotFound(t *testing.T) {
	tmp := t.TempDir()
	_, err := LoadPJRCFile(filepath.Join(tmp, PJRCFilename))
	if !errors.Is(err, ErrPJRCNotFound) {
		t.Errorf("expected ErrPJRCNotFound, got %v", err)
	}
}

// ============================================================
// FindPJRC
// ============================================================

func TestFindPJRC_WalksUp(t *testing.T) {
	tmp := t.TempDir()
	pjrc := filepath.Join(tmp, PJRCFilename)
	writeFile(t, pjrc, `pj_spec_version = "1"`)

	sub := filepath.Join(tmp, "a", "b", "c")
	if err := os.MkdirAll(sub, 0700); err != nil {
		t.Fatal(err)
	}

	found, err := FindPJRC(sub)
	if err != nil {
		t.Fatalf("FindPJRC: %v", err)
	}

	// Resolve symlinks (macOS /var -> /private/var) before comparing.
	wantEval, _ := filepath.EvalSymlinks(pjrc)
	gotEval, _ := filepath.EvalSymlinks(found)
	if gotEval != wantEval {
		t.Errorf("FindPJRC = %q, want %q", gotEval, wantEval)
	}
}

func TestFindPJRC_NotFound(t *testing.T) {
	tmp := t.TempDir()
	_, err := FindPJRC(tmp)
	if err == nil {
		t.Fatal("expected error")
	}
	if !errors.Is(err, ErrPJRCNotFound) {
		t.Errorf("expected ErrPJRCNotFound, got %v", err)
	}
}

// ============================================================
// Paths
// ============================================================

func clearPathEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{
		"PJ_CONFIG_DIR", "PJ_DATA_DIR", "PJ_CACHE_DIR", "PJ_RUNTIME_DIR",
		"XDG_CONFIG_HOME", "XDG_DATA_HOME", "XDG_CACHE_HOME", "XDG_RUNTIME_DIR",
		"XDG_STATE_HOME",
	} {
		t.Setenv(k, "")
	}
}

func TestPaths_XDGDefaults(t *testing.T) {
	clearPathEnv(t)

	home, err := os.UserHomeDir()
	if err != nil {
		t.Skip("no home directory available")
	}

	p := GetPaths()

	if p.ConfigDir != filepath.Join(home, ".config", "pj") {
		t.Errorf("ConfigDir = %q", p.ConfigDir)
	}
	if p.DataDir != filepath.Join(home, ".local", "share", "pj") {
		t.Errorf("DataDir = %q", p.DataDir)
	}
	if p.CacheDir != filepath.Join(home, ".cache", "pj") {
		t.Errorf("CacheDir = %q", p.CacheDir)
	}
	if p.StateDir != filepath.Join(home, ".local", "state", "pj") {
		t.Errorf("StateDir = %q", p.StateDir)
	}
	if p.UserBinDir != filepath.Join(home, ".pj", "bin") {
		t.Errorf("UserBinDir = %q", p.UserBinDir)
	}
}

func TestPaths_XDGOverrides(t *testing.T) {
	clearPathEnv(t)
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")
	t.Setenv("XDG_DATA_HOME", "/custom/data")
	t.Setenv("XDG_CACHE_HOME", "/custom/cache")
	t.Setenv("XDG_RUNTIME_DIR", "/run/user/1000")
	t.Setenv("XDG_STATE_HOME", "/custom/state")

	p := GetPaths()

	if p.ConfigDir != "/custom/config/pj" {
		t.Errorf("ConfigDir = %q", p.ConfigDir)
	}
	if p.DataDir != "/custom/data/pj" {
		t.Errorf("DataDir = %q", p.DataDir)
	}
	if p.CacheDir != "/custom/cache/pj" {
		t.Errorf("CacheDir = %q", p.CacheDir)
	}
	if p.RuntimeDir != "/run/user/1000/pj" {
		t.Errorf("RuntimeDir = %q", p.RuntimeDir)
	}
	if p.StateDir != "/custom/state/pj" {
		t.Errorf("StateDir = %q", p.StateDir)
	}
}

func TestPaths_PJOverridesTakePrecedence(t *testing.T) {
	clearPathEnv(t)
	t.Setenv("XDG_CONFIG_HOME", "/xdg/config")
	t.Setenv("XDG_RUNTIME_DIR", "/xdg/runtime")
	t.Setenv("PJ_CONFIG_DIR", "/pj/config")
	t.Setenv("PJ_DATA_DIR", "/pj/data")
	t.Setenv("PJ_CACHE_DIR", "/pj/cache")
	t.Setenv("PJ_RUNTIME_DIR", "/pj/runtime")

	p := GetPaths()

	if p.ConfigDir != "/pj/config" {
		t.Errorf("ConfigDir = %q, want /pj/config", p.ConfigDir)
	}
	if p.DataDir != "/pj/data" {
		t.Errorf("DataDir = %q, want /pj/data", p.DataDir)
	}
	if p.CacheDir != "/pj/cache" {
		t.Errorf("CacheDir = %q, want /pj/cache", p.CacheDir)
	}
	if p.RuntimeDir != "/pj/runtime" {
		t.Errorf("RuntimeDir = %q, want /pj/runtime", p.RuntimeDir)
	}
}

func TestPaths_RuntimeFallback(t *testing.T) {
	clearPathEnv(t)
	t.Setenv("USER", "testuser")

	p := GetPaths()
	want := filepath.Join(os.TempDir(), "pj-testuser")
	if p.RuntimeDir != want {
		t.Errorf("RuntimeDir = %q, want %q", p.RuntimeDir, want)
	}
}

func TestPaths_FilePaths(t *testing.T) {
	p := &Paths{ConfigDir: "/cfg"}
	if got := p.GlobalConfigFile(); got != "/cfg/config.toml" {
		t.Errorf("GlobalConfigFile = %q", got)
	}
	if got := p.ProjectsConfigFile(); got != "/cfg/projects.toml" {
		t.Errorf("ProjectsConfigFile = %q", got)
	}
	if got := p.ProjectsLockFile(); got != "/cfg/projects.lock" {
		t.Errorf("ProjectsLockFile = %q", got)
	}
}

func TestPaths_EnsureDirs(t *testing.T) {
	tmp := t.TempDir()
	p := &Paths{
		ConfigDir:  filepath.Join(tmp, "config"),
		DataDir:    filepath.Join(tmp, "data"),
		CacheDir:   filepath.Join(tmp, "cache"),
		RuntimeDir: filepath.Join(tmp, "runtime"),
		StateDir:   filepath.Join(tmp, "state"),
		UserBinDir: filepath.Join(tmp, "bin"),
	}

	if err := p.EnsureDirs(); err != nil {
		t.Fatalf("EnsureDirs: %v", err)
	}

	for _, dir := range []string{
		p.ConfigDir, p.DataDir, p.CacheDir, p.RuntimeDir, p.StateDir, p.UserBinDir,
	} {
		info, err := os.Stat(dir)
		if err != nil {
			t.Errorf("expected directory %q to exist: %v", dir, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("%q is not a directory", dir)
		}
		if info.Mode().Perm() != 0700 {
			t.Errorf("%q permissions = %o, want 0700", dir, info.Mode().Perm())
		}
	}
}

// ============================================================
// FileLock
// ============================================================

func TestAcquireLock_Reacquirable(t *testing.T) {
	tmp := t.TempDir()
	lockPath := filepath.Join(tmp, "test.lock")

	lock, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("AcquireLock: %v", err)
	}
	if err := lock.Unlock(); err != nil {
		t.Errorf("Unlock: %v", err)
	}

	// After unlock, the lock must be acquirable again.
	lock2, err := AcquireLock(lockPath)
	if err != nil {
		t.Fatalf("re-acquire: %v", err)
	}
	if err := lock2.Unlock(); err != nil {
		t.Errorf("Unlock (2): %v", err)
	}
}

func TestFileLock_UnlockWithoutFile(t *testing.T) {
	fl := &FileLock{}
	if err := fl.Unlock(); err != nil {
		t.Errorf("Unlock on empty FileLock should be a no-op, got %v", err)
	}
}
