package milestones

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

// FinishCmd returns a command that finishes the currently active milestone for the current project.
// If there is no active milestone, an error is displayed and the command exits.
func FinishCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "finish",
		Short: "Finish the active milestone",
		Long:  `Finish the currently active milestone for the current project. This marks the milestone as completed and stops auto-tagging new time entries with it.`,
		Run: func(cmd *cobra.Command, args []string) {
			ui.NewlineAbove()

			db, err := storage.Initialize()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}
			defer db.Close()

			projectName, err := project.DetectConfiguredProject()
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("detecting project: %v", err))
				os.Exit(1)
			}

			// Get active milestone
			activeMilestone, err := db.GetActiveMilestoneForProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if activeMilestone == nil {
				ui.PrintError(ui.EmojiError, "No active milestone found")
				ui.PrintMuted(0, "Use 'tmpo milestone start' to start a new milestone.")
				ui.NewlineBelow()
				os.Exit(1)
			}

			// Get entries for this milestone to show count
			entries, err := db.GetEntriesByMilestone(projectName, activeMilestone.Name)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// Finish the milestone
			err = db.FinishMilestone(activeMilestone.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			// Get updated milestone to show duration
			finishedMilestone, err := db.GetMilestone(activeMilestone.ID)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiMilestone, fmt.Sprintf("Finished milestone %s", ui.Bold(finishedMilestone.Name)))
			ui.PrintInfo(4, "Duration", ui.FormatDuration(finishedMilestone.Duration()))
			ui.PrintInfo(4, "Entries", fmt.Sprintf("%d", len(entries)))
			ui.NewlineBelow()
		},
	}

	return cmd
}
