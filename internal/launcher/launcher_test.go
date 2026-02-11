package launcher

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateWrappers(t *testing.T) {
	tmpDir := t.TempDir()
	profileDir := filepath.Join(tmpDir, "kyaraben")

	edenBinaryPath := filepath.Join(tmpDir, "packages", "eden", "bin", "eden")
	mgbaBinaryPath := filepath.Join(tmpDir, "packages", "mgba", "bin", "mgba")
	if err := os.MkdirAll(filepath.Dir(edenBinaryPath), 0755); err != nil {
		t.Fatalf("creating eden dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(mgbaBinaryPath), 0755); err != nil {
		t.Fatalf("creating mgba dir: %v", err)
	}
	if err := os.WriteFile(edenBinaryPath, []byte("#!/bin/sh\necho eden"), 0755); err != nil {
		t.Fatalf("creating eden binary: %v", err)
	}
	if err := os.WriteFile(mgbaBinaryPath, []byte("#!/bin/sh\necho mgba"), 0755); err != nil {
		t.Fatalf("creating mgba binary: %v", err)
	}

	m := &Manager{profileDir: profileDir}

	binaries := []InstalledBinary{
		{Name: "eden", Path: edenBinaryPath},
		{Name: "mgba", Path: mgbaBinaryPath},
	}

	if err := m.GenerateWrappers(binaries); err != nil {
		t.Fatalf("GenerateWrappers() error = %v", err)
	}

	t.Run("wrapper executes real binary path", func(t *testing.T) {
		wrapperPath := filepath.Join(m.BinDir(), "eden")
		content, err := os.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("reading wrapper: %v", err)
		}

		wrapperStr := string(content)
		expectedExec := `exec "` + edenBinaryPath + `"`
		if !strings.Contains(wrapperStr, expectedExec) {
			t.Errorf("wrapper should exec real binary path %s, got:\n%s", edenBinaryPath, wrapperStr)
		}

		info, err := os.Stat(wrapperPath)
		if err != nil {
			t.Fatalf("stat wrapper: %v", err)
		}
		if info.Mode()&0111 == 0 {
			t.Error("wrapper should be executable")
		}
	})

	t.Run("hidden binaries are skipped", func(t *testing.T) {
		binaries := []InstalledBinary{
			{Name: ".hidden", Path: "/some/path"},
			{Name: "eden", Path: edenBinaryPath},
		}

		if err := m.GenerateWrappers(binaries); err != nil {
			t.Fatalf("GenerateWrappers() error = %v", err)
		}

		hiddenWrapperPath := filepath.Join(m.BinDir(), ".hidden")
		if _, err := os.Stat(hiddenWrapperPath); !os.IsNotExist(err) {
			t.Error("hidden binaries should not be wrapped")
		}
	})
}

func TestGenerateWrappersEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{profileDir: filepath.Join(tmpDir, "kyaraben")}

	if err := m.GenerateWrappers(nil); err != nil {
		t.Fatalf("GenerateWrappers(nil) error = %v", err)
	}

	if _, err := os.Stat(m.BinDir()); os.IsNotExist(err) {
		t.Error("bin dir should be created even for empty binaries list")
	}
}

func TestGenerateCoreSymlinks(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_STATE_HOME", tmpDir)

	profileDir := filepath.Join(tmpDir, "kyaraben")
	m := &Manager{profileDir: profileDir}

	corePath := filepath.Join(tmpDir, "packages", "retroarch-cores", "lib", "retroarch", "cores", "bsnes_libretro.so")
	if err := os.MkdirAll(filepath.Dir(corePath), 0755); err != nil {
		t.Fatalf("creating core dir: %v", err)
	}
	if err := os.WriteFile(corePath, []byte("fake core"), 0644); err != nil {
		t.Fatalf("creating core: %v", err)
	}

	cores := []InstalledCore{
		{Filename: "bsnes_libretro.so", Path: corePath},
	}

	if err := m.GenerateCoreSymlinks(cores); err != nil {
		t.Fatalf("GenerateCoreSymlinks() error = %v", err)
	}

	coresDir := filepath.Join(tmpDir, "kyaraben", "cores")
	generatedSymlink := filepath.Join(coresDir, "bsnes_libretro.so")

	target, err := os.Readlink(generatedSymlink)
	if err != nil {
		t.Fatalf("reading generated symlink: %v", err)
	}

	if target != corePath {
		t.Errorf("symlink target = %q, want %q", target, corePath)
	}
}

func TestGenerateCoreSymlinksEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	m := &Manager{profileDir: filepath.Join(tmpDir, "kyaraben")}

	if err := m.GenerateCoreSymlinks(nil); err != nil {
		t.Fatalf("GenerateCoreSymlinks(nil) error = %v", err)
	}
}
