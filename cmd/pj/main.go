package main

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	version   = "dev"
	commit    = "unknown"
	buildDate = "unknown"
)

const pluginPrefix = "pj-"

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		usage(os.Stderr)
		os.Exit(2)
	}

	switch args[0] {
	case "-h", "--help", "help":
		usage(os.Stdout)
		os.Exit(0)
	case "-v", "--version", "version":
		fmt.Printf("pj %s (commit %s, built %s)\n", version, commit, buildDate)
		os.Exit(0)
	}

	sub := args[0]
	rest := args[1:]

	binary, err := resolvePlugin(sub)
	if err != nil {
		fmt.Fprintf(os.Stderr, "pj: %v\n", err)
		os.Exit(127)
	}

	os.Exit(run(binary, rest))
}

func run(binary string, args []string) int {
	self, err := os.Executable()
	if err != nil || self == "" {
		self = os.Args[0]
	}

	cmd := exec.Command(binary, args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = append(os.Environ(),
		"PJ_MANAGER="+self,
		"PJ_VERSION="+version,
	)

	if err := cmd.Run(); err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return exitErr.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "pj: %v\n", err)
		return 1
	}
	return 0
}

func resolvePlugin(name string) (string, error) {
	if name == "" || strings.HasPrefix(name, "-") {
		return "", fmt.Errorf("invalid subcommand %q", name)
	}
	binary := pluginPrefix + name

	for _, dir := range pluginDirs() {
		if p := executableIn(dir, binary); p != "" {
			return p, nil
		}
	}
	if p, err := exec.LookPath(binary); err == nil {
		return p, nil
	}

	return "", fmt.Errorf("unknown subcommand %q", name)
}

func pluginDirs() []string {
	var dirs []string
	if d := os.Getenv("PJ_PLUGIN_DIR"); d != "" {
		dirs = append(dirs, d)
	}
	home, _ := os.UserHomeDir()
	if home != "" {
		dirs = append(dirs, filepath.Join(home, ".pj", "bin"))
	}
	if d := os.Getenv("PJ_DATA_DIR"); d != "" {
		dirs = append(dirs, filepath.Join(d, "bin"))
	}
	if d := os.Getenv("XDG_DATA_HOME"); d != "" {
		dirs = append(dirs, filepath.Join(d, "pj", "bin"))
	} else if home != "" {
		dirs = append(dirs, filepath.Join(home, ".local", "share", "pj", "bin"))
	}
	return dirs
}

func executableIn(dir, name string) string {
	p := filepath.Join(dir, name)
	info, err := os.Stat(p)
	if err != nil || info.IsDir() {
		return ""
	}
	if info.Mode()&0o111 == 0 {
		return ""
	}
	return p
}

func usage(w *os.File) {
	fmt.Fprint(w, `pj — project jumper

Usage:
    pj <subcommand> [args...]
    pj --help
    pj --version

Subcommands resolve to executables named "pj-<subcommand>".
`)
}
