# Config Spec `pj`

## 1. File Locations

| File | Location | Format |
| --- | --- | --- |
| Global config | `$XDG_CONFIG_HOME/pj/config.toml` | TOML |
| Project registry | `$XDG_CONFIG_HOME/pj/projects.toml` | TOML |
| Per-project config | `<project-path>/.pjrc` | TOML |
| Templates | `$XDG_CONFIG_HOME/pj/templates/` | Directory |
| Cache | `$XDG_CACHE_HOME/pj/` | Arbitrary |
| Data | `$XDG_DATA_HOME/pj/` | Arbitrary |
| Plugins | `~/.pj/bin/` | Binary |

Default XDG paths:

* `XDG_CONFIG_HOME` = `~/.config`
* `XDG_DATA_HOME` = `~/.local/share`
* `XDG_CACHE_HOME` = `~/.cache`
* `XDG_RUNTIME_DIR` = `$XDG_RUNTIME_DIR` (fallback: `~/.run`)

## 2. `config.toml`

Global user configuration.

```toml
# ~/.config/pj/config.toml

pj_spec_version = "1"

[user]
name = "John Doe"
email = "john@example.com"

[defaults]
editor = "nvim"
shell = "bash"
template = "go-cli"

[paths]
# override default XDG paths (optional)
# projects = "/home/user/Code"
# templates = "/home/user/.config/pj/templates"

[behavior]
# auto_cd after jump? (wrapper only)
auto_cd = true
# auto_activate project env?
auto_activate = false
# update last_used on jump?
track_usage = true

[logging]
level = "info"        # debug | info | warn | error
format = "text"       # text | json
file = ""             # empty = stderr

```

### Fields

| Field | Type | Required | Default | Description |
| --- | --- | --- | --- | --- |
| `pj_spec_version` | string | yes | — | Config specification version |
| `user.name` | string | no | `$USER` | User name |
| `user.email` | string | no | — | User email |
| `defaults.editor` | string | no | `$EDITOR` | Default editor |
| `defaults.shell` | string | no | `$SHELL` | Default shell |
| `defaults.template` | string | no | — | Default template |
| `behavior.auto_cd` | bool | no | `true` | Auto `cd` on jump |
| `behavior.auto_activate` | bool | no | `false` | Auto-activate environment |
| `behavior.track_usage` | bool | no | `true` | Track `last_used` |
| `logging.level` | string | no | `info` | Logging level |
| `logging.format` | string | no | `text` | Logging format |

## 3. `projects.toml`

Project registry.

```toml
# ~/.config/pj/projects.toml

pj_spec_version = "1"

[project.foo]
path = "/home/user/projects/foo"
status = "active"           # active | archived | template
tags = ["go", "cli", "backend"]
description = "Foo CLI tool"
created = "2026-01-15T10:00:00Z"
last_used = "2026-09-11T08:30:00Z"
template = "go-cli"

[project.bar]
path = "/home/user/projects/bar"
status = "archived"
tags = ["rust"]
created = "2025-12-01T00:00:00Z"
last_used = "2026-08-01T12:00:00Z"

```

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `path` | string | yes | Absolute path to project |
| `status` | string | no | `active`, `archived`, `template` |
| `tags` | array[string] | no | Tags for filtering |
| `description` | string | no | Short description |
| `created` | string (RFC3339) | no | Creation timestamp |
| `last_used` | string (RFC3339) | no | Last used timestamp |
| `template` | string | no | Template used |

### Rules

* Key `project.<name>` must be unique.
* Project name: `[a-zA-Z0-9_-]+`, maximum 64 characters.
* `path` must be absolute.
* `path` does not need to exist upon registration (allows moved projects).
* `status` defaults to `active`.

## 4. `.pjrc`

Project-specific configuration. Located at project root.

```toml
# ~/projects/foo/.pjrc

pj_spec_version = "1"

[project]
name = "foo"
type = "go-cli"
description = "Foo CLI tool"

[env]
GOFLAGS = "-mod=vendor"
EDITOR = "nvim"
PROJECT_ROOT = "${PJ_PROJECT_PATH}"

[tasks]
build = "go build ./..."
test = "go test ./..."
lint = "golangci-lint run"
run = "go run ./cmd/foo"
clean = "rm -rf bin/"

[hooks]
before_jump = "echo 'entering foo'"
after_jump = "git status --short"
before_run = "echo 'running task'"
after_run = "echo 'done'"

[shell]
# specific shell initialization for this project
init = "source ./venv/bin/activate"

[template]
source = "go-cli"
version = "1.0.0"

```

