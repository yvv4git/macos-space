package fs

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yvv4git/macos-space/internal/domain"
	analyseports "github.com/yvv4git/macos-space/internal/ports/analyse"
)

func TestDirectoryUsagesExplicitRootOverridesExclude(t *testing.T) {
	root := t.TempDir()
	parent := filepath.Join(root, "parent")
	explicit := filepath.Join(parent, "explicit")
	child := filepath.Join(explicit, "child")
	skipped := filepath.Join(explicit, "skipped")

	for _, dir := range []string{child, skipped} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create directory %q: %v", dir, err)
		}
	}

	for _, dir := range []string{child, skipped} {
		path := filepath.Join(dir, "file.txt")
		if err := os.WriteFile(path, []byte("data"), 0o600); err != nil {
			t.Fatalf("create file %q: %v", path, err)
		}
	}

	tests := []struct {
		name      string
		recursive bool
		wantPath  string
	}{
		{
			name:      "non recursive",
			recursive: false,
			wantPath:  child,
		},
		{
			name:      "recursive",
			recursive: true,
			wantPath:  explicit,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			params := &analyseports.DirectoryUsagesParams{
				Paths:            []string{explicit},
				Recursive:        tt.recursive,
				Exclude:          []string{parent, skipped},
				DirectoryTimeout: time.Second,
			}

			usages, err := New().DirectoryUsages(context.Background(), params)
			if err != nil {
				t.Fatalf("directory usages: %v", err)
			}
			if !containsUsagePath(usages, tt.wantPath) {
				t.Errorf("missing path %q in %v", tt.wantPath, usages)
			}
			if containsUsagePath(usages, skipped) {
				t.Errorf("excluded path %q present in %v", skipped, usages)
			}
		})
	}
}

func containsUsagePath(usages []domain.DirectoryUsage, path string) bool {
	for _, usage := range usages {
		if usage.Path == path {
			return true
		}
	}

	return false
}
