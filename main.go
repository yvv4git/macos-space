package main

import (
	"context"
	_ "embed"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/yvv4git/macos-space/cmd"
	"github.com/yvv4git/macos-space/internal/adaptors/fs"
)

//go:embed config.example.toml
var defaultConfigContent string

func main() {
	rootCommand := &cobra.Command{
		Use:   "macos-space",
		Short: "CLI for macOS disk space analysis",
	}
	rootCommand.PersistentFlags().StringP("config", "c", defaultConfigPath(), "Path to config file")

	rootCommand.AddCommand(cmd.SetupAnalyseCommand())
	rootCommand.AddCommand(cmd.SetupClearCommand())
	rootCommand.AddCommand(cmd.SetupSetupCommand(defaultConfigContent))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := rootCommand.ExecuteContext(ctx); err != nil {
		stop()
		os.Exit(1)
	}
}

func defaultConfigPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return fs.ConfigFileName
	}

	return filepath.Join(home, fs.ConfigDirName, fs.ConfigAppDirName, fs.ConfigFileName)
}
