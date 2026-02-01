package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRealToVirtualStorePath(t *testing.T) {
	tests := []struct {
		name     string
		realPath string
		want     string
		wantErr  bool
	}{
		{
			name:     "typical nix-portable path",
			realPath: "/home/user/.local/state/kyaraben/build/nix/.nix-portable/nix/store/abc123-package",
			want:     "/nix/store/abc123-package",
		},
		{
			name:     "path with nested store reference",
			realPath: "/some/prefix/nix/store/hash-name/bin/program",
			want:     "/nix/store/hash-name/bin/program",
		},
		{
			name:     "no nix store in path",
			realPath: "/home/user/.local/bin/something",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := realToVirtualStorePath(tt.realPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("realToVirtualStorePath() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("realToVirtualStorePath() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestGenerateWrappers(t *testing.T) {
	tmpDir := t.TempDir()

	profileDir := filepath.Join(tmpDir, "kyaraben")
	currentDir := filepath.Join(tmpDir, "store", "nix", "store", "abc123-profile")
	currentBinDir := filepath.Join(currentDir, "bin")

	if err := os.MkdirAll(currentBinDir, 0755); err != nil {
		t.Fatalf("creating test dirs: %v", err)
	}

	testBinary := filepath.Join(currentBinDir, "testemu")
	if err := os.WriteFile(testBinary, []byte("#!/bin/sh\necho test"), 0755); err != nil {
		t.Fatalf("creating test binary: %v", err)
	}

	// Create hidden files that should be skipped (nixpkgs internal wrappers)
	hiddenBinary := filepath.Join(currentBinDir, ".testemu-wrapped")
	if err := os.WriteFile(hiddenBinary, []byte("#!/bin/sh\necho wrapped"), 0755); err != nil {
		t.Fatalf("creating hidden binary: %v", err)
	}

	m := &Manager{
		profileDir:        profileDir,
		nixPortableBinary: "/fake/nix-portable",
	}

	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("creating profile dir: %v", err)
	}
	if err := os.Symlink(currentDir, m.CurrentLink()); err != nil {
		t.Fatalf("creating current symlink: %v", err)
	}

	if err := m.GenerateWrappers(); err != nil {
		t.Fatalf("GenerateWrappers() error = %v", err)
	}

	wrapperPath := filepath.Join(m.BinDir(), "testemu")
	content, err := os.ReadFile(wrapperPath)
	if err != nil {
		t.Fatalf("reading wrapper: %v", err)
	}

	wrapperStr := string(content)

	if !strings.Contains(wrapperStr, "/fake/nix-portable") {
		t.Errorf("wrapper should contain nix-portable path, got:\n%s", wrapperStr)
	}

	if !strings.Contains(wrapperStr, "/nix/store/abc123-profile") {
		t.Errorf("wrapper should contain virtual store path, got:\n%s", wrapperStr)
	}

	if !strings.Contains(wrapperStr, `nix shell "/nix/store/abc123-profile" -c "testemu"`) {
		t.Errorf("wrapper should use nix shell with store path and binary, got:\n%s", wrapperStr)
	}

	info, err := os.Stat(wrapperPath)
	if err != nil {
		t.Fatalf("stat wrapper: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("wrapper should be executable")
	}

	// Verify hidden files are not wrapped
	hiddenWrapperPath := filepath.Join(m.BinDir(), ".testemu-wrapped")
	if _, err := os.Stat(hiddenWrapperPath); !os.IsNotExist(err) {
		t.Error("hidden files (nixpkgs internal wrappers) should not be wrapped")
	}
}
