package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yvv4git/macos-space/internal/application"
)

const bytesInUnit = 1024

func SetupAnalyseCommand() *cobra.Command {
	command := &cobra.Command{
		Use:          "analyse",
		Short:        "Show disk usage for configured directories",
		SilenceUsage: true,
		Example: `  # Run with default config.toml
  macos-space analyse

  # Run with explicit config path
  macos-space analyse --config config.toml
  macos-space analyse -c config.toml`,
		RunE: runAnalyse,
	}

	command.Flags().BoolP("recursive", "r", false, "walk directories recursively")

	return command
}

func runAnalyse(command *cobra.Command, _ []string) error {
	configPath, err := command.Flags().GetString("config")
	if err != nil {
		return fmt.Errorf("get config flag: %w", err)
	}

	recursive, err := command.Flags().GetBool("recursive")
	if err != nil {
		return fmt.Errorf("get recursive flag: %w", err)
	}

	app, err := application.NewAnalyseApplication(command.Context(), application.AnalyseOptions{
		ConfigPath: configPath,
	})
	if err != nil {
		return err
	}
	defer app.Shutdown()

	result, err := app.Run(recursive)
	if err != nil {
		return err
	}

	printAnalyseResult(result)

	return nil
}

func printAnalyseResult(result *application.AnalyseResult) {
	sizeWidth := 0
	formattedSizes := make([]string, 0, len(result.Directories))

	for _, directory := range result.Directories {
		formattedSize := formatBytes(directory.SizeBytes)
		if len(formattedSize) > sizeWidth {
			sizeWidth = len(formattedSize)
		}
		formattedSizes = append(formattedSizes, formattedSize)
	}

	for index, directory := range result.Directories {
		fmt.Printf("%*s  %s\n", sizeWidth, formattedSizes[index], directory.Path)
	}
}

func formatBytes(sizeBytes int64) string {
	if sizeBytes < bytesInUnit {
		return fmt.Sprintf("%d B", sizeBytes)
	}

	units := []string{"KB", "MB", "GB", "TB", "PB", "EB"}
	value := float64(sizeBytes)
	unit := "B"

	for _, currentUnit := range units {
		value /= bytesInUnit
		unit = currentUnit
		if value < bytesInUnit {
			break
		}
	}

	return fmt.Sprintf("%.1f %s", value, unit)
}
