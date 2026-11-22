package analyse

import (
	"context"
	"sort"
	"time"

	"github.com/yvv4git/macos-space/internal/domain"
	analyseports "github.com/yvv4git/macos-space/internal/ports/analyse"
)

type Result struct {
	Directories []domain.DirectoryUsage
}

type Service struct {
	sizer analyseports.DirectorySizer
}

func NewService(sizer analyseports.DirectorySizer) *Service {
	return &Service{sizer: sizer}
}

func (s *Service) Run(ctx context.Context, paths []string, recursive bool, depth int, directoryTimeout time.Duration) (*Result, error) {
	if len(paths) == 0 {
		return nil, domain.ErrNoDirectories
	}

	usages, err := s.sizer.DirectoryUsages(ctx, &analyseports.DirectoryUsagesParams{
		Paths:            paths,
		Recursive:        recursive,
		Depth:            depth,
		DirectoryTimeout: directoryTimeout,
	})
	if err != nil {
		return nil, err
	}

	filtered := make([]domain.DirectoryUsage, 0, len(usages))
	for _, usage := range usages {
		if usage.SizeBytes == 0 {
			continue
		}

		filtered = append(filtered, usage)
	}
	usages = filtered

	sort.Slice(usages, func(i, j int) bool {
		if usages[i].SizeBytes == usages[j].SizeBytes {
			return usages[i].Path < usages[j].Path
		}

		return usages[i].SizeBytes > usages[j].SizeBytes
	})

	return &Result{Directories: usages}, nil
}
