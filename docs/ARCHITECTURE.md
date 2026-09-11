# Arsitektur `pj`

## 1. Summary

`pj` is an extensible **project jumper**. Its architecture follows a **microkernel + plugin** pattern with a **shell adapter** for terminal integration.

Core principles:

> **Mechanism in manager, policy in plugin, integration in adapter, state in filesystem.**

## 2. Diagram

```text
    ┌──────────────────────────────────────────────────┐
    │                  USER INTERFACE                  │
    │                                                  │
    │   Terminal      GUI         Script      Remote   │
    │  (bash/zsh)   (GTK/Qt)    (Python)      (SSH)    │
    └──────┬───────────┬───────────┬───────────┬───────┘
           │           │           │           │
           │ fork+exec │           │           │
           ▼           ▼           ▼           ▼
    ┌──────────────────────────────────────────────────┐
    │              SHELL WRAPPER (adapter)             │
    │                                                  │
    │  • create tmpfile actions                        │
    │  • set PJ_ACTIONS_FILE, PJ_SHELL                 │
    │  • call binary pj                                │
    │  • source actions file                           │
    │  • cleanup tmpfile                               │
    │                                                  │
    │  NO business logic here.                         │
    └──────────────────────┬───────────────────────────┘
                           │ fork+exec
                           ▼
    ┌──────────────────────────────────────────────────┐
    │          MANAGER (microkernel, binary pj)        │
    │                                                  │
    │  • parse args                                    │
    │  • built-in: --help, --version, --list, --which  │
    │  • resolve plugin in ~/.pj/bin/                  │
    │  • setup env (PJ_*)                              │
    │  • fork+exec plugin                              │
    │  • waitpid, return exit code                     │
    │                                                  │
    │  NO business logic.                              │
    │  NO state.                                       │
    │  NO config parsing.                              │
    └──────────────────────┬───────────────────────────┘
                           │ fork+exec
                           ▼
    ┌──────────────────────────────────────────────────┐
    │             PLUGIN (binary in ~/.pj/bin/)        │
    │                                                  │
    │    pj-jump   pj-list   pj-init   pj-run   ...    │
    │                                                  │
    │  • read config                                   │
    │  • business logic                                │
    │  • print human output to stdout                  │
    │  • write actions to $PJ_ACTIONS_FILE (hybrid)    │
    │  • exit with meaningful code                     │
    └──────────────────────┬───────────────────────────┘
                           │ read/write
                           ▼
    ┌──────────────────────────────────────────────────┐
    │              STATE (filesystem, XDG)             │
    │                                                  │
    │  ~/.config/pj/       ~/.local/share/pj/          │
    │  ~/.cache/pj/        ~/.pj/bin/                  │
    │  ~/projects/*/.pjrc                              │
    └──────────────────────────────────────────────────┘

```

## 3. Components

### 3.1 Manager (`pj`)

Go binary, statically linked, ~10 MB. Its only responsibility is dispatching.

**Workload:**

* Parse arguments.
* Handle minimal built-ins (`--help`, `--version`, `--list`, `--which`).
* Resolve plugins: scan `~/.pj/bin/`, `$XDG_DATA_HOME/pj/bin/`, `$PATH`.
* Setup environment `PJ_*`.
* `fork+exec` plugin.
* Passthrough stdin/stdout/stderr.
* `waitpid`, return exit code.

**Constraints:**

* Maximum 500 lines of Go code.
* Must not import `internal/config`.
* Must not parse configurations.
* Must not store state.
* Must not be aware of specific plugins.

### 3.2 Plugin

Executable binary located in `~/.pj/bin/pj-<name>`. Can be written in any programming language.

**Contract:**

* Name: `pj-<subcommand>`, executable.
* Arguments: passed from manager.
* Environment: `PJ_ACTIONS_FILE`, `PJ_SHELL`, `PJ_CONFIG_DIR`, `PJ_VERSION`.
* Output: human-readable to stdout, errors to stderr.
* Shell actions: written to `$PJ_ACTIONS_FILE` (hybrid plugins only).
* Exit code: 0 on success, non-zero on failure.

### 3.3 Shell Wrapper

Shell function sourced from `~/.bashrc` or `~/.config/fish/config.fish`.

**Workload:**

* Create actions tmpfile.
* Set `PJ_ACTIONS_FILE`, `PJ_SHELL`.
* Call `pj` binary.
* `source` actions file in parent shell context.
* Cleanup tmpfile.

**Constraints:**

