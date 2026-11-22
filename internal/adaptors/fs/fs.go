package fs

import (
	"context"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/yvv4git/macos-space/internal/domain"
	analyseports "github.com/yvv4git/macos-space/internal/ports/analyse"
)

type FS struct{}

func New() *FS {
	return &FS{}
}

func (f *FS) DirectoryUsages(ctx context.Context, params *analyseports.DirectoryUsagesParams) ([]domain.DirectoryUsage, error) {
	if params.Recursive {
		return f.recursiveDirectoryUsages(ctx, params)
	}

	usages := make([]domain.DirectoryUsage, 0, len(params.Paths))

	for _, path := range params.Paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		sizeBytes, err := f.directorySizeWithTimeout(ctx, path, params.DirectoryTimeout)
		if err != nil {
			return nil, fmt.Errorf("calculate directory size %q: %w", path, err)
		}

		usages = append(usages, domain.DirectoryUsage{
			Path:      path,
			SizeBytes: sizeBytes,
		})

		if params.IncludeFiles {
			fileUsages, err := f.directoryFileUsages(ctx, path, params.DirectoryTimeout)
			if err != nil {
				return nil, fmt.Errorf("list files %q: %w", path, err)
			}

			usages = append(usages, fileUsages...)
		}
	}

	return usages, nil
}

func (f *FS) directorySize(ctx context.Context, root string) (int64, error) {
	var sizeBytes int64

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		sizeBytes += info.Size()

		return nil
	})
	if err != nil {
		if os.IsNotExist(err) {
			return 0, err
		}

		return 0, err
	}

	return sizeBytes, nil
}

func (f *FS) recursiveDirectoryUsages(ctx context.Context, params *analyseports.DirectoryUsagesParams) ([]domain.DirectoryUsage, error) {
	seen := make(map[string]struct{})
	usages := make([]domain.DirectoryUsage, 0)

	for _, root := range params.Paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if err := ctx.Err(); err != nil {
				return err
			}
			if !entry.IsDir() {
				if params.IncludeFiles && !f.isBeyondDepth(root, path, params.Depth) {
					if _, ok := seen[path]; ok {
						return nil
					}

					info, err := entry.Info()
					if err != nil {
						return err
					}

					seen[path] = struct{}{}
					usages = append(usages, domain.DirectoryUsage{
						Path:      path,
						SizeBytes: info.Size(),
					})
				}

				return nil
			}
			if f.isBeyondDepth(root, path, params.Depth) {
				return filepath.SkipDir
			}
			if _, ok := seen[path]; ok {
				return nil
			}

			seen[path] = struct{}{}
			sizeBytes, err := f.directorySizeWithTimeout(ctx, path, params.DirectoryTimeout)
			if err != nil {
				return err
			}

			usages = append(usages, domain.DirectoryUsage{
				Path:      path,
				SizeBytes: sizeBytes,
			})

			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("walk directory %q: %w", root, err)
		}
	}

	sort.Slice(usages, func(i, j int) bool {
		return usages[i].SizeBytes > usages[j].SizeBytes
	})

	return usages, nil
}

func (f *FS) directoryOwnSize(ctx context.Context, root string) (int64, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return 0, err
	}

	var sizeBytes int64
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return 0, err
		}
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return 0, err
		}

		sizeBytes += info.Size()
	}

	return sizeBytes, nil
}

func (f *FS) directoryFileUsages(ctx context.Context, root string, timeout time.Duration) ([]domain.DirectoryUsage, error) {
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return nil, err
	}

	usages := make([]domain.DirectoryUsage, 0, len(entries))
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if entry.IsDir() {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		usages = append(usages, domain.DirectoryUsage{
			Path:      filepath.Join(root, entry.Name()),
			SizeBytes: info.Size(),
		})
	}

	return usages, nil
}

func (f *FS) directorySizeWithTimeout(ctx context.Context, root string, timeout time.Duration) (int64, error) {
	if timeout <= 0 {
		return f.directorySize(ctx, root)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return f.directorySize(ctx, root)
}

func (f *FS) directoryOwnSizeWithTimeout(ctx context.Context, root string, timeout time.Duration) (int64, error) {
	if timeout <= 0 {
		return f.directoryOwnSize(ctx, root)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return f.directoryOwnSize(ctx, root)
}

func (f *FS) isBeyondDepth(root, path string, depth int) bool {
	if depth <= 0 {
		return false
	}

	relPath, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	if relPath == "." {
		return false
	}

	segments := 0
	for _, segment := range filepath.SplitList(filepath.ToSlash(relPath)) {
		_ = segment
	}
	segments = len(stringsSplitPath(relPath))

	return segments > depth
}

func stringsSplitPath(path string) []string {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." {
		return nil
	}

	parts := make([]string, 0)
	for _, part := range strings.Split(cleanPath, string(os.PathSeparator)) {
		if part == "" || part == "." {
			continue
		}

		parts = append(parts, part)
	}

	return parts
}
