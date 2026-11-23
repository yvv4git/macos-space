package config

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestExpandPaths(t *testing.T) {
	home, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("user home dir: %v", err)
	}

	tests := []struct {
		name  string
		paths []string
		want  []string
	}{
		{
			name:  "nil",
			paths: nil,
			want:  []string{},
		},
		{
			name:  "home",
			paths: []string{"~"},
			want:  []string{home},
		},
		{
			name:  "home subpath",
			paths: []string{"~/Downloads", "~/Library/Caches"},
			want: []string{
				filepath.Join(home, "Downloads"),
				filepath.Join(home, "Library", "Caches"),
			},
		},
		{
			name:  "mixed paths",
			paths: []string{"/Library/Logs", "~/Downloads", "relative/path", "~user/path"},
			want: []string{
				"/Library/Logs",
				filepath.Join(home, "Downloads"),
				"relative/path",
				"~user/path",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := expandPaths(tt.paths)
			if !slices.Equal(got, tt.want) {
				t.Errorf("expandPaths(%q) = %q, want %q", tt.paths, got, tt.want)
			}
		})
	}
}
