package config

import "time"

type Config struct {
	Analyse Analyse `toml:"analyse"`
}

type Analyse struct {
	Directories      []string      `toml:"directories"`
	DirectoryTimeout time.Duration `toml:"directoryTimeout"`
}
