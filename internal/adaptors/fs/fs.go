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
	excludePaths := normalizeExcludePaths(params.Exclude)
	if params.Recursive {
		return f.recursiveDirectoryUsages(ctx, params, excludePaths)
	}

	usages := make([]domain.DirectoryUsage, 0, len(params.Paths))

	for _, path := range params.Paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if isExcludedPath(path, excludePaths) {
			continue
		}

		f.reportProgress(params, path)
		sizeBytes, err := f.directorySizeWithTimeout(ctx, path, params.DirectoryTimeout, excludePaths, params.OnProgress)
		if err != nil {
			return nil, fmt.Errorf("calculate directory size %q: %w", path, err)
		}

		usages = append(usages, domain.DirectoryUsage{
			Path:      path,
			SizeBytes: sizeBytes,
			IsDir:     true,
		})

		if params.IncludeFiles {
			fileUsages, err := f.directoryFileUsages(ctx, path, params.DirectoryTimeout, excludePaths)
			if err != nil {
				return nil, fmt.Errorf("list files %q: %w", path, err)
			}

			usages = append(usages, fileUsages...)
		}
	}

	return usages, nil
}

func (f *FS) directorySize(ctx context.Context, root string, excludePaths []string, onProgress func(string)) (int64, error) {
	var sizeBytes int64
	var lastProgressDir string

	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return fmt.Errorf("access %q: %w", path, walkErr)
		}
		if err := ctx.Err(); err != nil {
			return fmt.Errorf("scan %q: %w", path, err)
		}
		if isExcludedPath(path, excludePaths) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if entry.IsDir() {
			if onProgress != nil && path != lastProgressDir {
				onProgress(path)
				lastProgressDir = path
			}
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return fmt.Errorf("read file info %q: %w", path, err)
		}

		sizeBytes += info.Size()

		return nil
	})
	if err != nil {
		return 0, err
	}

	return sizeBytes, nil
}

func (f *FS) recursiveDirectoryUsages(ctx context.Context, params *analyseports.DirectoryUsagesParams, excludePaths []string) ([]domain.DirectoryUsage, error) {
	seen := make(map[string]struct{})
	usages := make([]domain.DirectoryUsage, 0)

	for _, root := range params.Paths {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if isExcludedPath(root, excludePaths) {
			continue
		}

		err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
			if walkErr != nil {
				return fmt.Errorf("access %q: %w", path, walkErr)
			}
			if err := ctx.Err(); err != nil {
				return fmt.Errorf("scan %q: %w", path, err)
			}
			if isExcludedPath(path, excludePaths) {
				if entry.IsDir() {
					return filepath.SkipDir
				}
				return nil
			}
			if !entry.IsDir() {
				if params.IncludeFiles && !f.isBeyondDepth(root, path, params.Depth) {
					if _, ok := seen[path]; ok {
						return nil
					}

					info, err := entry.Info()
					if err != nil {
						return fmt.Errorf("read file info %q: %w", path, err)
					}

					seen[path] = struct{}{}
					usages = append(usages, domain.DirectoryUsage{
						Path:      path,
						SizeBytes: info.Size(),
						IsDir:     false,
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
			f.reportProgress(params, path)
			sizeBytes, err := f.directorySizeWithTimeout(ctx, path, params.DirectoryTimeout, excludePaths, params.OnProgress)
			if err != nil {
				return err
			}

			usages = append(usages, domain.DirectoryUsage{
				Path:      path,
				SizeBytes: sizeBytes,
				IsDir:     true,
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

func (f *FS) reportProgress(params *analyseports.DirectoryUsagesParams, path string) {
	if params.OnProgress != nil {
		params.OnProgress(path)
	}
}

func (f *FS) directoryFileUsages(ctx context.Context, root string, timeout time.Duration, excludePaths []string) ([]domain.DirectoryUsage, error) {
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

		path := filepath.Join(root, entry.Name())
		if isExcludedPath(path, excludePaths) {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			return nil, err
		}

		usages = append(usages, domain.DirectoryUsage{
			Path:      path,
			SizeBytes: info.Size(),
			IsDir:     false,
		})
	}

	return usages, nil
}

func (f *FS) directorySizeWithTimeout(ctx context.Context, root string, timeout time.Duration, excludePaths []string, onProgress func(string)) (int64, error) {
	if timeout <= 0 {
		return f.directorySize(ctx, root, excludePaths, onProgress)
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	return f.directorySize(ctx, root, excludePaths, onProgress)
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

func normalizeExcludePaths(paths []string) []string {
	if len(paths) == 0 {
		return nil
	}

	normalized := make([]string, 0, len(paths))
	seen := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		if path == "" {
			continue
		}

		cleanPath := filepath.Clean(path)
		if _, ok := seen[cleanPath]; ok {
			continue
		}

		seen[cleanPath] = struct{}{}
		normalized = append(normalized, cleanPath)
	}

	return normalized
}

func isExcludedPath(path string, excludePaths []string) bool {
	cleanPath := filepath.Clean(path)
	for _, excludePath := range excludePaths {
		if cleanPath == excludePath {
			return true
		}

		relPath, err := filepath.Rel(excludePath, cleanPath)
		if err != nil {
			continue
		}
		if relPath != ".." && !strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
			return true
		}
	}

	return false
}
