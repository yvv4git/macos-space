package fs

import (
	"context"
	"os"
	"path/filepath"
)

const (
	ConfigDirName    = ".config"
	ConfigAppDirName = "macos-space"
	ConfigFileName   = "config.toml"
	fileMode         = 0o600
	dirMode          = 0o700
)

type ConfigPath struct{}

func NewConfigPath() *ConfigPath {
	return &ConfigPath{}
}

func (c *ConfigPath) DefaultConfigDir(_ context.Context) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(home, ConfigDirName, ConfigAppDirName), nil
}
