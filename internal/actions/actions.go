// Package actions implements the shell-agnostic action channel used by
// hybrid plugins. See docs/PLUGIN_SPEC.md §6.2 and docs/ACTIONS_SPEC.md.
//
// Plugins construct an Actions value with the fluent builders, then call
// Write (or WriteEnv) to emit valid shell code for the target shell into
// $PJ_ACTIONS_FILE. The shell wrapper sources that file in the parent
// shell context.
//
// The library performs syntactic validation only (absolute paths, valid
// env keys, valid alias names). It does not touch the filesystem for
// path-existence checks — that is the plugin's responsibility per
// docs/ACTIONS_SPEC.md §7.
package actions

import (
	"errors"
	"fmt"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	// ErrInvalidPath is returned when a path-valued action is not absolute.
	ErrInvalidPath = errors.New("path must be absolute")

	// ErrInvalidKey is returned for malformed environment variable names.
	ErrInvalidKey = errors.New("invalid environment variable name")

	// ErrInvalidAlias is returned for malformed alias names.
	ErrInvalidAlias = errors.New("invalid alias name")

	// ErrUnsupported is returned when an action has no defined syntax for
	// the target shell.
	ErrUnsupported = errors.New("action not supported for shell")
)

var (
	envKeyRegex = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
	aliasRegex  = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// Action is one shell-level action. The interface is intentionally
// unexported in its method set: only this package defines the vocabulary.
type Action interface {
	render(s Shell) (string, error)
	validate() error
}

// Actions is an ordered collection of shell actions.
//
// Order is preserved. The zero value is usable and empty; prefer New
// for symmetry with the fluent builders.
type Actions struct {
	items []Action
}

// New returns an empty Actions.
func New() *Actions { return &Actions{} }

// Add appends an arbitrary action. Mostly useful internally; the
// strongly typed builders below are the intended public surface.
func (a *Actions) Add(act Action) *Actions {
	a.items = append(a.items, act)
	return a
}

// CD emits a directory change.
// The path must be absolute (checked by Validate).
func (a *Actions) CD(path string) *Actions {
	return a.Add(cdAction{path: path})
}

// SetEnv emits an environment variable assignment.
func (a *Actions) SetEnv(key, value string) *Actions {
	return a.Add(setEnvAction{key: key, value: value})
}

// UnsetEnv emits an environment variable removal.
func (a *Actions) UnsetEnv(key string) *Actions {
	return a.Add(unsetEnvAction{key: key})
}

// Alias emits a shell alias.
func (a *Actions) Alias(name, command string) *Actions {
	return a.Add(aliasAction{name: name, command: command})
}

// Unalias removes a shell alias.
func (a *Actions) Unalias(name string) *Actions {
	return a.Add(unaliasAction{name: name})
}

// PrependPath prepends dir to $PATH.
// The directory must be absolute (checked by Validate).
func (a *Actions) PrependPath(dir string) *Actions {
	return a.Add(prependPathAction{dir: dir})
}

// Source emits a source/dot directive for an existing shell file.
// The path must be absolute (checked by Validate).
func (a *Actions) Source(path string) *Actions {
	return a.Add(sourceAction{path: path})
}

// Prompt sets the interactive prompt (bash/zsh only).
func (a *Actions) Prompt(value string) *Actions {
	return a.Add(promptAction{value: value})
}

// Len returns the number of queued actions.
func (a *Actions) Len() int { return len(a.items) }

// IsEmpty reports whether no actions are queued.
func (a *Actions) IsEmpty() bool { return len(a.items) == 0 }

// Validate checks syntactic invariants without touching the filesystem.
// Called automatically by Write; exposed for callers that want to
// validate before committing to a file operation.
func (a *Actions) Validate() error {
	for i, act := range a.items {
		if err := act.validate(); err != nil {
			return fmt.Errorf("action[%d]: %w", i, err)
		}
	}
	return nil
}

// Render produces shell code for s.
//
// Each rendered line ends with a trailing newline. Empty lines produced
// by actions that intentionally no-op for s are dropped. Returns an
// empty string when no action emits output.
func (a *Actions) Render(s Shell) (string, error) {
	var b strings.Builder
	for i, act := range a.items {
		line, err := act.render(s)
		if err != nil {
			return "", fmt.Errorf("action[%d]: %w", i, err)
		}
		line = strings.TrimRight(line, "\n")
		if line == "" {
			continue
		}
		b.WriteString(line)
		b.WriteByte('\n')
	}
	return b.String(), nil
}

// ---------------------------------------------------------------------------
// Action implementations
// ---------------------------------------------------------------------------

type cdAction struct{ path string }

func (c cdAction) render(s Shell) (string, error) {
	esc := escapeFor(s, c.path)
	switch {
	case s.IsPOSIX(), s == Fish, s == Nu:
		return "cd " + esc, nil
	case s == Pwsh:
		return "Set-Location " + esc, nil
	}
	return "", fmt.Errorf("%w: cd for %q", ErrUnsupported, s)
}

func (c cdAction) validate() error {
	if !filepath.IsAbs(c.path) {
		return fmt.Errorf("%w: cd %q", ErrInvalidPath, c.path)
	}
	return nil
}

type setEnvAction struct{ key, value string }

func (e setEnvAction) render(s Shell) (string, error) {
	esc := escapeFor(s, e.value)
	switch {
	case s.IsPOSIX():
		return fmt.Sprintf("export %s=%s", e.key, esc), nil
	case s == Fish:
		return fmt.Sprintf("set -gx %s %s", e.key, esc), nil
	case s == Nu:
		return fmt.Sprintf("$env.%s = %s", e.key, esc), nil
	case s == Pwsh:
		return fmt.Sprintf("$env:%s = %s", e.key, esc), nil
	}
	return "", fmt.Errorf("%w: set env for %q", ErrUnsupported, s)
}

func (e setEnvAction) validate() error {
	if !envKeyRegex.MatchString(e.key) {
		return fmt.Errorf("%w: %q", ErrInvalidKey, e.key)
	}
	return nil
}

type unsetEnvAction struct{ key string }

func (u unsetEnvAction) render(s Shell) (string, error) {
	switch {
	case s.IsPOSIX():
		return "unset " + u.key, nil
	case s == Fish:
		return "set -e " + u.key, nil
	case s == Nu:
		return "hide " + u.key, nil
	case s == Pwsh:
		return fmt.Sprintf("Remove-Item Env:\\%s -ErrorAction SilentlyContinue", u.key), nil
	}
	return "", fmt.Errorf("%w: unset env for %q", ErrUnsupported, s)
}

func (u unsetEnvAction) validate() error {
	if !envKeyRegex.MatchString(u.key) {
		return fmt.Errorf("%w: %q", ErrInvalidKey, u.key)
	}
	return nil
}

type aliasAction struct{ name, command string }

func (a aliasAction) render(s Shell) (string, error) {
	esc := escapeFor(s, a.command)
	switch {
	case s.IsPOSIX(), s == Fish:
		return fmt.Sprintf("alias %s=%s", a.name, esc), nil
	case s == Nu:
		return fmt.Sprintf("alias %s = %s", a.name, esc), nil
	}
	return "", fmt.Errorf("%w: alias for %q", ErrUnsupported, s)
}

func (a aliasAction) validate() error {
	if !aliasRegex.MatchString(a.name) {
		return fmt.Errorf("%w: %q", ErrInvalidAlias, a.name)
	}
	return nil
}

type unaliasAction struct{ name string }

func (u unaliasAction) render(s Shell) (string, error) {
	switch {
	case s.IsPOSIX():
		return "unalias " + u.name, nil
	case s == Fish:
		return "functions -e " + u.name, nil
	case s == Nu:
		return "hide " + u.name, nil
	}
	return "", fmt.Errorf("%w: unalias for %q", ErrUnsupported, s)
}

func (u unaliasAction) validate() error {
	if !aliasRegex.MatchString(u.name) {
		return fmt.Errorf("%w: %q", ErrInvalidAlias, u.name)
	}
	return nil
}

type prependPathAction struct{ dir string }

func (p prependPathAction) render(s Shell) (string, error) {
	esc := escapeFor(s, p.dir)
	switch {
	case s.IsPOSIX():
		return fmt.Sprintf("export PATH=%s:\"$PATH\"", esc), nil
	case s == Fish:
		return fmt.Sprintf("set -gx PATH %s $PATH", esc), nil
	case s == Nu:
		return fmt.Sprintf("$env.PATH = ($env.PATH | prepend %s)", esc), nil
	case s == Pwsh:
		return fmt.Sprintf("$env:PATH = %s + [IO.Path]::PathSeparator + $env:PATH", esc), nil
	}
	return "", fmt.Errorf("%w: prepend PATH for %q", ErrUnsupported, s)
}

func (p prependPathAction) validate() error {
	if !filepath.IsAbs(p.dir) {
		return fmt.Errorf("%w: prepend PATH %q", ErrInvalidPath, p.dir)
	}
	return nil
}

type sourceAction struct{ path string }

func (src sourceAction) render(s Shell) (string, error) {
	esc := escapeFor(s, src.path)
	switch {
	case s == Sh:
		return ". " + esc, nil
	case s == Bash, s == Zsh:
		return "source " + esc, nil
	case s == Fish, s == Nu:
		return "source " + esc, nil
	case s == Pwsh:
		return ". " + esc, nil
	}
	return "", fmt.Errorf("%w: source for %q", ErrUnsupported, s)
}

func (src sourceAction) validate() error {
	if !filepath.IsAbs(src.path) {
		return fmt.Errorf("%w: source %q", ErrInvalidPath, src.path)
	}
	return nil
}

type promptAction struct{ value string }

func (p promptAction) render(s Shell) (string, error) {
	switch s {
	case Bash, Zsh:
		return "PS1=" + escapeFor(s, p.value), nil
	}
	return "", fmt.Errorf("%w: prompt for %q", ErrUnsupported, s)
}

func (p promptAction) validate() error { return nil }
