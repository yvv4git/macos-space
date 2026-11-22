package analyse

import (
	"context"
	"time"

	"github.com/yvv4git/macos-space/internal/domain"
)

type DirectoryUsagesParams struct {
	Paths            []string
	Recursive        bool
	DirectoryTimeout time.Duration
}

type DirectorySizer interface {
	DirectoryUsages(ctx context.Context, params *DirectoryUsagesParams) ([]domain.DirectoryUsage, error)
}
