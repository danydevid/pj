// Package deactivate implements `pj deactivate`.
//
// Reads PJ_ACTIVATED_KEYS from the environment (written by a previous
// pj-activate) and emits unset actions for each key, plus the three
// markers pj-activate maintains.
//
// Idempotent: when no project is active, it prints a message and exits 0.
package deactivate

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/danydevid/pj/internal/actions"
	"github.com/danydevid/pj/internal/plugin"
)

func Run(args []string) int {
	return run(args, os.Stdout, os.Stderr, plugin.Load())
}

func run(args []string, stdout, stderr io.Writer, env *plugin.Env) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: pj deactivate")
		return 2
	}

	if env.Project == "" {
		fmt.Fprintln(stdout, "No active project.")
		return 0
	}

	fmt.Fprintf(stdout, "Deactivating %s\n", env.Project)

	if !env.HasActions() {
		return 0
	}

	acts := actions.New()
	for _, k := range strings.Fields(env.ActivatedKeys) {
		if validEnvKey(k) {
			acts.UnsetEnv(k)
		}
	}
	acts.UnsetEnv("PJ_PROJECT").
		UnsetEnv("PJ_PROJECT_PATH").
		UnsetEnv("PJ_ACTIVATED_KEYS")

	if err := acts.Write(env.ActionsFile, actions.ParseShell(env.ShellOrPosix())); err != nil {
		fmt.Fprintf(stderr, "pj-deactivate: %v\n", err)
		return 1
	}
	return 0
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
