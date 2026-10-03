package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/BurntSushi/toml"

	"github.com/danydevid/pj/internal/config"
	"github.com/danydevid/pj/internal/plugin"
)

type jsonProject struct {
	Name     string   `json:"name"`
	Path     string   `json:"path"`
	Status   string   `json:"status"`
	Tags     []string `json:"tags,omitempty"`
	LastUsed string   `json:"last_used,omitempty"`
}

func main() {
	fs := flag.NewFlagSet("pj-list", flag.ExitOnError)
	jsonOut := fs.Bool("json", false, "output as JSON")
	all := fs.Bool("all", false, "include archived projects")
	_ = fs.Parse(os.Args[1:])

	env := plugin.Load()
	reg, err := loadRegistry(env.ConfigDir)
	if err != nil {
		fatal(err)
	}

	names := make([]string, 0, len(reg.Projects))
	for name, item := range reg.Projects {
		if !*all && item.Status == "archived" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)

	if len(names) == 0 {
		if *jsonOut {
			fmt.Println(`{"projects":[]}`)
		} else {
			fmt.Fprintln(os.Stderr, "pj-list: no projects registered")
		}
		return
	}

	if *jsonOut {
		out := make([]jsonProject, 0, len(names))
		for _, name := range names {
			item := reg.Projects[name]
			jp := jsonProject{
				Name:   name,
				Path:   item.Path,
				Status: normalStatus(item.Status),
				Tags:   item.Tags,
			}
			if !item.LastUsed.IsZero() {
				jp.LastUsed = item.LastUsed.Format(time.RFC3339)
			}
			out = append(out, jp)
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(map[string]any{"projects": out}); err != nil {
			fatal(err)
		}
		return
	}

	tw := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(tw, "NAME\tSTATUS\tPATH")
	for _, name := range names {
		item := reg.Projects[name]
		fmt.Fprintf(tw, "%s\t%s\t%s\n", name, normalStatus(item.Status), item.Path)
	}
	tw.Flush()
}

func normalStatus(s string) string {
	if s == "" {
		return "active"
	}
	return s
}

func loadRegistry(configDir string) (*config.ProjectsRegistry, error) {
	path := filepath.Join(configDir, "projects.toml")
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return emptyRegistry(), nil
		}
		return nil, fmt.Errorf("cannot read %s: %w", path, err)
	}

	var reg config.ProjectsRegistry
	if err := toml.Unmarshal(data, &reg); err != nil {
		return nil, fmt.Errorf("%w: cannot parse %s: %v", config.ErrConfigInvalid, path, err)
	}
	if reg.PJSpecVersion == "" {
		reg.PJSpecVersion = config.ExpectedSpecVersion
	}
	if err := reg.Validate(); err != nil {
		return nil, err
	}
	return &reg, nil
}

func emptyRegistry() *config.ProjectsRegistry {
	return &config.ProjectsRegistry{
		PJSpecVersion: config.ExpectedSpecVersion,
		Projects:      map[string]config.ProjectItem{},
	}
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "pj-list: %v\n", err)
	os.Exit(1)
}
