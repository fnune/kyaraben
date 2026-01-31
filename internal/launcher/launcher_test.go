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
	npLocation := filepath.Join(tmpDir, "nix-portable")
	currentDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "abc123-profile")
	currentBinDir := filepath.Join(currentDir, "bin")

	edenPkgDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-eden")
	edenBinDir := filepath.Join(edenPkgDir, "bin")

	if err := os.MkdirAll(currentBinDir, 0755); err != nil {
		t.Fatalf("creating test dirs: %v", err)
	}
	if err := os.MkdirAll(edenBinDir, 0755); err != nil {
		t.Fatalf("creating eden bin dir: %v", err)
	}

	nixpkgsBinary := filepath.Join(currentBinDir, "retroarch")
	if err := os.WriteFile(nixpkgsBinary, []byte("#!/bin/sh\necho retroarch"), 0755); err != nil {
		t.Fatalf("creating nixpkgs binary: %v", err)
	}

	realEdenBinary := filepath.Join(edenBinDir, "eden")
	if err := os.WriteFile(realEdenBinary, []byte("#!/bin/sh\necho eden"), 0755); err != nil {
		t.Fatalf("creating real eden binary: %v", err)
	}

	edenSymlink := filepath.Join(currentBinDir, "eden")
	if err := os.Symlink("/nix/store/xyz789-eden/bin/eden", edenSymlink); err != nil {
		t.Fatalf("creating eden symlink: %v", err)
	}

	hiddenBinary := filepath.Join(currentBinDir, ".retroarch-wrapped")
	if err := os.WriteFile(hiddenBinary, []byte("#!/bin/sh\necho wrapped"), 0755); err != nil {
		t.Fatalf("creating hidden binary: %v", err)
	}

	m := &Manager{
		profileDir:          profileDir,
		nixPortableBinary:   "/fake/nix-portable",
		nixPortableLocation: npLocation,
	}

	if err := os.MkdirAll(profileDir, 0755); err != nil {
		t.Fatalf("creating profile dir: %v", err)
	}
	if err := os.Symlink(currentDir, m.CurrentLink()); err != nil {
		t.Fatalf("creating current symlink: %v", err)
	}

	emulators := []EmulatorPackageInfo{
		{BinaryName: "retroarch", IsAppImage: false},
		{BinaryName: "eden", IsAppImage: true},
	}

	if err := m.GenerateWrappers(emulators); err != nil {
		t.Fatalf("GenerateWrappers() error = %v", err)
	}

	t.Run("nixpkgs wrapper uses nix shell", func(t *testing.T) {
		wrapperPath := filepath.Join(m.BinDir(), "retroarch")
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

		if !strings.Contains(wrapperStr, `nix shell "/nix/store/abc123-profile" -c "retroarch"`) {
			t.Errorf("wrapper should use nix shell with store path and binary, got:\n%s", wrapperStr)
		}

		info, err := os.Stat(wrapperPath)
		if err != nil {
			t.Fatalf("stat wrapper: %v", err)
		}
		if info.Mode()&0111 == 0 {
			t.Error("wrapper should be executable")
		}
	})

	t.Run("appimage wrapper runs directly with real path", func(t *testing.T) {
		wrapperPath := filepath.Join(m.BinDir(), "eden")
		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("reading wrapper: %v", err)
		}

		wrapperStr := string(content)

		if strings.Contains(wrapperStr, "nix shell") {
			t.Errorf("appimage wrapper should not use nix shell, got:\n%s", wrapperStr)
		}

		if strings.Contains(wrapperStr, "NP_LOCATION") {
			t.Errorf("appimage wrapper should not set NP_LOCATION, got:\n%s", wrapperStr)
		}

		expectedRealPath := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-eden", "bin", "eden")
		expectedExec := `exec "` + expectedRealPath + `"`
		if !strings.Contains(wrapperStr, expectedExec) {
			t.Errorf("appimage wrapper should exec real binary path %s, got:\n%s", expectedRealPath, wrapperStr)
		}

		info, err := os.Stat(wrapperPath)
		if err != nil {
			t.Fatalf("stat wrapper: %v", err)
		}
		if info.Mode()&0111 == 0 {
			t.Error("wrapper should be executable")
		}
	})

	t.Run("hidden files are skipped", func(t *testing.T) {
		hiddenWrapperPath := filepath.Join(m.BinDir(), ".retroarch-wrapped")
		if _, err := os.Stat(hiddenWrapperPath); !os.IsNotExist(err) {
			t.Error("hidden files (nixpkgs internal wrappers) should not be wrapped")
		}
	})
}
