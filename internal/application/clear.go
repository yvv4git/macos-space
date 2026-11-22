package application

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/yvv4git/macos-space/internal/config"
)

type ClearOptions struct {
	ConfigPath string
}

type ClearApplication struct {
	ctx  context.Context
	stop context.CancelFunc
	cfg  config.Config
}

func NewClearApplication(parent context.Context, opts ClearOptions) (*ClearApplication, error) {
	var cfg config.Config
	if err := config.Load(opts.ConfigPath, &cfg); err != nil {
		return nil, fmt.Errorf("load config: %w", err)
	}

	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)

	return &ClearApplication{
		ctx:  ctx,
		stop: stop,
		cfg:  cfg,
	}, nil
}

func (a *ClearApplication) Run() error {
	return nil
}

func (a *ClearApplication) Shutdown() {
	a.stop()
}
