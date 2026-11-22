package application

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/yvv4git/macos-space/internal/adaptors/fs"
	"github.com/yvv4git/macos-space/internal/core/setup"
)

type SetupOptions struct {
	ConfigContent string
	Force         bool
}

type SetupApplication struct {
	ctx     context.Context
	stop    context.CancelFunc
	service *setup.Service
	opts    SetupOptions
}

func NewSetupApplication(parent context.Context, opts SetupOptions) *SetupApplication {
	ctx, stop := signal.NotifyContext(parent, os.Interrupt, syscall.SIGTERM)

	configPath := fs.NewConfigPath()
	writer := fs.NewWriter()

	return &SetupApplication{
		ctx:     ctx,
		stop:    stop,
		service: setup.NewService(configPath, writer),
		opts:    opts,
	}
}

func (a *SetupApplication) Run() (*setup.Result, error) {
	return a.service.Run(a.ctx, &setup.Params{
		ConfigContent: a.opts.ConfigContent,
		Force:         a.opts.Force,
	})
}

func (a *SetupApplication) Shutdown() {
	a.stop()
}
