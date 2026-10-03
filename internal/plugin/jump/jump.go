// Package jump implements the pj-jump subcommand.
//
// It resolves a project by name from the registry, validates the target
// path, prints a human-readable message to stdout, and — when the shell
// wrapper provided an actions file — appends a cd action to it.
//
// See docs/PLUGIN_SPEC.md §6.2 (hybrid plugins) and docs/ACTIONS_SPEC.md.
package jump

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"

	"github.com/danydevid/pj/internal/actions"
	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

// Exit codes follow PLUGIN_SPEC.md §5.
const (
	exitOK         = 0
	exitGeneral    = 1
	exitUsage      = 2
	exitConfig     = 3
	exitNotFound   = 4
	exitAccess     = 5
	exitValidation = 6
)

// Run is the plugin entry point. It returns a process exit code.
func Run(args []string) int {
	return run(args, os.Stdout, os.Stderr, plugin.Load())
}

// run is the testable core. Writers and Env are injected so tests can
// capture output and point at a temp config dir without touching the
// real environment.
func run(args []string, stdout, stderr io.Writer, env *plugin.Env) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: pj jump <project>")
		return exitUsage
	}
	name := args[0]

	reg, err := loadRegistry(env.ConfigDir)
	if err != nil {
		fmt.Fprintf(stderr, "pj-jump: %v\n", err)
		return exitConfig
	}

	item, ok := reg.Projects[name]
	if !ok {
		fmt.Fprintf(stderr, "pj-jump: project '%s' not found\n", name)
		return exitNotFound
	}
	if item.Status == "archived" {
		fmt.Fprintf(stderr, "pj-jump: project '%s' is archived\n", name)
		return exitNotFound
	}
	if !filepath.IsAbs(item.Path) {
		fmt.Fprintf(stderr, "pj-jump: project '%s' has non-absolute path: %s\n", name, item.Path)
		return exitConfig
	}

	// Validate the target before emitting any action.
	// ACTIONS_SPEC.md §7.1 requires stat + is-dir checks here.
	info, err := os.Stat(item.Path)
	if err != nil {
		if os.IsPermission(err) {
			fmt.Fprintf(stderr, "pj-jump: %v\n", err)
			return exitAccess
		}
		fmt.Fprintf(stderr, "pj-jump: cannot access %s: %v\n", item.Path, err)
		return exitAccess
	}
	if !info.IsDir() {
		fmt.Fprintf(stderr, "pj-jump: not a directory: %s\n", item.Path)
		return exitValidation
	}

	// Human-readable output goes to stdout, always.
	fmt.Fprintf(stdout, "Jumping to %s (%s)\n", name, item.Path)

	// Shell actions go to the actions file — but only if the wrapper
	// provided one. Without it we were called directly (or from a GUI),
	// and there is no parent shell to change directory for.
	// See PLUGIN_SPEC.md §6.2 and ACTIONS_SPEC.md §5.1.
	if env.HasActions() {
		acts := actions.New().CD(item.Path)
		if err := acts.Write(env.ActionsFile, actions.ParseShell(env.ShellOrPosix())); err != nil {
			fmt.Fprintf(stderr, "pj-jump: %v\n", err)
			return exitGeneral
		}
	}

	return exitOK
}

// loadRegistry reads projects.toml, returning an empty registry when the
// file does not exist. Other I/O or parse errors are propagated.
func loadRegistry(configDir string) (*config.ProjectsRegistry, error) {
	path := filepath.Join(configDir, "projects.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return emptyRegistry(), nil
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var reg config.ProjectsRegistry
	if err := toml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("%w: cannot parse %s: %v", config.ErrConfigInvalid, path, err)
	}
	if reg.PJSpecVersion == "" {
		reg.PJSpecVersion = config.ExpectedSpecVersion
	}
	if err := reg.Validate(); err != nil {
		return nil, err
	}
	return &reg, nil
}

func emptyRegistry() *config.ProjectsRegistry {
	return &config.ProjectsRegistry{
		PJSpecVersion: config.ExpectedSpecVersion,
		Projects:      map[string]config.ProjectItem{},
	}
}
