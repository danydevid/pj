// Package root implements `pj root`.
//
// Walks upward from the current directory looking for the nearest
// .pjrc, then emits a cd action to that project's root.
// See docs/CONFIG_SPEC.md §4 and docs/ACTIONS_SPEC.md §5.1.
package root

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/danydevid/pj/internal/actions"
	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

func Run(args []string) int {
	cwd, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "pj-root: %v\n", err)
		return 1
	}
	return run(args, cwd, os.Stdout, os.Stderr, plugin.Load())
}

func run(args []string, cwd string, stdout, stderr io.Writer, env *plugin.Env) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: pj root")
		return 2
	}

	pjrcPath, err := config.FindPJRC(cwd)
	if err != nil {
		fmt.Fprintf(stderr, "pj-root: %v\n", err)
		return 4
	}
	projectRoot := filepath.Dir(pjrcPath)

	if samePath(cwd, projectRoot) {
		fmt.Fprintf(stdout, "Already at project root: %s\n", projectRoot)
		return 0
	}

	fmt.Fprintf(stdout, "Project root: %s\n", projectRoot)

	if env.HasActions() {
		acts := actions.New().CD(projectRoot)
		if err := acts.Write(env.ActionsFile, actions.ParseShell(env.ShellOrPosix())); err != nil {
			fmt.Fprintf(stderr, "pj-root: %v\n", err)
			return 1
		}
	}
	return 0
}

// samePath compares two paths after resolving symlinks when possible.
func samePath(a, b string) bool {
	ra, err1 := filepath.EvalSymlinks(a)
	rb, err2 := filepath.EvalSymlinks(b)
	if err1 == nil && err2 == nil {
		return ra == rb
	}
	return a == b
}
