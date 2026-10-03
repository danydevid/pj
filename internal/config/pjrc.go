package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// PJRCFilename is the standard name of the per-project configuration file.
// It is expected to live at the project root (see CONFIG_SPEC.md §4).
const PJRCFilename = ".pjrc"

// ErrPJRCNotFound is returned when no .pjrc file exists at the expected location.
var ErrPJRCNotFound = errors.New(".pjrc not found")

// LoadPJRC reads and validates the `.pjrc` file located at `projectPath/.pjrc`.
//
// The projectPath must be an absolute path to an existing directory.
// Returns ErrPJRCNotFound if the file does not exist.
// Returns an error wrapping ErrConfigInvalid if the file exists but is
// unparseable or fails validation.
func LoadPJRC(projectPath string) (*PJRC, error) {
	if !filepath.IsAbs(projectPath) {
		return nil, fmt.Errorf("%w: project path must be absolute: %s", ErrConfigInvalid, projectPath)
	}

	info, err := os.Stat(projectPath)
	if err != nil {
		return nil, fmt.Errorf("cannot access project path: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("%w: not a directory: %s", ErrConfigInvalid, projectPath)
	}

	return LoadPJRCFile(filepath.Join(projectPath, PJRCFilename))
}

// LoadPJRCFile reads and validates a `.pjrc` file at an explicit path.
//
// Errors:
//   - ErrPJRCNotFound        : file does not exist
//   - ErrConfigInvalid (wrap): parse error, validation error
//   - other                  : I/O errors
func LoadPJRCFile(path string) (*PJRC, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s", ErrPJRCNotFound, path)
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var pjrc PJRC
	if err := toml.Unmarshal(data, &pjrc); err != nil {
		return nil, fmt.Errorf("%w: cannot parse %s: %v", ErrConfigInvalid, path, err)
	}

	pjrc.ApplyDefaults()

	if err := pjrc.Validate(); err != nil {
		return nil, err
	}

	return &pjrc, nil
}

// FindPJRC walks upward from startPath looking for the first `.pjrc` file.
// It returns the absolute path to the discovered file, or ErrPJRCNotFound
// if the filesystem root is reached without finding one.
//
// This is a convenience helper for tools that want "search upwards" semantics
// similar to git or direnv; the canonical contract is still that `.pjrc`
// lives at the project root.
func FindPJRC(startPath string) (string, error) {
	dir, err := filepath.Abs(startPath)
	if err != nil {
		return "", err
	}

	// If startPath is a file, begin from its directory.
	if info, err := os.Stat(dir); err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	}

	for {
		candidate := filepath.Join(dir, PJRCFilename)
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("%w: no .pjrc found in %s or its parents", ErrPJRCNotFound, startPath)
}

// ApplyDefaults fills in fields that were left unset in the `.pjrc` file.
//
// Currently only the spec version is defaulted. Other fields (project name,
// env, tasks, hooks) have no meaningful defaults at this layer — they are
// optional and consumed as-is by plugins.
func (p *PJRC) ApplyDefaults() {
	if p.PJSpecVersion == "" {
		p.PJSpecVersion = ExpectedSpecVersion
	}
}
