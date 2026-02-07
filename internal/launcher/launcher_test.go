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

	mgbaPkgDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "def456-mgba")
	mgbaBinDir := filepath.Join(mgbaPkgDir, "bin")

	if err := os.MkdirAll(currentBinDir, 0755); err != nil {
		t.Fatalf("creating test dirs: %v", err)
	}
	if err := os.MkdirAll(edenBinDir, 0755); err != nil {
		t.Fatalf("creating eden bin dir: %v", err)
	}
	if err := os.MkdirAll(mgbaBinDir, 0755); err != nil {
		t.Fatalf("creating mgba bin dir: %v", err)
	}

	realEdenBinary := filepath.Join(edenBinDir, "eden")
	if err := os.WriteFile(realEdenBinary, []byte("#!/bin/sh\necho eden"), 0755); err != nil {
		t.Fatalf("creating real eden binary: %v", err)
	}

	realMgbaBinary := filepath.Join(mgbaBinDir, "mgba")
	if err := os.WriteFile(realMgbaBinary, []byte("#!/bin/sh\necho mgba"), 0755); err != nil {
		t.Fatalf("creating real mgba binary: %v", err)
	}

	edenSymlink := filepath.Join(currentBinDir, "eden")
	if err := os.Symlink("/nix/store/xyz789-eden/bin/eden", edenSymlink); err != nil {
		t.Fatalf("creating eden symlink: %v", err)
	}

	mgbaSymlink := filepath.Join(currentBinDir, "mgba")
	if err := os.Symlink("/nix/store/def456-mgba/bin/mgba", mgbaSymlink); err != nil {
		t.Fatalf("creating mgba symlink: %v", err)
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
		{BinaryName: "eden"},
		{BinaryName: "mgba"},
	}

	if err := m.GenerateWrappers(emulators); err != nil {
		t.Fatalf("GenerateWrappers() error = %v", err)
	}

	t.Run("wrapper runs directly with real path", func(t *testing.T) {
		wrapperPath := filepath.Join(m.BinDir(), "eden")
		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("reading wrapper: %v", err)
		}

		wrapperStr := string(content)

		if strings.Contains(wrapperStr, "nix shell") {
			t.Errorf("wrapper should not use nix shell, got:\n%s", wrapperStr)
		}

		expectedRealPath := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-eden", "bin", "eden")
		expectedExec := `exec "` + expectedRealPath + `"`
		if !strings.Contains(wrapperStr, expectedExec) {
			t.Errorf("wrapper should exec real binary path %s, got:\n%s", expectedRealPath, wrapperStr)
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
			t.Error("hidden files should not be wrapped")
		}
	})
}

func TestGenerateWrappersSkipsRetroArchCoreWrappers(t *testing.T) {
	tmpDir := t.TempDir()

	profileDir := filepath.Join(tmpDir, "kyaraben")
	npLocation := filepath.Join(tmpDir, "nix-portable")
	currentDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "abc123-profile")
	currentBinDir := filepath.Join(currentDir, "bin")

	if err := os.MkdirAll(currentBinDir, 0755); err != nil {
		t.Fatalf("creating test dirs: %v", err)
	}

	retroarchSymlink := filepath.Join(currentBinDir, "retroarch")
	if err := os.Symlink("/nix/store/xyz-retroarch/bin/retroarch", retroarchSymlink); err != nil {
		t.Fatalf("creating retroarch symlink: %v", err)
	}

	retroarchBsnesSymlink := filepath.Join(currentBinDir, "retroarch-bsnes")
	if err := os.Symlink("/nix/store/xyz-bsnes/bin/retroarch-bsnes", retroarchBsnesSymlink); err != nil {
		t.Fatalf("creating retroarch-bsnes symlink: %v", err)
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

	if err := m.GenerateWrappers(nil); err != nil {
		t.Fatalf("GenerateWrappers() error = %v", err)
	}

	if _, err := os.Stat(filepath.Join(m.BinDir(), "retroarch")); os.IsNotExist(err) {
		t.Error("retroarch wrapper should be created")
	}

	if _, err := os.Stat(filepath.Join(m.BinDir(), "retroarch-bsnes")); !os.IsNotExist(err) {
		t.Error("retroarch-bsnes wrapper should NOT be created")
	}
}

func TestGenerateCoreSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	profileDir := filepath.Join(tmpDir, "kyaraben")
	npLocation := filepath.Join(tmpDir, "nix-portable")
	currentDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "abc123-profile")
	profileCoresDir := filepath.Join(currentDir, "lib", "retroarch", "cores")

	bsnesPkgDir := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-bsnes", "lib", "retroarch", "cores")

	if err := os.MkdirAll(profileCoresDir, 0755); err != nil {
		t.Fatalf("creating profile cores dir: %v", err)
	}
	if err := os.MkdirAll(bsnesPkgDir, 0755); err != nil {
		t.Fatalf("creating bsnes cores dir: %v", err)
	}

	realCorePath := filepath.Join(bsnesPkgDir, "bsnes_libretro.so")
	if err := os.WriteFile(realCorePath, []byte("fake core"), 0644); err != nil {
		t.Fatalf("creating real core: %v", err)
	}

	coreSymlink := filepath.Join(profileCoresDir, "bsnes_libretro.so")
	if err := os.Symlink("/nix/store/xyz789-bsnes/lib/retroarch/cores/bsnes_libretro.so", coreSymlink); err != nil {
		t.Fatalf("creating core symlink: %v", err)
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

	if err := m.GenerateCoreSymlinks(); err != nil {
		t.Fatalf("GenerateCoreSymlinks() error = %v", err)
	}

	coresDir := filepath.Join(tmpDir, "kyaraben", "cores")
	generatedSymlink := filepath.Join(coresDir, "bsnes_libretro.so")

	target, err := os.Readlink(generatedSymlink)
	if err != nil {
		t.Fatalf("reading generated symlink: %v", err)
	}

	expectedTarget := filepath.Join(npLocation, ".nix-portable", "nix", "store", "xyz789-bsnes", "lib", "retroarch", "cores", "bsnes_libretro.so")
	if target != expectedTarget {
		t.Errorf("symlink target = %q, want %q", target, expectedTarget)
	}
}