* Maximum 50 lines of shell code.
* Must not know subcommand names.
* Must not parse configurations.
* Must not store state.

### 3.4 Actions File

Temporary file containing shell code generated by hybrid plugins.

Format: valid shell code for `PJ_SHELL`.

Example:

```sh
cd '/home/user/projects/foo'
export FOO='bar'
unset BAZ

```

## 4. Execution Flow

```text
User: pj jump foo
  │
  ▼ Shell wrapper
  ├─ mktemp → /tmp/pj-actions-xxx
  ├─ set PJ_ACTIONS_FILE=/tmp/pj-actions-xxx
  ├─ set PJ_SHELL=bash
  └─ fork+exec pj jump foo
        │
        ▼ Manager
        ├─ parse args
        ├─ resolve plugin → ~/.pj/bin/pj-jump
        ├─ setup env
        └─ fork+exec pj-jump foo
              │
              ▼ Plugin pj-jump
              ├─ read ~/.config/pj/projects.toml
              ├─ read ~/projects/foo/.pjrc
              ├─ validate path
              ├─ print "Jumping to foo..."
              ├─ write actions to $PJ_ACTIONS_FILE
              └─ exit 0
              │
        ◄─────┘
        │
  ◄─────┘
  ├─ source /tmp/pj-actions-xxx
  │   → cd '/home/user/projects/foo'
  │   → export FOO='bar'
  ├─ rm -f /tmp/pj-actions-xxx
  └─ return 0

```

## 5. Boundary Responsibilities

| Component | Responsibility | Prohibited |
| --- | --- | --- |
| Shell wrapper | Execute shell actions | Business logic, config, state |
| Manager | Dispatch, resolve, exec | Config, business logic, state |
| Plugin | Business logic, config | Dispatching other plugins |
| Actions file | Shell action channel | Human output |
| Filesystem | State, config, registry | (none) |

## 6. Design Patterns

| Pattern | Location | Function |
| --- | --- | --- |
| Microkernel | Manager | Dispatch to plugins |
| Ports & Adapters | Overall architecture | Plugins & wrappers as adapters |
| Command Dispatcher | Manager | `pj <cmd>` → resolve plugin |
| Plugin | Plugin layer | Features as binaries |
| Adapter | Shell wrapper | Translate actions to shell |
| Repository | `internal/config/` | Read/write state |
| Strategy | `internal/cli/resolve.go` | Plugin discovery logic |

## 7. Architectural Decisions and Rationales

### 7.1 Binary, not shared object

**Rationale:**

* Shells cannot `dlopen`.
* Go runtime inside `.so` has issues with `fork()`.
* ABI is a permanent maintenance burden.
* Binary plugins are already language-agnostic.

**Revisal condition:** Actual requirement for embedding emerges.

### 7.2 Microkernel, not core+plugin

**Rationale:**

* No core is truly stable.
* A thin manager will not bloated over time.
* Maximum extensibility.

**Revisal condition:** Truly shared core logic emerges.

### 7.3 Shell wrapper for actions

**Rationale:**

* A child process cannot modify its parent environment (Unix law).
* Wrapper acts as a driving adapter.

**Revisal condition:** Never.

### 7.4 Actions file, not eval/fd3

**Rationale:**

* Clean separation of human output vs shell actions.
* Streaming continues to work seamlessly.
* Portable across all shells.
* Easily debuggable.

**Revisal condition:** Never.

### 7.5 Filesystem state, not daemon

**Rationale:**

* No real-time state required.
* Files are transparent: back up, edit, and sync easily.
* File locks are sufficient for race conditions.

**Revisal condition:** Real-time synchronization is required.

### 7.6 Directory scan, not registry

**Rationale:**

* The filesystem is already a registry.
* Directory scanning is fast (< 2 ms for 70 plugins).
* New plugins are immediately detected.

**Revisal condition:** Plugin count exceeds 1000.

## 8. Design Principles

1. **Mechanism in manager, policy in plugin.**
2. **State in files, not in processes.**
3. **Adapters for integration.**
4. **Constrain what can grow.**
5. **Start simple, add only when necessary.**
6. **Follow the Unix philosophy.**

## 9. Non-Goals in v1

* Shared objects.
* Daemons.
* Plugin registries.
* SQLite.
* Event bus.
* Worker pools.
* GUI.
* HTTP API.
* Plugin sandboxing.
* Multi-user support.

Every item listed above represents unnecessary complexity at this stage. Add only when supported by concrete evidence.
