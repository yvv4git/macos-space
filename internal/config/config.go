package config

import "time"

type Config struct {
	Analyse Analyse `toml:"analyse"`
}

type Analyse struct {
	Directories      []string      `toml:"directories"`
	Exclude          []string      `toml:"exclude"`
	DirectoryTimeout time.Duration `toml:"directoryTimeout"`
	Depth            int           `toml:"depth"`
}
