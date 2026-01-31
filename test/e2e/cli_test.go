package e2e

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestCLIInit(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "e2e-test")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file not created: %v", err)
	}

	if !strings.Contains(string(output), "E2E Test") {
		t.Errorf("Output doesn't mention E2E Test: %s", output)
	}
}

func TestCLIStatus(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	// First init
	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "snes", "-s", "psx")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Run status command
	cmd = kyarabenCmd(t, "-c", configPath, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status failed: %v\nOutput: %s", err, output)
	}

	// Verify output contains expected information
	outputStr := string(output)
	if !strings.Contains(outputStr, "Config:") {
		t.Errorf("Output doesn't contain Config: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Emulation folder:") {
		t.Errorf("Output doesn't contain Emulation folder: %s", outputStr)
	}
}

func TestCLIDoctor(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	// Initialize with PSX (which has BIOS requirements)
	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "psx")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	_ = os.MkdirAll(filepath.Join(userStore, "bios", "psx"), 0755)

	// Run doctor command - should report missing BIOS
	cmd = kyarabenCmd(t, "-c", configPath, "doctor")
	output, _ := cmd.CombinedOutput() // Doctor exits with non-zero when files missing

	outputStr := string(output)
	if !strings.Contains(outputStr, "Checking provisions") {
		t.Errorf("Output doesn't show provision check: %s", outputStr)
	}
	if !strings.Contains(outputStr, "scph5501.bin") {
		t.Errorf("Output doesn't mention PSX BIOS file: %s", outputStr)
	}
	if !strings.Contains(outputStr, "MISSING") || !strings.Contains(outputStr, "required") {
		t.Errorf("Output doesn't indicate missing required file: %s", outputStr)
	}
}

func TestCLIDoctorE2ETest(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "e2e-test")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	_ = os.MkdirAll(filepath.Join(userStore, "bios", "e2e-test"), 0755)

	cmd = kyarabenCmd(t, "-c", configPath, "doctor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor failed for E2E Test: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "No provisions required") {
		t.Errorf("Output should indicate no provisions required: %s", outputStr)
	}
}

func TestCLIApplyDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "e2e-test")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, "-c", configPath, "apply", "--dry-run")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("apply --dry-run failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Applying kyaraben configuration") {
		t.Errorf("Output doesn't show apply started: %s", outputStr)
	}
	if !strings.Contains(strings.ToLower(outputStr), "dry run") {
		t.Errorf("Output doesn't mention dry run: %s", outputStr)
	}

	if _, err := os.Stat(userStore); err == nil {
		t.Errorf("UserStore should not be created in dry run")
	}
}

func TestCLIHelp(t *testing.T) {
	cmd := kyarabenCmd(t, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "apply") {
		t.Errorf("Help doesn't mention apply command: %s", outputStr)
	}
	if !strings.Contains(outputStr, "doctor") {
		t.Errorf("Help doesn't mention doctor command: %s", outputStr)
	}
	if !strings.Contains(outputStr, "status") {
		t.Errorf("Help doesn't mention status command: %s", outputStr)
	}
}

func TestCLIInitForce(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "e2e-test")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "snes")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("init without --force should fail when config exists")
	}
	if !strings.Contains(string(output), "already exists") {
		t.Errorf("Error message should mention config already exists: %s", output)
	}

	cmd = kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "snes", "-f")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init --force failed: %v\nOutput: %s", err, output)
	}
}

func kyarabenCmd(t *testing.T, args ...string) *exec.Cmd {
	t.Helper()

	// Find the binary - look in project root first
	binary := filepath.Join(projectRoot(t), "kyaraben")
	if _, err := os.Stat(binary); err != nil {
		t.Fatalf("kyaraben binary not found at %s. Run 'go build -o ./kyaraben ./cmd/kyaraben' first", binary)
	}

	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()

	return cmd
}

func projectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up to find go.mod
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatalf("Could not find project root (go.mod)")
		}
		dir = parent
	}
}

func TestCLIUninstall(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "e2e-test")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, "-c", configPath, "uninstall", "-f")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("uninstall failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "This will remove") {
		t.Errorf("Output doesn't show what will be removed: %s", outputStr)
	}
	if !strings.Contains(outputStr, "This will NOT remove") {
		t.Errorf("Output doesn't show what will be preserved: %s", outputStr)
	}
}

func TestCLISyncStatusDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, "-c", configPath, "init", "-u", userStore, "-s", "e2e-test")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, "-c", configPath, "sync", "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "disabled") {
		t.Errorf("Output should indicate sync is disabled: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Enable sync") {
		t.Errorf("Output should show how to enable sync: %s", outputStr)
	}
}

func TestCLISyncHelp(t *testing.T) {
	cmd := kyarabenCmd(t, "sync", "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("sync help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "status") {
		t.Errorf("Sync help doesn't mention status: %s", outputStr)
	}
	if !strings.Contains(outputStr, "add-device") {
		t.Errorf("Sync help doesn't mention add-device: %s", outputStr)
	}
	if !strings.Contains(outputStr, "remove-device") {
		t.Errorf("Sync help doesn't mention remove-device: %s", outputStr)
	}
}
