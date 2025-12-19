package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

var rootCmd = &cobra.Command{
	Use:   "tmpo",
	Short: "Minimal CLI time tracker for developers",
	Long: `tmpo - Set the tmpo

A minimal, developer-friendly time tracking tool that lives in your terminal.
Track time effortlessly with automatic project detection and simple commands.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Check if version flag was set
		versionFlag, _ := cmd.Flags().GetBool("version")

		if versionFlag {
			DisplayVersionWithUpdateCheck()
			return
		}

		// Otherwise show help
		cmd.Help()
	},
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "version for tmpo")
}
