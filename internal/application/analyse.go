package application

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/yvv4git/macos-space/internal/adaptors/fs"
	"github.com/yvv4git/macos-space/internal/config"
	"github.com/yvv4git/macos-space/internal/core/analyse"
	"github.com/yvv4git/macos-space/internal/domain"
)

type AnalyseOptions struct {
	ConfigPath string
}

type AnalyseApplication struct {
	ctx     context.Context
	stop    context.CancelFunc
	service *analyse.Service
	cfg     config.Config
}

type AnalyseRunInfo struct {
	Directories      []string
	Recursive        bool
	Depth            int
	DirectoryTimeout time.Duration
	IncludeFiles     bool
}

func NewAnalyseApplication(parent context.Context, opts AnalyseOptions) (*AnalyseApplication, error) {
	var cfg config.Config
	if err := config.Load(opts.ConfigPath, &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)

	return &AnalyseApplication{
		ctx:     ctx,
		stop:    stop,
		service: analyse.NewService(fs.New()),
		cfg:     cfg,
	}, nil
}

func (a *AnalyseApplication) RunInfo(recursive bool, depth int, path string, includeFiles bool) *AnalyseRunInfo {
	directories := a.cfg.Analyse.Directories
	if path != "" {
		directories = []string{path}
	}

	return &AnalyseRunInfo{
		Directories:      directories,
		Recursive:        recursive,
		Depth:            depth,
		DirectoryTimeout: a.cfg.Analyse.DirectoryTimeout,
		IncludeFiles:     includeFiles,
	}
}

func (a *AnalyseApplication) Run(recursive bool, depth int, path string, includeFiles bool) (*analyse.Result, error) {
	directories := a.cfg.Analyse.Directories
	if path != "" {
		directories = []string{path}
	}

	return a.service.Run(a.ctx, directories, recursive, depth, includeFiles, a.cfg.Analyse.DirectoryTimeout)
}

func (a *AnalyseApplication) Shutdown() {
	a.stop()
}

type AnalyseResult = analyse.Result
type DirectoryUsage = domain.DirectoryUsage
