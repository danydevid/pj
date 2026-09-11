# Plugin Specification `pj`

## 1. Summary

A plugin is an **executable binary** that implements a single feature. The manager only performs `fork+exec` on the plugin. There is no IPC protocol and no registry required.

## 2. Base Contract

### 2.1 Naming

```text
pj-<subcommand>

```

* `subcommand`: `[a-z0-9-]+`, maximum 32 characters.
* Examples: `pj-jump`, `pj-list`, `pj-init`, `pj-template-create`.
* Case-sensitive: must be lowercase.

### 2.2 Location

Plugins are searched for in the following directories, ordered by priority:

1. `$PJ_PLUGIN_DIR` (optional override)
2. `~/.pj/bin/`
3. `$XDG_DATA_HOME/pj/bin/`
4. `$PATH` (fallback)

If present in multiple locations, the first match takes precedence.

### 2.3 Permissions

* The file must be executable (`chmod +x`).
* Ownership must belong to the user.
* Must not be world-writable.
* Must not be a symlink target outside the plugin directory.

### 2.4 Arguments

The manager executes the plugin as follows:

```text
pj-<subcommand> [args...]

```

`args` represents arguments passed after the subcommand. Example:

```text
$ pj jump foo --verbose
→ manager exec: pj-jump foo --verbose

```

Plugins are free to interpret arguments as needed.

## 3. Environment Variables

The manager sets the following environment variables prior to executing a plugin:

| Variable | Required | Description |
| --- | --- | --- |
| `PJ_MANAGER` | yes | Manager binary path (`pj`) |
| `PJ_VERSION` | yes | Manager version (semver) |
| `PJ_CONFIG_DIR` | yes | `$XDG_CONFIG_HOME/pj` |
| `PJ_DATA_DIR` | yes | `$XDG_DATA_HOME/pj` |
| `PJ_CACHE_DIR` | yes | `$XDG_CACHE_HOME/pj` |
| `PJ_RUNTIME_DIR` | yes | `$XDG_RUNTIME_DIR/pj` |
| `PJ_ACTIONS_FILE` | if wrapper | Actions tmpfile path |
| `PJ_SHELL` | if wrapper | `bash`, `zsh`, `fish`, `nu`, `sh` |
| `PJ_PROJECT` | if active | Active project name (optional) |
| `PJ_PROJECT_PATH` | if active | Active project path (optional) |

Plugins **must not** assume environment variables are always set. Always check their existence.

## 4. Output

### 4.1 stdout

Human-readable output. Can be:

* Plain text.
* Tables.
* JSON (if `--json` is provided).

The manager **passes stdout through** to the terminal. Output formatting is left to the plugin.

### 4.2 stderr

Errors and logs. The manager passes stderr through to the terminal.

### 4.3 Error Formatting

```text
pj-<subcommand>: <message>

```

Examples:

```text
pj-jump: project 'foo' not found
pj-jump: cannot read config: permission denied

```

### 4.4 `--json`

Plugins are **recommended** to support `--json` for machine-readable output.

```text
$ pj list --json
{"projects":[{"name":"foo","path":"/home/user/foo","status":"active"}]}

```

The JSON structure is arbitrary but must be valid JSON.

## 5. Exit Codes

| Code | Meaning |
| --- | --- |
| 0 | Success |
| 1 | General error |
| 2 | Usage error (invalid arguments) |
| 3 | Configuration error |
| 4 | Not found (project, plugin, file) |
| 5 | Permission error |
| 6 | Validation error |
| 127 | Plugin not found (emitted by manager) |
| 128+N | Terminated by signal N |

Plugins **must** return meaningful exit codes. Do not default everything to 0.

## 6. Plugin Classification

### 6.1 Pure Plugins

Plugins that do not modify the parent shell environment. Output is strictly directed to stdout/stderr.

Examples: `pj-list`, `pj-status`, `pj-info`, `pj-template-list`.

Contract:

* Do not write to `PJ_ACTIONS_FILE`.
* Invokable by GUI applications, scripts, or human users.

### 6.2 Hybrid Plugins

Plugins that need to modify the parent shell environment (`cd`, `export`, `alias`).

Examples: `pj-jump`, `pj-activate`, `pj-use`.

Contract:

* Write shell actions to `$PJ_ACTIONS_FILE`.
* Inspect `PJ_SHELL` for appropriate syntax.
* If `PJ_ACTIONS_FILE` is empty or unset, continue execution (omitting `cd`).

### 6.3 Internal Plugins

Plugins called by wrappers or the manager directly, rather than human users.

Examples: `pj-completion`, `pj-prompt`.

Contract:

* Same behavior as pure plugins, but with defined output formats.
* Refer to respective individual specifications.

## 7. Actions File

Refer to `ACTIONS_SPEC.md` for complete details.

Summary:

* Hybrid plugins write shell code to `$PJ_ACTIONS_FILE`.
* Format: valid shell code for the specified `PJ_SHELL`.
* Shell wrapper sources this file in the parent shell context.

