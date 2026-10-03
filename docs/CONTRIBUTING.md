# Contributing to `pj`

## 1. Welcome

Thank you for your interest in contributing to `pj`. This document explains how to contribute.

## 2. Code of Conduct

* Be polite.
* Respect differing opinions.
* Focus on the problem, not the person.
* Help new contributors.

## 3. How to Contribute

### 3.1 Report a Bug

Open an issue on GitHub using the following template:

```text
**Description**: What happened?
**Steps to reproduce**: 1. ... 2. ... 3. ...
**Expected behavior**: What should have happened?
**Environment**: OS, pj version, shell
**Log**: Error output

```

### 3.2 Propose a Feature

Open an issue using the following template:

```text
**Feature**: What is proposed?
**Problem**: What problem does it solve?
**Alternatives**: What other solutions were considered?
**Context**: Who needs it? How often?

```

### 3.3 Submit a Pull Request

1. Fork the repo.
2. Create a branch: `git checkout -b feat/feature-name`.
3. Write code + tests.
4. Update documentation.
5. Commit with clear messages.
6. Push to your fork.
7. Open a PR.

## 4. Development Setup

### 4.1 Prerequisites

* Go 1.22+
* Git
* Make
* Shell: bash, zsh, or fish (for testing)

### 4.2 Clone

```bash
git clone https://github.com/danydevid/pj.git
cd pj

```

### 4.3 Build

```bash
make build

```

Binary will be in `bin/`.

### 4.4 Test

```bash
make test

```

### 4.5 Local Install

```bash
make install

```

Installs to `~/.local/bin/` and `~/.pj/bin/`.

### 4.6 Enable Shell Wrapper

Add to `~/.bashrc`:

```bash
source /path/to/pj/contrib/shell/init.bash

```

Or for fish:

```fish
source /path/to/pj/contrib/shell/init.fish

```

## 5. Project Structure

```text
pj/
  cmd/                # entry points (binary)
  internal/           # private code
  pkg/                # public API (optional)
  contrib/            # non-Go integrations (shell)
  docs/               # documentation & specs
  test/               # integration tests
  scripts/            # build & dev scripts

```

See `docs/ARCHITECTURE.md` for details.

## 6. Adding a Plugin

### 6.1 Go Plugin

1. Create `cmd/pj-foo/main.go`.

```go
package main

import (
    "os"
    "github.com/danydevid/pj/internal/plugin/foo"
)

func main() {
    os.Exit(foo.Run(os.Args[1:]))
}

```

2. Create `internal/plugin/foo/foo.go`.

```go
package foo

func Run(args []string) int {
    // plugin logic
    return 0
}

```

3. Add to `Makefile`:

```makefile
PLUGINS += pj-foo

```

4. Test.

### 6.2 Non-Go Plugin

1. Create a binary `pj-foo` in your choice of language.
2. Place it in `~/.pj/bin/`.
3. Follow `docs/PLUGIN_SPEC.md`.
4. Test manually.

### 6.3 Plugin Checklist

* [ ] Named `pj-<subcommand>`, executable.
* [ ] Handles `--help`.
* [ ] Handles `--json` (if pure).
* [ ] Meaningful exit code (see spec).
* [ ] Errors logged to stderr with prefix `pj-<subcommand>:`.
* [ ] If hybrid: write actions to `$PJ_ACTIONS_FILE`.
* [ ] If hybrid: check `PJ_SHELL`.
* [ ] If hybrid: validate before writing actions.
* [ ] Manifest (optional).
* [ ] Tests included.

## 7. Code Standards

### 7.1 Go

* Follow `gofmt` and `goimports`.
* Run `golangci-lint run`.
* No errors from linter.
* Minimum coverage 70%.
* Comments for public functions.
* Descriptive variable names.

### 7.2 Shell

* Follow `shellcheck`.
* Use `set -euo pipefail`.
* Quote all variables.
* Comments for non-trivial logic.

### 7.3 TOML

* 2-space indentation.
* Comments for non-obvious fields.
* Order fields logically.

## 8. Commit Convention

Format:

```text
<type>(<scope>): <subject>

<body>

<footer>

```

Type:

* `feat`: new feature
* `fix`: bug fix
* `docs`: documentation
* `test`: testing
* `refactor`: refactoring
* `chore`: maintenance
* `perf`: performance

Example:

```text
feat(jump): support project alias

Allow project to have alias via .pjrc:

    [project]
    alias = ["f", "fo"]

Closes #42

```

## 9. Pull Request

### 9.1 Before Submitting

* [ ] Tests pass (`make test`).
* [ ] Linter passes (`make lint`).
* [ ] Build passes (`make build`).
* [ ] Documentation updated.
* [ ] Commit message follows convention.
* [ ] Branch rebased from main.

### 9.2 PR Description

```text
## What
Brief description of changes.

## Why
Reason for changes.

## How
Approach used.

## Test
How to test.

## Checklist
- [x] Tests pass
- [x] Linter passes
- [x] Documentation updated

```

### 9.3 Review

* Minimum 1 approval from a maintainer.
* All CI checks green.
* No conflicts.
* All comments resolved.

## 10. Testing

### 10.1 Unit Tests

Alongside code: `internal/foo/foo_test.go`.

```go
func TestFoo(t *testing.T) {
    got := Foo("input")
    want := "output"
    if got != want {
        t.Errorf("got %q, want %q", got, want)
    }
}

```

### 10.2 Integration Tests

Located in `test/`.

```go
func TestE2EJump(t *testing.T) {
    tmp := t.TempDir()
    // setup
    // run
    // assert
}

```

### 10.3 Shell Testing

Test wrappers in bash, zsh, and fish.

```bash
# manual test
bash -c 'source contrib/shell/init.bash; pj jump foo'
zsh -c 'source contrib/shell/init.zsh; pj jump foo'
fish -c 'source contrib/shell/init.fish; pj jump foo'

```

## 11. Documentation

### 11.1 Specs

Specs located in `docs/`. If behavior changes, update the specs.

* `ARCHITECTURE.md`: architecture
* `CONFIG_SPEC.md`: config format
* `PLUGIN_SPEC.md`: plugin contract
* `ACTIONS_SPEC.md`: actions format
* `XDG_SPEC.md`: file locations
* `CONTRIBUTING.md`: this document

### 11.2 README

Update `README.md` if there are new user-facing features.

### 11.3 Examples

Update `contrib/examples/` if there are new configs or plugins.

## 12. Release

Maintainers handle releases. Process:

1. Update `CHANGELOG.md`.
2. Tag: `git tag v1.0.0`.
3. Push: `git push --tags`.
4. GitHub Actions builds & releases.
5. Update Homebrew, AUR, etc.

## 13. Adding a Spec

If adding a new spec:

1. Create a file in `docs/`.
2. Follow existing format.
3. Update `README.md` with links.
4. Update `CONTRIBUTING.md` with references.
5. Discuss in an issue first.

## 14. Ethics

* Do not spam.
* Do not repeatedly request reviews.
* Do not get upset if a PR is rejected.
* Do not commit secrets (API keys, passwords).
* Do not commit large files (> 1 MB).

## 15. Getting Help

* GitHub Issues: for bugs & features.
* GitHub Discussions: for questions.
* Discord/Matrix: for chat (if available).

## 16. License

By contributing, you agree that your contributions will be licensed under the project's license (MIT/Apache-2.0).

## 17. Thank You

Every contribution is valuable. Thank you for taking the time.
