package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yvv4git/macos-space/internal/application"
)

func SetupClearCommand() *cobra.Command {
	command := &cobra.Command{
		Use:          "clear",
		Short:        "Clear files in configured directories",
		SilenceUsage: true,
		Example: `  # Run with default config.toml
  macos-space clear

  # Run with explicit config path
  macos-space clear --config config.toml
  macos-space clear -c config.toml`,
		RunE: runClear,
	}

	return command
}

func runClear(command *cobra.Command, _ []string) error {
	configPath, err := command.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("get config flag: %w", err)
	}

	app, err := application.NewClearApplication(command.Context(), application.ClearOptions{
		ConfigPath: configPath,
	})
	if err != nil {
		return err
	}
	defer app.Shutdown()

	if err := app.Run(); err != nil {
		return err
	}

	fmt.Println("clear command is not implemented yet")

	return nil
}
