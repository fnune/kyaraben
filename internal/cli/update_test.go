package cli

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestExpectedAssetNameMatchesReleaseScript(t *testing.T) {
	if runtime.GOOS != "linux" || runtime.GOARCH != "amd64" {
		t.Skip("release script only builds linux/amd64")
	}

	script, err := os.ReadFile(filepath.Join("..", "..", "scripts", "build-release-cli.sh"))
	if err != nil {
		t.Fatalf("reading build script: %v", err)
	}

	name := expectedAssetName()
	if !strings.Contains(string(script), name) {
		t.Errorf("build-release-cli.sh does not produce %q, so `kyaraben update` cannot find its asset", name)
	}
}

func TestInstallUpdateReplacesBinary(t *testing.T) {
	dir := t.TempDir()

	binaryPath := filepath.Join(dir, "kyaraben")
	if err := os.WriteFile(binaryPath, []byte("old"), 0755); err != nil {
		t.Fatal(err)
	}

	downloadPath := filepath.Join(dir, "download")
	if err := os.WriteFile(downloadPath, []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}

	if err := installUpdate(downloadPath, binaryPath); err != nil {
		t.Fatalf("installing update: %v", err)
	}

	installed, err := os.ReadFile(binaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(installed) != "new" {
		t.Errorf("binary contents = %q, want new", installed)
	}

	info, err := os.Stat(binaryPath)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm()&0111 == 0 {
		t.Errorf("installed binary is not executable, mode is %v", info.Mode().Perm())
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), "kyaraben-new-") {
			t.Errorf("temp file %s left behind", entry.Name())
		}
	}
}
