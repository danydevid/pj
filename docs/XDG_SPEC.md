# XDG Spec `pj`

## 1. Summary

`pj` follows the **XDG Base Directory Specification** for all user files. This ensures:

* Linux standard compatibility.
* Easy to backup.
* Easy to sync.
* Prevents cluttering the home directory.
* Can be overridden via environment variables.

## 2. Environment Variables

| Variable | Default | Used for |
| --- | --- | --- |
| `XDG_CONFIG_HOME` | `~/.config` | Config, registry, templates |
| `XDG_DATA_HOME` | `~/.local/share` | Data, user plugins |
| `XDG_CACHE_HOME` | `~/.cache` | Cache, index |
| `XDG_RUNTIME_DIR` | `/run/user/$UID` | Temporary files, sockets |
| `XDG_STATE_HOME` | `~/.local/state` | State, logs, history |

If an environment variable is empty or unset, use the default.

If `XDG_RUNTIME_DIR` is not set, fallback to `/tmp` with a `pj-$UID` subdirectory.

## 3. Directory Structure

### 3.1 Config: `$XDG_CONFIG_HOME/pj/`

```text
~/.config/pj/
  config.toml              # user global config
  projects.toml            # project registry
  projects.lock            # lock file for registry
  templates/               # project templates
    go-cli/
      template.toml
      files/
      hooks/
    rust-cli/
      ...
  init.bash                # shell wrapper (optional)
  init.zsh
  init.fish
  completions/
    pj.bash
    _pj
    pj.fish

```

**Contents:** configurations that the user **edits** or **wants to backup**.

### 3.2 Data: `$XDG_DATA_HOME/pj/`

```text
~/.local/share/pj/
  plugins/                 # plugins installed via pj-plugin-install
    pj-docker
    pj-git
  logs/                    # logs (if any)
  history                  # history (optional)

```

**Contents:** data **generated** by `pj`, not edited by the user.

### 3.3 Cache: `$XDG_CACHE_HOME/pj/`

```text
~/.cache/pj/
  projects.idx             # project index for completion
  completion.cache         # completion cache
  template.cache           # template cache

```

**Contents:** data that **can be deleted** at any time. If missing, `pj` regenerates it.

### 3.4 Runtime: `$XDG_RUNTIME_DIR/pj/`

```text
/run/user/1000/pj/
  pj-actions-xxxxx         # temporary action files

```

**Contents:** temporary files that are **lost** upon logout.

### 3.5 State: `$XDG_STATE_HOME/pj/`

```text
~/.local/state/pj/
  history                  # command history
  session                  # session state
  recent                   # recent projects

```

**Contents:** **persistent** state that is not configuration.

### 3.6 User Plugins: `~/.pj/bin/`

```text
~/.pj/bin/
  pj-jump
  pj-list
  pj-init
  ...

```

**Contents:** plugin binaries placed manually by the user. This is not XDG, but a `pj` convention.

Reason: `~/.pj/bin/` is easier to remember and `ls` than `$XDG_DATA_HOME/pj/bin/`. Plugins can also be placed in XDG data.

### 3.7 Per-Project: `<project-path>/.pjrc`

```text
~/projects/foo/.pjrc

```

**Contents:** project-specific config. Located at the project root, typically committed to Git.

## 4. Search Priority

### 4.1 Plugins

In order of priority:

1. `$PJ_PLUGIN_DIR` (override)
2. `~/.pj/bin/`
3. `$XDG_DATA_HOME/pj/plugins/`
4. `$XDG_DATA_HOME/pj/bin/`
5. `/usr/local/lib/pj/plugins/`
6. `/usr/lib/pj/plugins/`
7. `$PATH` (fallback)

If a plugin exists in multiple locations, the first one takes precedence.

### 4.2 Config

1. `$PJ_CONFIG_DIR` (override)
2. `$XDG_CONFIG_HOME/pj/`

### 4.3 Templates

1. `<project-path>/.pj/templates/` (project-local)
2. `$XDG_CONFIG_HOME/pj/templates/` (user)
3. `/usr/local/share/pj/templates/` (system)
4. `/usr/share/pj/templates/` (distro)

## 5. Overriding via Environment

Users can override all paths:

```bash
export PJ_CONFIG_DIR=/custom/config
export PJ_DATA_DIR=/custom/data
export PJ_CACHE_DIR=/custom/cache
export PJ_RUNTIME_DIR=/custom/runtime
export PJ_PLUGIN_DIR=/custom/plugins

```

