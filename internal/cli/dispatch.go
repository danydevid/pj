package cli

import (
	"fmt"
	"os"
	"os/exec"
)

func Dispatch() error {
	fmt.Println(os.Args[:])
	plugin := os.Args[1]
	env, err := Resolve("PJ-PLUGIN_DIR")
	if(err != nil) {
		return err
	}
	fmt.Println(env)

	pluginArgs := os.Args[2:]
	cmd := exec.Command(plugin, pluginArgs...)
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
    	return fmt.Errorf("proses gagal: %w", err)
	}

	return nil
}
