package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yvv4git/macos-space/internal/application"
)

func SetupSetupCommand(configContent string) *cobra.Command {
	command := &cobra.Command{
		Use:          "setup",
		Short:        "Create default config file",
		SilenceUsage: true,
		Example: `  # Create default config in the OS-specific config directory
  macos-space setup

  # Overwrite existing config
  macos-space setup --force
  macos-space setup -f`,
		RunE: func(command *cobra.Command, _ []string) error {
			force, err := command.Flags().GetBool("force")
			if err != nil {
				return fmt.Errorf("get force flag: %w", err)
			}

			app := application.NewSetupApplication(command.Context(), application.SetupOptions{
				ConfigContent: configContent,
				Force:         force,
			})
			defer app.Shutdown()

			result, err := app.Run()
			if err != nil {
				return err
			}

			for _, path := range result.Skipped {
				command.PrintErrf("warning: %s already exists, skipping (use --force to overwrite)\n", path)
			}

			return nil
		},
	}

	command.Flags().BoolP("force", "f", false, "overwrite existing config file")

	return command
}
