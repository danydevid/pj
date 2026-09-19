package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	listFlag, whichFlag bool
)
var rootCmd = &cobra.Command{
	Use:     "pj",
	Short:   "pj - Project & Plugin CLI Manager",
	Version: "0.1.0",
	// Allow arbitrary arguments so subcommands/plugins can be passed through
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// Handle global flags if set
		if listFlag {
			fmt.Println("Listing plugins...")
			return nil
		}

		if whichFlag {
			fmt.Println("Showing plugin path...")
			return nil
		}

		// If subcommands or positional arguments are provided, pass execution to Dispatch
		if len(args) > 0 {
			return Dispatch()
		}

		// Default behavior when no arguments or flags are provided
		return cmd.Help()
	},
}

func init() {
	rootCmd.SilenceUsage = true
	rootCmd.Flags().BoolVarP(&listFlag, "list", "l", false, "list")
	rootCmd.Flags().BoolVarP(&whichFlag, "which", "w", false, "to find out where the plugin is executed")
}

func Execute() error {
	return rootCmd.Execute()
}
