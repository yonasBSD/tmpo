package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [description]",
	Short: "Start tracking time",
	Long: `Start a new time tracking session for the current project.`,
	Run: func(cmd* cobra.Command, args []string) {
		db, err := storage.Initialize()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		defer db.Close()

		running, err := db.GetRunningEntry()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		if running != nil {
			fmt.Fprintf(os.Stderr, "Error: Already tracking time for `%s\n", running.ProjectName)
			fmt.Println("Use 'tmpo stop' to stop the current session first.")

			os.Exit(1)
		}

		cwd, err := os.Getwd()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error getting current directory: %v\n", err)

			os.Exit(1)
		}
		
		projectName := filepath.Base(cwd)
		
		description := ""
		if len(args) > 0 {
			description = args[0]
		}

		entry, err := db.CreateEntry(projectName, description)

		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)

			os.Exit(1)
		}

		fmt.Printf("[tmpo] Started tracking time for '%s'\n", entry.ProjectName)

		if description != "" {
			fmt.Printf("	Description: %s\n", description)
		}
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
