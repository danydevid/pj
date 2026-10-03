package actions

import "strings"

// EscapePOSIX single-quotes s for POSIX shells (sh, bash, zsh).
//
// A literal single quote is closed, escaped, and reopened — the canonical
// '\” sequence. Mirrors ACTIONS_SPEC.md §6.1.
func EscapePOSIX(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

// EscapeFish single-quotes s for fish.
//
// Backslashes are doubled first, then single quotes are backslash-escaped,
// then the whole value is wrapped in single quotes. Mirrors
// ACTIONS_SPEC.md §6.2.
func EscapeFish(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "'", "\\'")
	return "'" + s + "'"
}

// EscapeNu single-quotes s for nushell.
//
// Nushell uses doubled single quotes inside a single-quoted string.
// Mirrors ACTIONS_SPEC.md §6.3.
func EscapeNu(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

// EscapePwsh single-quotes s for PowerShell.
//
// Same rule as nushell: doubled single quotes inside single quotes.
// Mirrors ACTIONS_SPEC.md §6.4.
func EscapePwsh(s string) string {
	return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}
