package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

const defaultDirectoryTimeout = 3 * time.Second

func Load(path string, cfg *Config) error {
	if err := cleanenv.ReadConfig(path, cfg); err != nil {
		return fmt.Errorf("read config: %w", err)
	}

	if cfg.Analyse.DirectoryTimeout == 0 {
		cfg.Analyse.DirectoryTimeout = defaultDirectoryTimeout
	}

	cfg.Analyse.Directories = expandPaths(cfg.Analyse.Directories)
	cfg.Analyse.Exclude = expandPaths(cfg.Analyse.Exclude)

	return nil
}

func expandPaths(paths []string) []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return paths
	}

	expanded := make([]string, 0, len(paths))
	for _, path := range paths {
		if path == "~" {
			expanded = append(expanded, home)
			continue
		}
		if strings.HasPrefix(path, "~/") {
			expanded = append(expanded, filepath.Join(home, path[2:]))
			continue
		}
		expanded = append(expanded, path)
	}

	return expanded
}
