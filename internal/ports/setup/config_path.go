package setup

import "context"

type ConfigPath interface {
	DefaultConfigDir(ctx context.Context) (string, error)
}