If set, `pj` uses those values instead of XDG defaults.

## 6. Initialization

`pj init` creates the following structure:

```text
~/.config/pj/
  config.toml
  projects.toml
  templates/
  init.bash
  init.zsh
  init.fish
~/.local/share/pj/
  plugins/
  logs/
~/.cache/pj/
~/.pj/bin/

```

`pj init` does **not** overwrite existing files. If files already exist, it prints a warning.

## 7. Backup and Sync

Items to backup:

* `$XDG_CONFIG_HOME/pj/` (configs, registry, templates)
* `~/.pj/bin/` (user plugins)

Items **not** to backup:

* `$XDG_CACHE_HOME/pj/` (can be regenerated)
* `$XDG_RUNTIME_DIR/pj/` (temporary)
* `$XDG_DATA_HOME/pj/logs/` (logs)

Example sync via Git:

```bash
cd ~/.config/pj
git init
git add .
git commit -m "pj config"
git remote add origin git@github.com:user/pj-config.git
git push

```

## 8. Migration

### 8.1 From Legacy Locations

If `pj` previously stored files in non-XDG locations, `pj init` can migrate:

* `~/.pj/` → `$XDG_CONFIG_HOME/pj/`
* `~/.pjrc` → `<project>/.pjrc`
* `~/.pj/templates/` → `$XDG_CONFIG_HOME/pj/templates/`

Migration is **not automatic**. The user must run `pj migrate`.

### 8.2 To New Locations

If the specification changes, `pj` can perform migration:

* Backup old files.
* Write new files.
* Print a summary.

## 9. Windows and macOS

### 9.1 macOS

macOS does not natively have XDG, but many tools follow it. `pj` follows it as well.

Defaults on macOS:

* `XDG_CONFIG_HOME` = `~/.config`
* `XDG_DATA_HOME` = `~/.local/share`
* `XDG_CACHE_HOME` = `~/.cache`

These are **not** native macOS locations (`~/Library/Application Support`), but they maintain consistency with other tools.

Users who prefer native paths can override them:

```bash
export XDG_CONFIG_HOME="$HOME/Library/Application Support"
export XDG_DATA_HOME="$HOME/Library/Application Support"
export XDG_CACHE_HOME="$HOME/Library/Caches"

```

### 9.2 Windows

Windows does not have XDG. `pj` uses:

* `XDG_CONFIG_HOME` = `%APPDATA%\pj`
* `XDG_DATA_HOME` = `%LOCALAPPDATA%\pj`
* `XDG_CACHE_HOME` = `%LOCALAPPDATA%\pj\cache`

However, `pj` focuses primarily on Unix-like systems (Linux, macOS, BSD).

## 10. Permissions

| Directory | Permissions |
| --- | --- |
| `$XDG_CONFIG_HOME/pj/` | `0700` |
| `$XDG_DATA_HOME/pj/` | `0700` |
| `$XDG_CACHE_HOME/pj/` | `0700` |
| `$XDG_RUNTIME_DIR/pj/` | `0700` |
| `~/.pj/bin/` | `0700` |
| Config files | `0600` |
| Plugin files | `0755` |

## 11. Complete Example

```text
~/
  .config/
    pj/
      config.toml
      projects.toml
      templates/
        go-cli/
      init.bash
      init.zsh
      init.fish
  .local/
    share/
      pj/
        plugins/
          pj-docker
        logs/
    state/
      pj/
        history
        recent
  .cache/
    pj/
      projects.idx
  .pj/
    bin/
      pj-jump
      pj-list
      pj-init
      pj-docker

```

## 12. Checklist

* [ ] All configs in `$XDG_CONFIG_HOME/pj/`.
* [ ] All data in `$XDG_DATA_HOME/pj/`.
* [ ] All caches in `$XDG_CACHE_HOME/pj/`.
* [ ] All temporary files in `$XDG_RUNTIME_DIR/pj/`.
* [ ] All state files in `$XDG_STATE_HOME/pj/`.
* [ ] User plugins in `~/.pj/bin/`.
* [ ] Project configs in `<project>/.pjrc`.
* [ ] All paths overridable via `PJ_*`.
* [ ] `pj init` creates the complete directory structure.
* [ ] `pj migrate` available for migration.
