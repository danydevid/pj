package actions

import "os"

// Write renders the actions for shell s and writes them to path.
//
// Semantics:
//   - Empty path is a no-op. This matches the plugin contract that
//     hybrid plugins continue execution when the shell wrapper did not
//     set PJ_ACTIONS_FILE (see PLUGIN_SPEC.md §6.2).
//   - Empty Actions is a no-op — the file is not touched.
//   - The file is written with 0600. The wrapper is expected to have
//     created it via mktemp; the perm only applies if the file is new.
//   - Validate is called before any bytes are written, so a validation
//     failure leaves the actions file untouched.
//   - The file is truncated, not appended. One plugin, one writer, one
//     actions file — see ARCHITECTURE.md §4.
func (a *Actions) Write(path string, s Shell) error {
	if path == "" {
		return nil
	}
	if a.IsEmpty() {
		return nil
	}
	if err := a.Validate(); err != nil {
		return err
	}
	code, err := a.Render(s)
	if err != nil {
		return err
	}
	if code == "" {
		return nil
	}
	return os.WriteFile(path, []byte(code), 0600)
}

// WriteEnv is a convenience for Write that reads PJ_ACTIONS_FILE and
// PJ_SHELL from the process environment. Plugins that already carry a
// plugin.Env should call Write directly with the cached values.
func (a *Actions) WriteEnv() error {
	return a.Write(os.Getenv("PJ_ACTIONS_FILE"), DetectShell())
}
