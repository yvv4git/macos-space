package fs

import (
	"context"
	"errors"
	"fmt"
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

func TestDirectoryUsagesSkipsPermissionErrors(t *testing.T) {
	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}

	root := t.TempDir()
	restricted := filepath.Join(root, "restricted")
	allowed := filepath.Join(root, "allowed")
	for _, dir := range []string{restricted, allowed} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create directory %q: %v", dir, err)
		}
	}

	path := filepath.Join(allowed, "file.txt")
	if err := os.WriteFile(path, make([]byte, 4096), 0o600); err != nil {
		t.Fatalf("create file %q: %v", path, err)
	}
	if err := os.Chmod(restricted, 0o000); err != nil {
		t.Fatalf("chmod directory %q: %v", restricted, err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(restricted, 0o755)
	})

	for _, recursive := range []bool{false, true} {
		params := &analyseports.DirectoryUsagesParams{
			Paths:            []string{root},
			Recursive:        recursive,
			DirectoryTimeout: time.Second,
		}

		usages, err := New().DirectoryUsages(context.Background(), params)
		if err != nil {
			t.Fatalf("directory usages, recursive %t: %v", recursive, err)
		}

		found := false
		for _, usage := range usages {
			if usage.Path == allowed {
				found = true
				if usage.SizeBytes == 0 {
					t.Errorf("zero size for %q, recursive %t", allowed, recursive)
				}
			}
		}
		if !found {
			t.Errorf("missing path %q, recursive %t", allowed, recursive)
		}
	}
}

func TestDirectoryUsagesSkipsTimeout(t *testing.T) {
	root := t.TempDir()
	slow := filepath.Join(root, "slow")
	fast := filepath.Join(root, "fast")
	slowChild := filepath.Join(slow, "child")
	fastChild := filepath.Join(fast, "child")

	for _, dir := range []string{slowChild, fastChild} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create directory %q: %v", dir, err)
		}
	}

	const slowFiles = 20000
	for index := 0; index < slowFiles; index++ {
		path := filepath.Join(slowChild, fmt.Sprintf("file-%d.txt", index))
		if err := os.WriteFile(path, nil, 0o600); err != nil {
			t.Fatalf("create file %q: %v", path, err)
		}
	}

	path := filepath.Join(fastChild, "file.txt")
	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("create file %q: %v", path, err)
	}

	for _, recursive := range []bool{false, true} {
		params := &analyseports.DirectoryUsagesParams{
			Paths:            []string{slow, fast},
			Recursive:        recursive,
			DirectoryTimeout: 20 * time.Millisecond,
		}

		usages, err := New().DirectoryUsages(context.Background(), params)
		if err != nil {
			t.Fatalf("directory usages, recursive %t: %v", recursive, err)
		}
		if !containsUsagePath(usages, fastChild) {
			t.Errorf("missing path %q, recursive %t", fastChild, recursive)
		}
		if containsUsagePath(usages, slowChild) {
			t.Errorf("timed out path %q present, recursive %t", slowChild, recursive)
		}
	}
}

func TestDirectoryUsagesReportsNestedProgress(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	nested := filepath.Join(child, "nested")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatalf("create directory %q: %v", nested, err)
	}
	if err := os.WriteFile(filepath.Join(nested, "file.txt"), nil, 0o600); err != nil {
		t.Fatalf("create file: %v", err)
	}

	progress := make([]string, 0)
	params := &analyseports.DirectoryUsagesParams{
		Paths:            []string{root},
		DirectoryTimeout: time.Second,
		OnProgress: func(path string) {
			progress = append(progress, path)
		},
	}
	if _, err := New().DirectoryUsages(context.Background(), params); err != nil {
		t.Fatalf("directory usages: %v", err)
	}

	found := false
	for _, path := range progress {
		if path == nested {
			found = true
		}
	}
	if !found {
		t.Errorf("missing nested path %q in progress %v", nested, progress)
	}
}

func TestDirectoryUsagesStopsOnCanceledContext(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "child")
	if err := os.MkdirAll(child, 0o755); err != nil {
		t.Fatalf("create directory %q: %v", child, err)
	}
	if err := os.WriteFile(filepath.Join(child, "file.txt"), nil, 0o600); err != nil {
		t.Fatalf("create file: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	params := &analyseports.DirectoryUsagesParams{
		Paths:            []string{root},
		DirectoryTimeout: time.Minute,
	}
	_, err := New().DirectoryUsages(ctx, params)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context canceled, got %v", err)
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
