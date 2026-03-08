package sync

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"
)

func TestScanForConflicts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		files    map[string]string
		expected []string
	}{
		{
			name:     "no conflicts",
			files:    map[string]string{"normal.txt": "content"},
			expected: nil,
		},
		{
			name: "single conflict file",
			files: map[string]string{
				"file.sync-conflict-ABCD123-20260307-120000.txt": "conflict",
				"normal.txt": "content",
			},
			expected: []string{"/folder/file.sync-conflict-ABCD123-20260307-120000.txt"},
		},
		{
			name: "multiple conflict files",
			files: map[string]string{
				"a.sync-conflict-ABC1234-20260307-120000.txt": "conflict1",
				"b.sync-conflict-XYZ9876-20260308-150000.txt": "conflict2",
				"normal.txt": "content",
			},
			expected: []string{
				"/folder/a.sync-conflict-ABC1234-20260307-120000.txt",
				"/folder/b.sync-conflict-XYZ9876-20260308-150000.txt",
			},
		},
		{
			name: "nested conflict file",
			files: map[string]string{
				"subdir/game.sync-conflict-DEF4567-20260309-180000.srm": "conflict",
			},
			expected: []string{"/folder/subdir/game.sync-conflict-DEF4567-20260309-180000.srm"},
		},
		{
			name: "file with similar but invalid pattern",
			files: map[string]string{
				"file.sync-conflict.txt":    "not a conflict",
				"file.sync-conflict-ab.txt": "invalid short id",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			fsMap := map[string]any{
				"/folder": map[string]any{},
			}
			folderMap := fsMap["/folder"].(map[string]any)
			for name, content := range tt.files {
				folderMap[name] = content
			}

			fs, cleanup, err := vfst.NewTestFS(fsMap)
			if err != nil {
				t.Fatalf("creating test fs: %v", err)
			}
			defer cleanup()

			conflicts, err := ScanForConflicts(fs, "/folder")
			if err != nil {
				t.Fatalf("ScanForConflicts: %v", err)
			}

			if len(conflicts) != len(tt.expected) {
				t.Errorf("expected %d conflicts, got %d: %v", len(tt.expected), len(conflicts), conflicts)
				return
			}

			expectedSet := make(map[string]bool)
			for _, e := range tt.expected {
				expectedSet[e] = true
			}
			for _, c := range conflicts {
				if !expectedSet[c] {
					t.Errorf("unexpected conflict: %s", c)
				}
			}
		})
	}
}

func TestConflictPattern(t *testing.T) {
	t.Parallel()

	tests := []struct {
		filename string
		matches  bool
	}{
		{"file.sync-conflict-ABCD123-20260307-120000.txt", true},
		{"file.sync-conflict-XYZ9876-20260308-150000", true},
		{"game.sync-conflict-ABC1234-20261231-235959.srm", true},
		{"file.sync-conflict.txt", false},
		{"file.sync-conflict-abc-123.txt", false},
		{"normal.txt", false},
		{"file.sync-conflict-AB-20260307-120000.txt", false},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			t.Parallel()
			got := conflictPattern.MatchString(tt.filename)
			if got != tt.matches {
				t.Errorf("conflictPattern.MatchString(%q) = %v, want %v", tt.filename, got, tt.matches)
			}
		})
	}
}
