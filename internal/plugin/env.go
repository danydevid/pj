package plugin

import (
	"os"

	"github.com/danydevid/pj/internal/config"
)

// Env holds the PJ_* environment variables exposed to plugins by the
// manager and shell wrapper, with XDG-derived fallbacks for directories.
//
// See PLUGIN_SPEC.md §3 and XDG_SPEC.md §5.
type Env struct {
	Manager       string
	Version       string
	ConfigDir     string
	DataDir       string
	CacheDir      string
	RuntimeDir    string
	ActionsFile   string
	Shell         string
	Project       string
	ProjectPath   string
	ActivatedKeys string // PJ_ACTIVATED_KEYS: space-separated
}

// Load reads PJ_* variables from the process environment.
//
// Directory variables (ConfigDir, DataDir, CacheDir, RuntimeDir) fall back
// to XDG defaults via config.GetPaths — so plugins behave identically
// whether invoked by the manager or run standalone during development.
//
// Optional variables (ActionsFile, Shell, Project, ProjectPath) are left
// empty when unset; callers must check before use.
func Load() *Env {
	p := config.GetPaths()
	return &Env{
		Manager:       os.Getenv("PJ_MANAGER"),
		Version:       os.Getenv("PJ_VERSION"),
		ConfigDir:     p.ConfigDir,
		DataDir:       p.DataDir,
		CacheDir:      p.CacheDir,
		RuntimeDir:    p.RuntimeDir,
		ActionsFile:   os.Getenv("PJ_ACTIONS_FILE"),
		Shell:         os.Getenv("PJ_SHELL"),
		Project:       os.Getenv("PJ_PROJECT"),
		ProjectPath:   os.Getenv("PJ_PROJECT_PATH"),
		ActivatedKeys: os.Getenv("PJ_ACTIVATED_KEYS"),
	}
}

// HasActions reports whether the shell wrapper provided an actions file.
// Hybrid plugins use this to decide whether to emit shell actions.
func (e *Env) HasActions() bool {
	return e.ActionsFile != ""
}

// ShellOrPosix returns PJ_SHELL, defaulting to "sh" (POSIX) when unset.
// See ACTIONS_SPEC.md §4.
func (e *Env) ShellOrPosix() string {
	if e.Shell == "" {
		return "sh"
	}
	return e.Shell
}
