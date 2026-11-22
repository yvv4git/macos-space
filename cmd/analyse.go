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
		Example: `  # Run with default config
  macos-space analyse

  # Run with explicit config path
  macos-space analyse --config config.toml
  macos-space analyse -c config.toml

  # Scan a specific directory
  macos-space analyse --path /Users/vladimireliseev/Downloads
  macos-space analyse -p /Users/vladimireliseev/Downloads

  # Walk directories recursively
  macos-space analyse --recursive
  macos-space analyse -r

  # Limit recursive depth (0 means unlimited)
  macos-space analyse --depth 2
  macos-space analyse -d 2

  # Combine flags
  macos-space analyse --recursive --depth 2
  macos-space analyse -r -d 2
  macos-space analyse -p ~/Downloads -r -d 1
  macos-space analyse --path /Users/vladimireliseev/Downloads --recursive
  macos-space analyse -p /Users/vladimireliseev/Downloads -r
  macos-space analyse --path /Users/vladimireliseev/Downloads --recursive --depth 2
  macos-space analyse -p /Users/vladimireliseev/Downloads -r -d 2
  macos-space analyse --config config.toml --path /Users/vladimireliseev/Downloads --recursive --depth 2
  macos-space analyse -c config.toml -p /Users/vladimireliseev/Downloads -r -d 2`,
		RunE: runAnalyse,
	}

	command.Flags().BoolP("recursive", "r", false, "walk directories recursively")
	command.Flags().IntP("depth", "d", 0, "limit recursive depth (0 means unlimited)")
	command.Flags().StringP("path", "p", "", "scan a specific directory instead of configured directories")

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

	depth, err := command.Flags().GetInt("depth")
	if err != nil {
		return fmt.Errorf("get depth flag: %w", err)
	}

	path, err := command.Flags().GetString("path")
	if err != nil {
		return fmt.Errorf("get path flag: %w", err)
	}

	app, err := application.NewAnalyseApplication(command.Context(), application.AnalyseOptions{
		ConfigPath: configPath,
	})
	if err != nil {
		return err
	}
	defer app.Shutdown()

	result, err := app.Run(recursive, depth, path)
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