### Fields

| Section | Field | Type | Description |
| --- | --- | --- | --- |
| `project` | `name` | string | Project name |
| `project` | `type` | string | Project type |
| `project` | `description` | string | Description |
| `env` | `KEY` | string | Environment variable |
| `tasks` | `name` | string | Task command |
| `hooks` | `before_jump` | string | Hook before jump |
| `hooks` | `after_jump` | string | Hook after jump |
| `hooks` | `before_run` | string | Hook before run |
| `hooks` | `after_run` | string | Hook after run |
| `shell` | `init` | string | Shell init command |
| `template` | `source` | string | Source template |
| `template` | `version` | string | Template version |

### Variable Substitution

Inside `.pjrc`, the following variables are available:

| Variable | Value |
| --- | --- |
| `${PJ_PROJECT_PATH}` | Absolute path to project |
| `${PJ_PROJECT_NAME}` | Project name |
| `${PJ_CONFIG_DIR}` | `$XDG_CONFIG_HOME/pj` |
| `${PJ_DATA_DIR}` | `$XDG_DATA_HOME/pj` |
| `${HOME}` | User home directory |

Substitution is handled by plugins, not the manager.

## 5. Templates

Directory: `$XDG_CONFIG_HOME/pj/templates/<name>/`

Structure:

```text
templates/
  go-cli/
    template.toml       # metadata + variables
    files/              # copied files
      go.mod.tmpl
      main.go.tmpl
      README.md.tmpl
      .pjrc.tmpl
    hooks/
      post-create.sh    # executed after creation

```

### `template.toml`

```toml
# ~/.config/pj/templates/go-cli/template.toml

pj_spec_version = "1"

[template]
name = "go-cli"
version = "1.0.0"
description = "Go CLI project"
author = "John Doe"

[variables]
name = { prompt = "Project name", default = "myapp" }
module = { prompt = "Go module", default = "example.com/myapp" }
author = { prompt = "Author", default = "$USER" }

[files]
# files with .tmpl extension will be rendered
# files without .tmpl will be copied directly
include = ["files/**"]
exclude = ["files/**/*.bak"]

[hooks]
post_create = "hooks/post-create.sh"

```

### Variables

| Field | Type | Description |
| --- | --- | --- |
| `prompt` | string | Prompt displayed to user |
| `default` | string | Default value |
| `required` | bool | Whether value is mandatory |

## 6. TOML Format

All configuration files use **TOML v1.0.0**.

Rationales:

* Human-readable.
* Supports comments.
* Type-safe.
* Parsers available across all programming languages.
* Superior to INI, simpler than YAML.

## 7. File Locking

`projects.toml` may be written by multiple plugins simultaneously. Mandatory locking required.

```text
~/.config/pj/projects.lock

```

Use `flock` (Unix) or `LockFileEx` (Windows).

Rules:

* Lock only during write operations, not during read operations.
* 5-second timeout, then exit with an error.
* Lock is automatically released upon process exit.

## 8. Validation

All configurations are validated upon reading:

* `pj_spec_version` must match.
* Paths must be absolute (except templates).
* Project names must be valid.
* Task names must be valid.
* Hook names must be valid.

On validation failure:

* Output error to stderr.
* Exit code 3 (configuration error).
* Do not modify files.

## 9. Migration

If `pj_spec_version` is older than supported:

* Output a warning.
* Attempt automatic migration if possible.
* Backup original file to `.bak`.
* Write updated version.

If newer:

* Output an error.
* Suggest updating `pj`.
* Exit code 3.

## 10. Complete Example

```toml
# ~/.config/pj/config.toml
pj_spec_version = "1"

[user]
name = "John Doe"
email = "john@example.com"

[defaults]
editor = "nvim"
shell = "bash"

[behavior]
auto_cd = true
track_usage = true

[logging]
level = "info"
format = "text"

```

```toml
# ~/.config/pj/projects.toml
pj_spec_version = "1"

[project.foo]
path = "/home/user/projects/foo"
status = "active"
tags = ["go", "cli"]
created = "2026-01-15T10:00:00Z"
last_used = "2026-09-11T08:30:00Z"

```

```toml
# ~/projects/foo/.pjrc
pj_spec_version = "1"

[project]
name = "foo"
type = "go-cli"

[env]
GOFLAGS = "-mod=vendor"

[tasks]
build = "go build ./..."
test = "go test ./..."

[hooks]
after_jump = "git status --short"

```
