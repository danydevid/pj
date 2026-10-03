// Package activate implements `pj activate`.
//
// Resolves a project (by name or via the nearest .pjrc from cwd), reads
// its [env] section, and emits shell actions to:
//   - unset any env vars from a previous activation,
//   - cd into the project (when activated by name),
//   - set the .pjrc env vars with variable substitution,
//   - record PJ_PROJECT, PJ_PROJECT_PATH and PJ_ACTIVATED_KEYS.
//
// See docs/CONFIG_SPEC.md §4 for the .pjrc contract.
package activate

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/danydevid/pj/internal/actions"
	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

const (
	exitOK         = 0
	exitGeneral    = 1
	exitUsage      = 2
	exitConfig     = 3
	exitNotFound   = 4
	exitAccess     = 5
	exitValidation = 6
)

func Run(args []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pj-activate: %v\n", err)
		return exitGeneral
	}
	return run(args, cwd, os.Stdout, os.Stderr, plugin.Load())
}

func run(args []string, cwd string, stdout, stderr io.Writer, env *plugin.Env) int {
	if len(args) > 1 {
		fmt.Fprintln(stderr, "usage: pj activate [project]")
		return exitUsage
	}

	var (
		projectPath string
		projectName string
		cdHere      bool
	)

	if len(args) == 1 {
		name := args[0]
		reg, err := config.LoadProjectsRegistry(env.ConfigDir)
		if err != nil {
			fmt.Fprintf(stderr, "pj-activate: %v\n", err)
			return exitConfig
		}
		item, ok := reg.Projects[name]
		if !ok {
			fmt.Fprintf(stderr, "pj-activate: project '%s' not found\n", name)
			return exitNotFound
		}
		if item.Status == "archived" {
			fmt.Fprintf(stderr, "pj-activate: project '%s' is archived\n", name)
			return exitNotFound
		}
		if !filepath.IsAbs(item.Path) {
			fmt.Fprintf(stderr, "pj-activate: non-absolute path: %s\n", item.Path)
			return exitConfig
		}
		projectPath = item.Path
		projectName = name
		cdHere = true
	} else {
		// In-place: find nearest .pjrc upward from cwd.
		pjrcPath, err := config.FindPJRC(cwd)
		if err != nil {
			fmt.Fprintf(stderr, "pj-activate: no project found from %s\n", cwd)
			return exitNotFound
		}
		projectPath = filepath.Dir(pjrcPath)
		projectName = filepath.Base(projectPath)
	}

	info, err := os.Stat(projectPath)
	if err != nil {
		fmt.Fprintf(stderr, "pj-activate: %v\n", err)
		return exitAccess
	}
	if !info.IsDir() {
		fmt.Fprintf(stderr, "pj-activate: not a directory: %s\n", projectPath)
		return exitValidation
	}

	// Load .pjrc if present. Absence is fine: we still register the
	// project markers so `pj deactivate` and future tools see it.
	pjrcPath := filepath.Join(projectPath, config.PJRCFilename)
	pjrc := &config.PJRC{}
	if _, err := os.Stat(pjrcPath); err == nil {
		loaded, err := config.LoadPJRCFile(pjrcPath)
		if err != nil {
			fmt.Fprintf(stderr, "pj-activate: %v\n", err)
			return exitConfig
		}
		pjrc = loaded
	} else if !os.IsNotExist(err) {
		fmt.Fprintf(stderr, "pj-activate: %v\n", err)
		return exitAccess
	}

	// Sorted keys for deterministic output.
	keys := make([]string, 0, len(pjrc.Env))
	for k := range pjrc.Env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	fmt.Fprintf(stdout, "Activating %s (%s)\n", projectName, projectPath)
	if len(keys) > 0 {
		fmt.Fprintf(stdout, "  env: %s\n", strings.Join(keys, ", "))
	}

	if !env.HasActions() {
		return exitOK
	}

	vars := substitutionVars(projectPath, projectName, env)

	acts := actions.New()

	// 1. Clean up a prior activation, if any.
	for _, k := range strings.Fields(env.ActivatedKeys) {
		if validEnvKey(k) {
			acts.UnsetEnv(k)
		}
	}

	// 2. cd when activated by name.
	if cdHere {
		acts.CD(projectPath)
	}

	// 3. New env, with substitution.
	for _, k := range keys {
		acts.SetEnv(k, substitute(pjrc.Env[k], vars))
	}

	// 4. Markers and bookkeeping. These are set last so they win against
	// any same-named key in .pjrc[env] — the plugin owns these names.
	acts.SetEnv("PJ_PROJECT", projectName)
	acts.SetEnv("PJ_PROJECT_PATH", projectPath)
	acts.SetEnv("PJ_ACTIVATED_KEYS", strings.Join(keys, " "))

	if err := acts.Write(env.ActionsFile, actions.ParseShell(env.ShellOrPosix())); err != nil {
		fmt.Fprintf(stderr, "pj-activate: %v\n", err)
		return exitGeneral
	}
	return exitOK
}

func substitutionVars(projectPath, projectName string, env *plugin.Env) map[string]string {
	return map[string]string{
		"PJ_PROJECT_PATH": projectPath,
		"PJ_PROJECT_NAME": projectName,
		"PJ_PROJECT":      projectName,
		"PJ_CONFIG_DIR":   env.ConfigDir,
		"PJ_DATA_DIR":     env.DataDir,
		"HOME":            os.Getenv("HOME"),
	}
}

// substitute replaces ${VAR} for every VAR in vars.
// Unknown ${...} forms are left intact so typos are visible.
func substitute(s string, vars map[string]string) string {
	for k, v := range vars {
		s = strings.ReplaceAll(s, "${"+k+"}", v)
	}
	return s
}

func validEnvKey(k string) bool {
	if k == "" {
		return false
	}
	for i, r := range k {
		switch {
		case r == '_':
		case r >= 'A' && r <= 'Z':
		case r >= 'a' && r <= 'z':
		case r >= '0' && r <= '9' && i > 0:
		default:
			return false
		}
	}
	return true
}
