package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func Dispatch() error {
	env, err := ResolvePlugin("jump.sh")
	if err != nil {
		return err
	}

	pluginArgs := os.Args[2:]
	cmd := exec.Command(env, pluginArgs...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	if err := cmd.Wait(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			fmt.Printf("Exit code: %d\n", exitErr.ExitCode())
		}
		return fmt.Errorf("process failed: %w", err)
	}

	return nil
}
