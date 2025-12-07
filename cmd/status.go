package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show current tracking status",
	Long:  `Display information about the currently running time tracking session.`,

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

		if running == nil {
			fmt.Println("[tmpo] Not currently tracking time")
			fmt.Println("\nUse 'tmpo start' to begin tracking")
			
			return
		}

		duration := time.Since(running.StartTime)

		fmt.Printf("[tmpo] Currently tracking: %s\n", running.ProjectName)
		fmt.Printf("    Started: %s\n", running.StartTime.Format("3:04 PM"))
		fmt.Printf("    Duration: %s\n", formatDuration(duration))

		if running.Description != "" {
			fmt.Printf("    Description: %s\n", running.Description)
		}
	},
}

func init() {
	rootCmd.AddCommand(statusCmd)
}
