package config

import (
	"os"
	"path/filepath"
)

// Paths holds all directory and file paths used by pj.
type Paths struct {
	ConfigDir  string
	DataDir    string
	CacheDir   string
	RuntimeDir string
	StateDir   string
	UserBinDir string
}

// GetPaths returns a Paths struct taking into account:
// 1. PJ_* environment variable overrides (highest priority)
// 2. XDG Base Directory Specification
// 3. Fallback defaults if XDG variables are unset
func GetPaths() *Paths {
	homeDir, _ := os.UserHomeDir()

	return &Paths{
		ConfigDir:  getEnvOrDefault("PJ_CONFIG_DIR", getXDGDir("XDG_CONFIG_HOME", filepath.Join(homeDir, ".config"), "pj")),
		DataDir:    getEnvOrDefault("PJ_DATA_DIR", getXDGDir("XDG_DATA_HOME", filepath.Join(homeDir, ".local", "share"), "pj")),
		CacheDir:   getEnvOrDefault("PJ_CACHE_DIR", getXDGDir("XDG_CACHE_HOME", filepath.Join(homeDir, ".cache"), "pj")),
		RuntimeDir: getRuntimeDir(homeDir),
		StateDir:   getXDGDir("XDG_STATE_HOME", filepath.Join(homeDir, ".local", "state"), "pj"),
		UserBinDir: filepath.Join(homeDir, ".pj", "bin"),
	}
}

// GlobalConfigFile returns the full path to config.toml (~/.config/pj/config.toml).
func (p *Paths) GlobalConfigFile() string {
	return filepath.Join(p.ConfigDir, "config.toml")
}

// ProjectsConfigFile returns the full path to projects.toml (~/.config/pj/projects.toml).
func (p *Paths) ProjectsConfigFile() string {
	return filepath.Join(p.ConfigDir, "projects.toml")
}

// ProjectsLockFile returns the full path to projects.lock (~/.config/pj/projects.lock).
func (p *Paths) ProjectsLockFile() string {
	return filepath.Join(p.ConfigDir, "projects.lock")
}

// EnsureDirs creates all required directories on the file system if they do not exist.
func (p *Paths) EnsureDirs() error {
	dirs := []string{
		p.ConfigDir,
		p.DataDir,
		p.CacheDir,
		p.RuntimeDir,
		p.StateDir,
		p.UserBinDir,
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0700); err != nil {
			return err
		}
	}
	return nil
}

// Helper to retrieve an environment variable value or return a default.
func getEnvOrDefault(envKey, defaultValue string) string {
	if val := os.Getenv(envKey); val != "" {
		return val
	}
	return defaultValue
}

// Helper to construct XDG-based directory paths.
func getXDGDir(xdgEnv, defaultBaseDir, subDir string) string {
	baseDir := os.Getenv(xdgEnv)
	if baseDir == "" {
		baseDir = defaultBaseDir
	}
	return filepath.Join(baseDir, subDir)
}

// Helper to determine Runtime Directory (XDG_RUNTIME_DIR or fallback to /tmp/pj-$USER).
func getRuntimeDir(homeDir string) string {
	if val := os.Getenv("PJ_RUNTIME_DIR"); val != "" {
		return val
	}
	if runtime := os.Getenv("XDG_RUNTIME_DIR"); runtime != "" {
		return filepath.Join(runtime, "pj")
	}

	// Fallback to /tmp with a pj-<USER> directory name
	user := os.Getenv("USER")
	if user == "" {
		user = "user"
	}
	return filepath.Join(os.TempDir(), "pj-"+user)
}
