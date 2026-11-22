package setup

import (
	"context"
	"fmt"
	"path/filepath"

	setupports "github.com/yvv4git/macos-space/internal/ports/setup"
)

type Params struct {
	ConfigContent string
	Force         bool
}

type Result struct {
	Skipped []string
}

type Service struct {
	configPath setupports.ConfigPath
	writer     setupports.FileWriter
}

func NewService(configPath setupports.ConfigPath, writer setupports.FileWriter) *Service {
	return &Service{
		configPath: configPath,
		writer:     writer,
	}
}

func (s *Service) Run(ctx context.Context, params *Params) (*Result, error) {
	dir, err := s.configPath.DefaultConfigDir(ctx)
	if err != nil {
		return nil, fmt.Errorf("get config dir: %w", err)
	}

	if err := s.writer.MkdirAll(ctx, &setupports.MkdirAllParams{DirPath: dir}); err != nil {
		return nil, fmt.Errorf("create config dir: %w", err)
	}

	result := &Result{}
	configPath := filepath.Join(dir, "config.toml")
	writeConfig, err := s.shouldWrite(ctx, configPath, params.Force)
	if err != nil {
		return nil, fmt.Errorf("check config: %w", err)
	}
	if writeConfig {
		if err := s.writer.WriteFile(ctx, &setupports.WriteFileParams{
			FilePath: configPath,
			Content:  params.ConfigContent,
		}); err != nil {
			return nil, fmt.Errorf("write config: %w", err)
		}

		return result, nil
	}

	result.Skipped = append(result.Skipped, configPath)

	return result, nil
}

func (s *Service) shouldWrite(ctx context.Context, path string, force bool) (bool, error) {
	if force {
		return true, nil
	}

	exists, err := s.writer.FileExists(ctx, &setupports.FileExistsParams{FilePath: path})
	if err != nil {
		return false, err
	}

	return !exists, nil
}
