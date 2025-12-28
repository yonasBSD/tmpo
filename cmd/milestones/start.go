package milestones

import (
	"fmt"
	"os"

	"github.com/DylanDevelops/tmpo/internal/project"
	"github.com/DylanDevelops/tmpo/internal/storage"
	"github.com/DylanDevelops/tmpo/internal/ui"
	"github.com/spf13/cobra"
)

// StartCmd returns a command that creates and activates a new milestone for the current project.
// If a milestone with the same name already exists for the project, or if there is already an
// active milestone, an error is displayed and the command exits.
func StartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [name]",
		Short: "Start a new milestone",
		Long:  `Start a new milestone for the current project. Time entries created after starting a milestone will be automatically tagged with it.`,
		Args:  cobra.ExactArgs(1),
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

			milestoneName := args[0]

			// Check if there's already an active milestone
			activeMilestone, err := db.GetActiveMilestoneForProject(projectName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			if activeMilestone != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("Milestone '%s' is already active for %s", activeMilestone.Name, projectName))
				ui.PrintMuted(0, "Use 'tmpo milestone finish' to finish it first.")
				ui.NewlineBelow()
				os.Exit(1)
			}

			// Create the milestone
			milestone, err := db.CreateMilestone(projectName, milestoneName)
			if err != nil {
				ui.PrintError(ui.EmojiError, fmt.Sprintf("%v", err))
				os.Exit(1)
			}

			ui.PrintSuccess(ui.EmojiMilestone, fmt.Sprintf("Started milestone %s for %s", ui.Bold(milestone.Name), ui.Bold(projectName)))
			ui.PrintMuted(4, "└─ New time entries will be automatically tagged")
			ui.NewlineBelow()
		},
	}

	return cmd
}
