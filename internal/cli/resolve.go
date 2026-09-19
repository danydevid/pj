package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

func ResolvePlugin(subcommand string) (string, error) {
	binName := "pj-" + subcommand
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	var searchDirs []string

	if customDir := os.Getenv("PJ_PLUGIN_DIR"); customDir != "" {
		searchDirs = append(searchDirs, customDir)
	}

	if home, err := os.UserHomeDir(); err == nil {
		searchDirs = append(searchDirs, filepath.Join(home, ".pj", "bin"))
	}

	xdgData := os.Getenv("XDG_DATA_HOME")
	if xdgData == "" {
		if home, err := os.UserHomeDir(); err == nil {
			xdgData = filepath.Join(home, ".local", "share")
		}
	}
	if xdgData != "" {
		searchDirs = append(searchDirs,
			filepath.Join(xdgData, "pj", "plugins"),
			filepath.Join(xdgData, "pj", "bin"),
		)
	}

	for _, dir := range searchDirs {
		candidate := filepath.Join(dir, binName)
		if isExecutable(candidate) {
			return candidate, nil
		}
	}

	if pathInSystem, err := exec.LookPath(binName); err == nil {
		return pathInSystem, nil
	}

	return "", fmt.Errorf("pj: plugin '%s' not found (exit status 127)", subcommand)
}

func isExecutable(path string) bool {
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	return info.Mode()&0111 != 0
}
