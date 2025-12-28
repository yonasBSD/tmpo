package milestones

import "github.com/spf13/cobra"

// MilestoneCmds returns the root milestone command that groups all milestone-related subcommands.
// It provides functionality to manage milestones (time-boxed periods like sprints, releases, or phases)
// for grouping time entries.
func MilestoneCmds() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "milestone",
		Short: "Manage milestones",
		Long:  `Manage milestones to group time entries into time-boxed periods (sprints, releases, phases).`,
	}

	cmd.AddCommand(StartCmd())
	cmd.AddCommand(FinishCmd())
	cmd.AddCommand(StatusCmd())
	cmd.AddCommand(ListCmd())

	return cmd
}
