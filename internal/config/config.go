package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"
)

const ExpectedSpecVersion = "1"

var (
	projectNameRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]+$`)
	ErrConfigInvalid = errors.New("configuration validation failed")
)

// ==========================================
// 1. GLOBAL CONFIG (config.toml)
// ==========================================

type GlobalConfig struct {
	PJSpecVersion string         `toml:"pj_spec_version"`
	User          UserConfig     `toml:"user"`
	Defaults      DefaultsConfig `toml:"defaults"`
	Paths         PathsConfig    `toml:"paths"`
	Behavior      BehaviorConfig `toml:"behavior"`
	Logging       LoggingConfig  `toml:"logging"`
}

type UserConfig struct {
	Name  string `toml:"name"`
	Email string `toml:"email"`
}

type DefaultsConfig struct {
	Editor   string `toml:"editor"`
	Shell    string `toml:"shell"`
	Template string `toml:"template"`
}

type PathsConfig struct {
	Projects  string `toml:"projects"`
	Templates string `toml:"templates"`
}

type BehaviorConfig struct {
	AutoCD       *bool `toml:"auto_cd"`
	AutoActivate *bool `toml:"auto_activate"`
	TrackUsage   *bool `toml:"track_usage"`
}

type LoggingConfig struct {
	Level  string `toml:"level"`
	Format string `toml:"format"`
	File   string `toml:"file"`
}

func (c *GlobalConfig) ApplyDefaults() {
	if c.PJSpecVersion == "" {
		c.PJSpecVersion = ExpectedSpecVersion
	}
	if c.User.Name == "" {
		c.User.Name = os.Getenv("USER")
	}
	if c.Defaults.Editor == "" {
		c.Defaults.Editor = os.Getenv("EDITOR")
	}
	if c.Defaults.Shell == "" {
		c.Defaults.Shell = os.Getenv("SHELL")
	}

	trueVal := true
	falseVal := false

	if c.Behavior.AutoCD == nil {
		c.Behavior.AutoCD = &trueVal
	}
	if c.Behavior.AutoActivate == nil {
		c.Behavior.AutoActivate = &falseVal
	}
	if c.Behavior.TrackUsage == nil {
		c.Behavior.TrackUsage = &trueVal
	}

	if c.Logging.Level == "" {
		c.Logging.Level = "info"
	}
	if c.Logging.Format == "" {
		c.Logging.Format = "text"
	}
}

func (c *GlobalConfig) Validate() error {
	if c.PJSpecVersion != ExpectedSpecVersion {
		return fmt.Errorf("%w: unsupported pj_spec_version '%s'", ErrConfigInvalid, c.PJSpecVersion)
	}

	switch c.Logging.Level {
	case "debug", "info", "warn", "error":
	default:
		return fmt.Errorf("%w: invalid logging.level '%s'", ErrConfigInvalid, c.Logging.Level)
	}

	switch c.Logging.Format {
	case "text", "json":
	default:
		return fmt.Errorf("%w: invalid logging.format '%s'", ErrConfigInvalid, c.Logging.Format)
	}

	return nil
}

// ==========================================
// 2. PROJECT REGISTRY (projects.toml)
// ==========================================

type ProjectsRegistry struct {
	PJSpecVersion string                 `toml:"pj_spec_version"`
	Projects      map[string]ProjectItem `toml:"project"`
}

type ProjectItem struct {
	Path        string    `toml:"path"`
	Status      string    `toml:"status"`
	Tags        []string  `toml:"tags"`
	Description string    `toml:"description"`
	Created     time.Time `toml:"created"`
	LastUsed    time.Time `toml:"last_used"`
	Template    string    `toml:"template"`
}

func (p *ProjectsRegistry) Validate() error {
	if p.PJSpecVersion != ExpectedSpecVersion {
		return fmt.Errorf("%w: unsupported pj_spec_version '%s'", ErrConfigInvalid, p.PJSpecVersion)
	}

	for name, item := range p.Projects {
		if len(name) > 64 || !projectNameRegex.MatchString(name) {
			return fmt.Errorf("%w: invalid project name '%s'", ErrConfigInvalid, name)
		}

		if !filepath.IsAbs(item.Path) {
			return fmt.Errorf("%w: project path for '%s' must be absolute: %s", ErrConfigInvalid, name, item.Path)
		}

		status := item.Status
		if status == "" {
			status = "active"
		}

		if status != "active" && status != "archived" && status != "template" {
			return fmt.Errorf("%w: invalid status '%s' for project '%s'", ErrConfigInvalid, item.Status, name)
		}
	}

	return nil
}

// ==========================================
// 3. PER-PROJECT CONFIG (.pjrc)
// ==========================================

type PJRC struct {
	PJSpecVersion string            `toml:"pj_spec_version"`
	Project       PJRCProject       `toml:"project"`
	Env           map[string]string `toml:"env"`
	Tasks         map[string]string `toml:"tasks"`
	Hooks         JunctionHooks     `toml:"hooks"`
	Shell         PJRCShell         `toml:"shell"`
	Template      JunctionTemplate  `toml:"template"`
}

type PJRCProject struct {
	Name        string `toml:"name"`
	Type        string `toml:"type"`
	Description string `toml:"description"`
}

type JunctionHooks struct {
	BeforeJump string `toml:"before_jump"`
	AfterJump  string `toml:"after_jump"`
	BeforeRun  string `toml:"before_run"`
	AfterRun   string `toml:"after_run"`
}

type PJRCShell struct {
	Init string `toml:"init"`
}

type JunctionTemplate struct {
	Source  string `toml:"source"`
	Version string `toml:"version"`
}

func (p *PJRC) Validate() error {
	if p.PJSpecVersion != ExpectedSpecVersion {
		return fmt.Errorf("%w: unsupported pj_spec_version '%s'", ErrConfigInvalid, p.PJSpecVersion)
	}
	return nil
}

// ==========================================
// 4. TEMPLATE METADATA (template.toml)
// ==========================================

type TemplateManifest struct {
	PJSpecVersion string                      `toml:"pj_spec_version"`
	Template      TemplateMeta                `toml:"template"`
	Variables     map[string]TemplateVariable `toml:"variables"`
	Files         TemplateFiles               `toml:"files"`
	Hooks         TemplateHooks               `toml:"hooks"`
}

type TemplateMeta struct {
	Name        string `toml:"name"`
	Version     string `toml:"version"`
	Description string `toml:"description"`
	Author      string `toml:"author"`
}

type TemplateVariable struct {
	Prompt   string `toml:"prompt"`
	Default  string `toml:"default"`
	Required bool   `toml:"required"`
}

type TemplateFiles struct {
	Include []string `toml:"include"`
	Exclude []string `toml:"exclude"`
}

type TemplateHooks struct {
	PostCreate string `toml:"post_create"`
}

func (t *TemplateManifest) Validate() error {
	if t.PJSpecVersion != ExpectedSpecVersion {
		return fmt.Errorf("%w: unsupported pj_spec_version '%s'", ErrConfigInvalid, t.PJSpecVersion)
	}
	return nil
}
