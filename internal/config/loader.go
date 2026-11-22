package config

import (
	"fmt"
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

	return nil
}
