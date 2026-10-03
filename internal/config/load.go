package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// LoadProjectsRegistry reads projects.toml from configDir.
//
// Returns an empty but valid registry when the file does not exist, so
// plugins can treat "no projects" and "no file" uniformly.
// Wraps ErrConfigInvalid for parse or validation failures.
func LoadProjectsRegistry(configDir string) (*ProjectsRegistry, error) {
	path := filepath.Join(configDir, "projects.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &ProjectsRegistry{
				PJSpecVersion: ExpectedSpecVersion,
				Projects:      map[string]ProjectItem{},
			}, nil
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var reg ProjectsRegistry
	if err := toml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("%w: cannot parse %s: %v", ErrConfigInvalid, path, err)
	}
	if reg.PJSpecVersion == "" {
		reg.PJSpecVersion = ExpectedSpecVersion
	}
	if err := reg.Validate(); err != nil {
		return nil, err
	}
	return &reg, nil
}