## 8. Manifest (Optional)

Plugins **may** include a manifest file for metadata.

Location: `pj-<subcommand>.manifest.toml` in the same directory as the executable.

```toml
# ~/.pj/bin/pj-jump.manifest.toml

[plugin]
name = "jump"
version = "1.0.0"
description = "Jump to a project directory"
api_version = "1"
class = "hybrid"        # pure | hybrid | internal
entry = "pj-jump"
author = "John Doe"
license = "MIT"
homepage = "https://github.com/example/pj-jump"

[actions]
provides = ["cd"]       # types of shell actions generated

[permissions]
filesystem = ["$HOME/.config/pj", "$HOME/projects"]
network = false
exec = ["git", "go"]

```

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | string | yes | Plugin name (excluding `pj-`) |
| `version` | string | yes | Semver |
| `description` | string | no | Short description |
| `api_version` | string | yes | Plugin API version |
| `class` | string | yes | `pure`, `hybrid`, `internal` |
| `entry` | string | yes | Binary executable name |
| `author` | string | no | Author |
| `license` | string | no | License |
| `homepage` | string | no | URL |
| `actions.provides` | array[string] | no | Action types provided |
| `permissions.filesystem` | array[string] | no | Accessed paths |
| `permissions.network` | bool | no | Requires network access |
| `permissions.exec` | array[string] | no | Binaries executed |

If the manifest is missing, the plugin remains valid. Manifests are purely metadata.

## 9. Plugin API Versioning

The Plugin API version is a string. Current version: `"1"`.

The manager inspects `api_version` in the manifest:

* If no manifest exists, assume compatibility.
* If `api_version` does not match the manager's supported version, reject with an error.
* If `api_version` is older, attempt execution with a warning.

## 10. Plugin Example: Pure (Shell Script)

```bash
#!/usr/bin/env bash
# ~/.pj/bin/pj-list

set -euo pipefail

config_dir="${PJ_CONFIG_DIR:-$HOME/.config/pj}"
registry="$config_dir/projects.toml"

if [[ ! -f "$registry" ]]; then
    echo "no projects registered" >&2
    exit 4
fi

# Simple parser (for demonstration purposes)
grep -E '^\[project\.' "$registry" \
    | sed 's/\[project\.\(.*\)\]/\1/' \
    | sort

```

## 11. Plugin Example: Hybrid (Go)

```go
// ~/.pj/bin/pj-jump (compiled Go binary)
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type ShellActions struct {
	CD    string            `json:"cd,omitempty"`
	Env   map[string]string `json:"env,omitempty"`
	Unset []string          `json:"unset,omitempty"`
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: pj jump <project>")
		os.Exit(2)
	}

	name := os.Args[1]
	configDir := os.Getenv("PJ_CONFIG_DIR")
	if configDir == "" {
		configDir = filepath.Join(os.Getenv("HOME"), ".config", "pj")
	}

	// Read registry (simple implementation assumption)
	proj, err := findProject(configDir, name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pj-jump: %v\n", err)
		os.Exit(4)
	}

	// Path validation
	if _, err := os.Stat(proj); err != nil {
		fmt.Fprintf(os.Stderr, "pj-jump: %v\n", err)
		os.Exit(5)
	}

	// Print human output
	fmt.Printf("Jumping to %s (%s)\n", name, proj)

	// Write shell actions
	if f := os.Getenv("PJ_ACTIONS_FILE"); f != "" {
		a := ShellActions{CD: proj}
		data, _ := json.Marshal(a)
		if err := os.WriteFile(f, data, 0600); err != nil {
			fmt.Fprintf(os.Stderr, "pj-jump: %v\n", err)
			os.Exit(1)
		}
	}
}

```

## 12. Plugin Testing

Plugins can be tested independently of the manager:

```text
$ PJ_CONFIG_DIR=/tmp/pj-test \
  PJ_ACTIONS_FILE=/tmp/actions \
  PJ_SHELL=bash \
  ./pj-jump foo

$ cat /tmp/actions
cd '/home/user/projects/foo'

```

## 13. Plugin Distribution

Recommended distribution channels:

* Standalone binaries (Go, Rust).
* Executable shell scripts.
* Archives containing binary + manifest.
* `pj plugin install <url>` (when plugin management is available).

## 14. Checklist for New Plugins

* [ ] Executable named `pj-<subcommand>`.
* [ ] Handles `--help`.
* [ ] Handles `--json` (if pure).
* [ ] Returns meaningful exit codes.
* [ ] Outputs errors to stderr prefixed with `pj-<subcommand>:`.
* [ ] If hybrid: writes actions to `$PJ_ACTIONS_FILE`.
* [ ] If hybrid: checks `PJ_SHELL` for syntax compliance.
* [ ] If hybrid: validates payload before writing actions.
* [ ] Includes manifest (optional, but recommended).
* [ ] Verified via standalone testing without manager.
