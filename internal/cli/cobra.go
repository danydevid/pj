package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	versionFlag, listFlag, whichFlag bool
)

var rootCmd = &cobra.Command{
	Use:   "pj",
	Short: "a extensible project manager",
	Long:  "`pj` is an extensible **project jumper**. Its architecture follows a **microkernel + plugin** pattern with a **shell adapter** for terminal integration.",
	Version: "0.1.0",
}

func init() {
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "list")
	rootCmd.Flags().BoolVarP(&whichFlag, "which", "w", false, "to find out where the plugin is executed")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Execution failed: %v\n", err)
		os.Exit(1)
	}
}
