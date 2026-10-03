package actions

import (
	"os"
	"strings"
)

// Shell identifies the target shell for rendering.
//
// Values mirror the PJ_SHELL variable set by the shell wrapper.
// See docs/ACTIONS_SPEC.md §4.
type Shell string

const (
	Sh     Shell = "sh"
	Bash   Shell = "bash"
	Zsh    Shell = "zsh"
	Fish   Shell = "fish"
	Nu     Shell = "nu"
	Pwsh   Shell = "pwsh"
	Elvish Shell = "elvish"
)

// ParseShell normalizes a raw PJ_SHELL value.
//
// Matching is case-insensitive and trims surrounding whitespace.
// Unknown or empty values fall back to Sh per ACTIONS_SPEC.md §4.
func ParseShell(s string) Shell {
	n := Shell(strings.TrimSpace(strings.ToLower(s)))
	switch n {
	case Sh, Bash, Zsh, Fish, Nu, Pwsh, Elvish:
		return n
	default:
		return Sh
	}
}

// DetectShell reads PJ_SHELL from the process environment and normalizes
// it. Convenience for plugins that don't already carry a plugin.Env.
func DetectShell() Shell {
	return ParseShell(os.Getenv("PJ_SHELL"))
}

// IsPOSIX reports whether s shares POSIX sh syntax for the action
// vocabulary we emit (cd, export, unset, alias, source).
func (s Shell) IsPOSIX() bool {
	switch s {
	case Sh, Bash, Zsh:
		return true
	}
	return false
}

// String returns the canonical lowercase name.
func (s Shell) String() string { return string(s) }

// escapeFor dispatches to the appropriate escape function for s.
// Unknown shells default to POSIX escaping.
func escapeFor(s Shell, v string) string {
	switch {
	case s.IsPOSIX():
		return EscapePOSIX(v)
	case s == Fish:
		return EscapeFish(v)
	case s == Nu:
		return EscapeNu(v)
	case s == Pwsh:
		return EscapePwsh(v)
	}
	return EscapePOSIX(v)
}
