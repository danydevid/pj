package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

const defaultConfigTOML = `# pj — global configuration
# Reference: docs/CONFIG_SPEC.md

pj_spec_version = "1"

[user]
# name  = "Your Name"
# email = "you@example.com"

[defaults]
# editor   = "nvim"
# shell    = "bash"
# template = "go-cli"

[behavior]
auto_cd       = true
auto_activate = false
track_usage   = true

[logging]
level  = "info"   # debug | info | warn | error
format = "text"   # text | json
file   = ""       # empty = stderr
`

const defaultProjectsTOML = `# pj — project registry
# Reference: docs/CONFIG_SPEC.md §3

pj_spec_version = "1"
`

func main() {
	paths := config.GetPaths()
	env := plugin.Load()

	if err := paths.EnsureDirs(); err != nil {
		fatal(err)
	}

	// Extra subdirs that EnsureDirs does not create.
	extraDirs := []string{
		filepath.Join(env.ConfigDir, "templates"),
		filepath.Join(env.DataDir, "plugins"),
		filepath.Join(env.DataDir, "logs"),
	}
	for _, d := range extraDirs {
		if err := os.MkdirAll(d, 0700); err != nil {
			fatal(err)
		}
	}

	created, err := writeIfMissing(paths.GlobalConfigFile(), defaultConfigTOML)
	if err != nil {
		fatal(err)
	}
	report(paths.GlobalConfigFile(), created)

	created, err = writeIfMissing(paths.ProjectsConfigFile(), defaultProjectsTOML)
	if err != nil {
		fatal(err)
	}
	report(paths.ProjectsConfigFile(), created)

	fmt.Printf("pj init: config dir %s\n", env.ConfigDir)
	fmt.Println("pj init: done")
}

func writeIfMissing(path, content string) (bool, error) {
	if _, err := os.Stat(path); err == nil {
		return false, nil
	} else if !os.IsNotExist(err) {
		return false, err
	}
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		return false, err
	}
	return true, nil
}

func report(path string, created bool) {
	if created {
		fmt.Printf("pj init: created %s\n", path)
	} else {
		fmt.Printf("pj init: kept    %s (already exists)\n", path)
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "pj-init: %v\n", err)
	os.Exit(1)
}
