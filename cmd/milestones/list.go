package milestones

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/settings"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

var (
	listProject string
	listAll     bool
)

// ListCmd returns a command that lists milestones. By default, it lists milestones for the current project.
// Use --all to list milestones from all projects, or --project to list milestones for a specific project.
func ListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List milestones",
		Long:  `List milestones for the current project. Use --all to list milestones from all projects.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			var milestones []*storage.Milestone
			var projectName string

			// Determine scope
			if listAll {
				milestones, err = db.GetAllMilestones()
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}
			} else {
				if listProject != "" {
					projectName = listProject
				} else {
					projectName, err = project.DetectConfiguredProject()
					if err != nil {
						ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
						os.Exit(1)
					}
				}

				milestones, err = db.GetMilestonesByProject(projectName)
				if err != nil {
					ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
					os.Exit(1)
				}
			}

			if len(milestones) == 0 {
				ui.PrintWarning(ui.EmojiWarning, "No milestones found")
				ui.NewlineBelow()
				return
			}

			// Print header
			if listAll {
				ui.PrintSuccess(ui.EmojiMilestone, "All Milestones")
			} else {
				ui.PrintSuccess(ui.EmojiMilestone, fmt.Sprintf("Milestones for %s", ui.Bold(projectName)))
			}
			ui.NewlineBelow()

			// Group by active/finished
			var activeMilestones []*storage.Milestone
			var finishedMilestones []*storage.Milestone

			for _, m := range milestones {
				if m.IsActive() {
					activeMilestones = append(activeMilestones, m)
				} else {
					finishedMilestones = append(finishedMilestones, m)
				}
			}

			// Print active milestones
			if len(activeMilestones) > 0 {
				fmt.Printf("%s Active %s\n", ui.Muted("───"), ui.Muted("───"))
				for _, m := range activeMilestones {
					// Get entry count for this milestone
					entries, _ := db.GetEntriesByMilestone(m.ProjectName, m.Name)
					entryCount := len(entries)

					if listAll {
						fmt.Printf("  %s (%s)\n", ui.Bold(m.Name), m.ProjectName)
					} else {
						fmt.Printf("  %s\n", ui.Bold(m.Name))
					}
					fmt.Printf("    Started: %s  Duration: %s  Entries: %d\n",
						settings.FormatTime(m.StartTime),
						ui.FormatDuration(m.Duration()),
						entryCount)
					fmt.Println()
				}
			}

			// Print finished milestones
			if len(finishedMilestones) > 0 {
				fmt.Printf("%s Finished %s\n", ui.Muted("───"), ui.Muted("───"))
				for _, m := range finishedMilestones {
					// Get entry count for this milestone
					entries, _ := db.GetEntriesByMilestone(m.ProjectName, m.Name)
					entryCount := len(entries)

					if listAll {
						fmt.Printf("  %s (%s)\n", ui.Bold(m.Name), m.ProjectName)
					} else {
						fmt.Printf("  %s\n", ui.Bold(m.Name))
					}
					fmt.Printf("    %s - %s  Duration: %s  Entries: %d\n",
						settings.FormatTime(m.StartTime),
						settings.FormatTime(*m.EndTime),
						ui.FormatDuration(m.Duration()),
						entryCount)
					fmt.Println()
				}
			}

			ui.NewlineBelow()
		},
	}

	cmd.Flags().StringVarP(&listProject, "project", "p", "", "Show milestones for specific project")
	cmd.Flags().BoolVarP(&listAll, "all", "a", false, "Show milestones from all projects")

	return cmd
}
