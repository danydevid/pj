# pj — project jumper

A tiny, extensible project jumper. `pj jump foo` moves your shell to a
registered project, applies its environment, and gets out of the way.

`pj` is built as a **microkernel + plugin** architecture:

- **Manager** (`pj`) resolves a subcommand to an executable and execs it.
  No config parsing, no state, no business logic.
- **Plugins** (`pj-<subcommand>`) are executables that do the actual work.
  Any language, any size, installed independently.
- **Shell wrapper** sources shell actions written by plugins, so plugins
  can change the parent shell (`cd`, `export`, `alias`).

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) for the full design.

---

## Status

`v0.1.0-mvp` — usable, but read the fine print.

| Area | State |
| --- | --- |
| `pj init`, `pj list`, `pj jump` | working |
| `pj activate`, `pj deactivate`, `pj root` | working |
| Bash wrapper | verified end-to-end |
| Zsh wrapper | code complete, **untested** |
| Fish wrapper | code complete, **untested** |
| Windows | not supported |

No plugin management yet. No template instantiation yet (`.pjrc` templates
are read but not executed). No completions yet.

---

## Install

### From source (only supported path today)

Requires Go 1.22+ and a POSIX shell.

```bash
git clone https://github.com/danydevid/pj
cd pj
./scripts/build.sh
./scripts/install.sh
```

`install.sh` installs:

```
~/.local/bin/pj            manager
~/.pj/bin/pj-*             plugins
~/.config/pj/init.bash     shell wrappers
~/.config/pj/init.zsh
~/.config/pj/init.fish
```

If `~/.local/bin` is not on your `PATH`, the installer prints the line to
add to your shell rc. Override locations with `PREFIX`, `BIN_DIR`,
`PLUGIN_DIR`, or `CONFIG_DIR`:

```bash
BIN_DIR=/usr/local/bin PLUGIN_DIR=/opt/pj/bin ./scripts/install.sh
```

