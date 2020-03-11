package main

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/cli"
	"github.com/awslabs/amazon-ecs-for-open-application-model/internal/pkg/version"
)

func init() {
	cobra.EnableCommandSorting = false // Maintain the order in which we add commands.
}

func main() {
	cmd := buildRootCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func buildRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "oam-ecs",
		Short: "Provision core Open Application Model (OAM) workload types as Amazon ECS services and tasks",
		Example: `
  Display the help menu for the apply command
  $ oam-ecs apply --help`,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// If we don't set a Run() function the help menu doesn't show up.
			// See https://github.com/spf13/cobra/issues/790
		},
		SilenceUsage: true,
	}

	// Sets version for --version flag. Version command gives more detailed
	// version information.
	cmd.Version = version.Version
	cmd.SetVersionTemplate("oam-ecs version: {{.Version}}\n")

	// Commands (in the order they will show up in the help menu)
	cmd.AddCommand(cli.BuildApplyCmd())
	cmd.AddCommand(cli.BuildDeployEnvironmentCmd())

	return cmd
}
