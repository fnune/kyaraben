package launcher

import (
	"os"
	"strings"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"
)

func TestGenerateWrappers(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben":               &vfst.Dir{Perm: 0755},
		"/packages/eden/bin/eden": "#!/bin/sh\necho eden",
		"/packages/mgba/bin/mgba": "#!/bin/sh\necho mgba",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	profileDir := "/kyaraben"
	edenBinaryPath := "/packages/eden/bin/eden"
	mgbaBinaryPath := "/packages/mgba/bin/mgba"

	m := &Manager{fs: fs, profileDir: profileDir}

	binaries := []InstalledBinary{
		{Name: "eden", Path: edenBinaryPath},
		{Name: "mgba", Path: mgbaBinaryPath},
	}

	if err := m.GenerateWrappers(binaries); err != nil {
		t.Fatalf("GenerateWrappers() error = %v", err)
	}

	t.Run("wrapper executes real binary path", func(t *testing.T) {
		wrapperPath := m.BinDir() + "/eden"
		content, err := fs.ReadFile(wrapperPath)
		if err != nil {
			t.Fatalf("reading wrapper: %v", err)
		}

		wrapperStr := string(content)
		expectedExec := `exec "` + edenBinaryPath + `"`
		if !strings.Contains(wrapperStr, expectedExec) {
			t.Errorf("wrapper should exec real binary path %s, got:\n%s", edenBinaryPath, wrapperStr)
		}

		info, err := fs.Stat(wrapperPath)
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

		hiddenWrapperPath := m.BinDir() + "/.hidden"
		if _, err := fs.Stat(hiddenWrapperPath); err == nil {
			t.Error("hidden binaries should not be wrapped")
		}
	})
}

func TestGenerateWrappersEmpty(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	m := &Manager{fs: fs, profileDir: "/kyaraben"}

	if err := m.GenerateWrappers(nil); err != nil {
		t.Fatalf("GenerateWrappers(nil) error = %v", err)
	}

	if _, err := fs.Stat(m.BinDir()); err != nil {
		t.Error("bin dir should be created even for empty binaries list")
	}
}

func TestGenerateCoreSymlinksEmpty(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	m := &Manager{fs: fs, profileDir: "/kyaraben"}

	if err := m.GenerateCoreSymlinks(nil); err != nil {
		t.Fatalf("GenerateCoreSymlinks(nil) error = %v", err)
	}
}

func TestGenerateCoreSymlinks(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben":                      &vfst.Dir{Perm: 0755},
		"/packages/retroarch/cores/a.so": "fake core a",
		"/packages/retroarch/cores/b.so": "fake core b",
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	m := &Manager{fs: fs, profileDir: "/kyaraben"}

	cores := []InstalledCore{
		{Filename: "core_a.so", Path: "/packages/retroarch/cores/a.so"},
		{Filename: "core_b.so", Path: "/packages/retroarch/cores/b.so"},
	}

	if err := m.GenerateCoreSymlinks(cores); err != nil {
		t.Fatalf("GenerateCoreSymlinks() error = %v", err)
	}

	coresDir := m.CoresDir()

	info, err := fs.Lstat(coresDir + "/core_a.so")
	if err != nil {
		t.Fatalf("symlink core_a.so not found: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("core_a.so should be a symlink")
	}

	info, err = fs.Lstat(coresDir + "/core_b.so")
	if err != nil {
		t.Fatalf("symlink core_b.so not found: %v", err)
	}
	if info.Mode()&os.ModeSymlink == 0 {
		t.Error("core_b.so should be a symlink")
	}
}