There is no `curl | sh` installer yet. See [Future](#future).

---

## Shell integration

`pj jump` can only change your shell's working directory if a wrapper is
sourced in your **interactive** shell. Without it, `pj jump` still works,
but prints the path instead of `cd`-ing.

**Bash** — add to `~/.bashrc`:

```bash
source ~/.config/pj/init.bash
```

**Zsh** — add to `~/.zshrc`:

```zsh
source ~/.config/pj/init.zsh
```

**Fish** — add to `~/.config/fish/config.fish`:

```fish
source ~/.config/pj/init.fish
```

Reload your rc (or open a new shell) and verify:

```bash
type pj        # should say "pj is a function"
```

---

## Getting started

```bash
pj init                     # create ~/.config/pj and friends
```

Register a project by editing `~/.config/pj/projects.toml`:

```toml
pj_spec_version = "1"

[project.foo]
path = "/home/user/projects/foo"
status = "active"
tags = ["go", "cli"]
```

```bash
pj list                     # tabular view
pj list --json              # machine-readable
pj jump foo                 # cd into it (wrapper required)
```

---

## Commands

### `pj init`

Create the XDG directory structure, `config.toml`, `projects.toml`, and
install shell wrappers to `~/.config/pj/`. Idempotent — existing files are
never overwritten.

### `pj list`

List registered projects. Archived projects are hidden by default.

```
pj list           # table
pj list --all     # include archived
pj list --json    # JSON
```

### `pj jump <project>`

`cd` into the project (wrapper) and print a one-line summary.

Exit codes:

| Code | Meaning |
| --- | --- |
| 0 | success |
| 2 | usage error |
| 3 | configuration error |
| 4 | project not found or archived |
| 5 | target path missing or unreadable |
| 6 | target path is not a directory |

### `pj activate [project]`

Load the project's `.pjrc` `[env]` section into the current shell. With a
project name, it also `cd`s. Without arguments, it activates the nearest
project discovered by walking upward from `$PWD` looking for `.pjrc`.

Records activated keys in `PJ_ACTIVATED_KEYS`, so a later `pj deactivate`
knows what to unset.

### `pj deactivate`

Unset every key recorded by the last `pj activate`, plus the `PJ_PROJECT`
markers. Idempotent — reports "No active project" and exits 0 when nothing
is active.

### `pj root`

Walk upward from `$PWD` to the nearest `.pjrc` and `cd` there. No registry
needed — works for freshly cloned repositories.

---

## Configuration

All configuration lives under `$XDG_CONFIG_HOME/pj/`. Full details in
[`docs/CONFIG_SPEC.md`](docs/CONFIG_SPEC.md).

| File | Purpose |
| --- | --- |
| `config.toml` | User preferences |
| `projects.toml` | Project registry |
| `templates/` | Project templates |
| `<project>/.pjrc` | Per-project env, tasks, hooks |

Example `.pjrc`:

```toml
pj_spec_version = "1"

[project]
name = "foo"
type = "go-cli"

[env]
GOFLAGS = "-mod=vendor"
PROJECT_ROOT = "${PJ_PROJECT_PATH}"

[tasks]
build = "go build ./..."
test  = "go test ./..."
```

Variable substitution supported in `.pjrc`: `${PJ_PROJECT_PATH}`,
`${PJ_PROJECT_NAME}`, `${PJ_CONFIG_DIR}`, `${PJ_DATA_DIR}`, `${HOME}`.

---

## File locations

`pj` follows the [XDG Base Directory Specification](docs/XDG_SPEC.md).

| Purpose | Default |
| --- | --- |
| Config | `~/.config/pj/` |
| Data | `~/.local/share/pj/` |
| Cache | `~/.cache/pj/` |
| Runtime | `$XDG_RUNTIME_DIR/pj/` |
| State | `~/.local/state/pj/` |
| Plugins | `~/.pj/bin/` |

Every path can be overridden via `PJ_CONFIG_DIR`, `PJ_DATA_DIR`,
`PJ_CACHE_DIR`, `PJ_RUNTIME_DIR`, `PJ_PLUGIN_DIR`.

---

## Writing a plugin

A plugin is an executable named `pj-<subcommand>`. The manager resolves it
from `$PJ_PLUGIN_DIR`, `~/.pj/bin/`, `$XDG_DATA_HOME/pj/bin/`, or `$PATH`
(in that order), then execs it with the user's arguments.

### Pure plugin

Prints to stdout/stderr, exits with a meaningful code. Examples: `pj-list`,
`pj-info`.

```bash
#!/usr/bin/env bash
# ~/.pj/bin/pj-hello
set -euo pipefail
printf 'hello, %s\n' "${1:-world}"
```

### Hybrid plugin

Also writes shell actions to `$PJ_ACTIONS_FILE`. The wrapper sources that
file in the parent shell.

```go
package main

import (
	"os"

	"danydevid/pj/internal/actions"
	"danydevid/pj/internal/plugin"
)

func main() {
	env := plugin.Load()
	if !env.HasActions() {
		// No wrapper — print only, do not attempt cd.
		return
	}

	acts := actions.New().CD("/some/absolute/path")
	_ = acts.Write(env.ActionsFile, actions.ParseShell(env.ShellOrPosix()))
}
```

Full contract in [`docs/PLUGIN_SPEC.md`](docs/PLUGIN_SPEC.md) and
[`docs/ACTIONS_SPEC.md`](docs/ACTIONS_SPEC.md).

---

## Development

```bash
make build        # compile manager + plugins into bin/
make test         # go test ./...
make install      # build, then install
make lint         # golangci-lint
make fmt          # gofmt + goimports
```

End-to-end smoke test in an isolated `$HOME`:

```bash
./test.sh
```

It builds, installs into a temp directory, registers a project whose path
contains spaces, a single quote, and a `$`, then verifies that each shell
wrapper actually changes `$PWD` and cleans up its tmpfile.

Requires `zsh` and `fish` for full coverage; each is skipped if absent.

See [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md).

---

## Design rules

`pj` is intentionally small. These rules are load-bearing:

1. **Mechanism in the manager, policy in plugins.**
2. **State on the filesystem, not in a daemon.**
3. **Shell integration via wrapper, not `eval`.**
4. **Actions in a dedicated file, not interleaved with stdout.**
5. **Constrain what can grow** — the manager stays under 500 lines.

The manager does not import `internal/config`; it does not parse
configuration; it does not store state.

See [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) §8 for the full list.

---

## Future

Planned, in rough order:

- `pj init` embeds the shell wrappers via `go:embed` (single-binary init).
- `pj plugin install <url>` for plugin distribution.
- `pj template create` for project scaffolding.
- Shell completions for bash, zsh, fish.
- Verified zsh and fish support.
- `curl | sh` bootstrap installer, once releases are published.

Not planned: shared-object plugins, daemons, plugin registries, SQLite,
event buses, worker pools, GUIs, HTTP APIs, plugin sandboxing.

Each of those adds complexity without a demonstrated need. They will be
added only when concrete evidence demands it.

---

## Documentation

| File | Contents |
| --- | --- |
| [`docs/ARCHITECTURE.md`](docs/ARCHITECTURE.md) | Components, flow, decisions |
| [`docs/CONFIG_SPEC.md`](docs/CONFIG_SPEC.md) | TOML schemas |
| [`docs/PLUGIN_SPEC.md`](docs/PLUGIN_SPEC.md) | Plugin contract |
| [`docs/ACTIONS_SPEC.md`](docs/ACTIONS_SPEC.md) | Shell action channel |
| [`docs/XDG_SPEC.md`](docs/XDG_SPEC.md) | File locations |
| [`docs/CONTRIBUTING.md`](docs/CONTRIBUTING.md) | How to contribute |

---

## License

MIT. See [`LICENSE`](LICENSE).
