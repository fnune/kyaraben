package e2e

import (
	"bytes"
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

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	if _, err := os.Stat(configPath); err != nil {
		t.Errorf("Config file not created: %v", err)
	}

	if !strings.Contains(string(output), "Game Boy Advance") {
		t.Errorf("Output doesn't mention Game Boy Advance: %s", output)
	}
}

func TestCLIStatus(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	// First init
	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes", "-s", "psx")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Run status command
	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "status")
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
	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "psx")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	_ = os.MkdirAll(filepath.Join(userStore, "bios", "psx"), 0755)

	// Run doctor command - should report missing BIOS
	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "doctor")
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

func TestCLIDoctorGBA(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	_ = os.MkdirAll(filepath.Join(userStore, "bios", "gba"), 0755)

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "doctor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor failed for GBA: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	// GBA BIOS is optional, so doctor should pass (exit 0) but may report optional files missing
	if strings.Contains(outputStr, "required") && strings.Contains(outputStr, "MISSING") {
		t.Errorf("No required provisions should be missing for GBA: %s", outputStr)
	}
}

func TestCLIApplyDryRun(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "apply", "--dry-run")
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
	tmpDir := t.TempDir()
	cmd := kyarabenCmd(t, tmpDir, "--help")
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

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("init without --force should fail when config exists")
	}
	if !strings.Contains(string(output), "already exists") {
		t.Errorf("Error message should mention config already exists: %s", output)
	}

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes", "-f")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init --force failed: %v\nOutput: %s", err, output)
	}
}

type cmdOption func(*exec.Cmd, *testing.T)

func withFakeNix(storeDir string) cmdOption {
	return func(cmd *exec.Cmd, t *testing.T) {
		t.Helper()
		fakeNixPath := filepath.Join(projectRoot(t), "ui", "e2e", "fixtures", "fake-nix-portable")
		if _, err := os.Stat(fakeNixPath); err != nil {
			t.Fatalf("fake-nix-portable not found at %s", fakeNixPath)
		}
		cmd.Env = append(cmd.Env,
			"KYARABEN_NIX_PORTABLE_PATH="+fakeNixPath,
			"FAKE_NIX_STORE="+storeDir,
		)
	}
}

func withFakeNixFail(storeDir string) cmdOption {
	return func(cmd *exec.Cmd, t *testing.T) {
		t.Helper()
		fakeNixPath := filepath.Join(projectRoot(t), "ui", "e2e", "fixtures", "fake-nix-portable")
		if _, err := os.Stat(fakeNixPath); err != nil {
			t.Fatalf("fake-nix-portable not found at %s", fakeNixPath)
		}
		cmd.Env = append(cmd.Env,
			"KYARABEN_NIX_PORTABLE_PATH="+fakeNixPath,
			"FAKE_NIX_STORE="+storeDir,
			"FAKE_NIX_FAIL=1",
		)
	}
}

func kyarabenCmd(t *testing.T, tmpDir string, args ...string) *exec.Cmd {
	return kyarabenCmdWith(t, tmpDir, nil, args...)
}

func kyarabenCmdWith(t *testing.T, tmpDir string, opts []cmdOption, args ...string) *exec.Cmd {
	t.Helper()

	binary := filepath.Join(projectRoot(t), "kyaraben")
	if _, err := os.Stat(binary); err != nil {
		t.Fatalf("kyaraben binary not found at %s. Run 'go build -o ./kyaraben ./cmd/kyaraben' first", binary)
	}

	cmd := exec.Command(binary, args...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env,
		"HOME="+tmpDir,
		"XDG_CONFIG_HOME="+filepath.Join(tmpDir, "xdg-config"),
		"XDG_STATE_HOME="+filepath.Join(tmpDir, "xdg-state"),
		"XDG_DATA_HOME="+filepath.Join(tmpDir, "xdg-data"),
	)

	for _, opt := range opts {
		opt(cmd, t)
	}

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

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "uninstall", "-f")
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

func TestCLIUninstallCorruptedManifest(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// Write corrupted manifest to the isolated XDG_STATE_HOME
	stateDir := filepath.Join(tmpDir, "xdg-state")
	manifestDir := filepath.Join(stateDir, "kyaraben", "build")
	manifestPath := filepath.Join(manifestDir, "manifest.json")
	_ = os.MkdirAll(manifestDir, 0755)
	_ = os.WriteFile(manifestPath, []byte("not valid json {{{"), 0644)

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "uninstall", "-n")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("uninstall failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Warning: could not load manifest") {
		t.Errorf("Output should warn about corrupted manifest: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Some files may not be listed") {
		t.Errorf("Output should mention files may not be listed: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Manifest path:") {
		t.Errorf("Output should show manifest path: %s", outputStr)
	}
	if !strings.Contains(outputStr, "not valid json") {
		t.Errorf("Output should show manifest contents: %s", outputStr)
	}
}

func TestCLISyncStatusDisabled(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "sync", "status")
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
	tmpDir := t.TempDir()
	cmd := kyarabenCmd(t, tmpDir, "sync", "--help")
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

func TestCLIStatusNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.toml")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "status")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("status command failed unexpectedly: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "not found or invalid") {
		t.Errorf("Output should mention config not found: %s", outputStr)
	}
	if !strings.Contains(outputStr, "kyaraben init") {
		t.Errorf("Output should suggest running init: %s", outputStr)
	}
}

func TestCLIDoctorNoConfig(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "nonexistent", "config.toml")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "doctor")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("doctor without config should fail")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "loading config") {
		t.Errorf("Output should mention config loading issue: %s", outputStr)
	}
}

func TestCLIInitInvalidSystem(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "nonexistent-system")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("init with invalid system should fail")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "unknown system") {
		t.Errorf("Output should indicate unknown system: %s", outputStr)
	}
}

func TestCLIDoctorAllProvisionsSatisfied(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	_ = os.MkdirAll(filepath.Join(userStore, "bios", "snes"), 0755)

	cmd = kyarabenCmd(t, tmpDir, "-c", configPath, "doctor")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("doctor failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "No provisions required") {
		t.Errorf("Output should show no provisions required for SNES: %s", outputStr)
	}
}

func TestCLIApply(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")
	fakeStore := filepath.Join(tmpDir, "fake-store")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNix(fakeStore)},
		"-c", configPath, "apply", "--no-show-diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("apply failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "Applying kyaraben configuration") {
		t.Errorf("Output doesn't show apply started: %s", outputStr)
	}
	if !strings.Contains(outputStr, "Done!") {
		t.Errorf("Output doesn't show completion: %s", outputStr)
	}

	if _, err := os.Stat(filepath.Join(userStore, "roms", "snes")); err != nil {
		t.Errorf("ROM directory not created: %v", err)
	}
}

func TestCLIApplyBuildFailure(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")
	fakeStore := filepath.Join(tmpDir, "fake-store")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNixFail(fakeStore)},
		"-c", configPath, "apply", "--no-show-diff")
	output, err := cmd.CombinedOutput()
	if err == nil {
		t.Errorf("apply should fail when nix build fails")
	}

	outputStr := string(output)
	if !strings.Contains(outputStr, "failed") {
		t.Errorf("Output should mention build failure: %s", outputStr)
	}
}

func TestCLIApplyMultipleSystems(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")
	fakeStore := filepath.Join(tmpDir, "fake-store")

	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "snes", "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNix(fakeStore)},
		"-c", configPath, "apply", "--no-show-diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("apply failed: %v\nOutput: %s", err, output)
	}

	for _, system := range []string{"snes", "gba"} {
		romDir := filepath.Join(userStore, "roms", system)
		if _, err := os.Stat(romDir); err != nil {
			t.Errorf("ROM directory for %s not created: %v", system, err)
		}
	}
}

func TestCLIApplyPromptsOnUserModifiedKeys(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")
	fakeStore := filepath.Join(tmpDir, "fake-store")

	// Init with GBA (uses mGBA which has managed config keys)
	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	// First apply to create baseline config and manifest
	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNix(fakeStore)},
		"-c", configPath, "apply", "--no-show-diff")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("first apply failed: %v\nOutput: %s", err, output)
	}

	// Modify a managed config key (mGBA config at XDG_CONFIG_HOME/mgba/config.ini)
	mgbaConfig := filepath.Join(tmpDir, "xdg-config", "mgba", "config.ini")
	data, err := os.ReadFile(mgbaConfig)
	if err != nil {
		t.Fatalf("reading mGBA config: %v", err)
	}

	// Change the value of the "bios" managed key
	lines := strings.Split(string(data), "\n")
	found := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "bios") && strings.Contains(trimmed, "=") {
			lines[i] = "bios = /user/modified/path"
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("could not find bios key in mGBA config:\n%s", string(data))
	}
	if err := os.WriteFile(mgbaConfig, []byte(strings.Join(lines, "\n")), 0644); err != nil {
		t.Fatalf("modifying mGBA config: %v", err)
	}

	// Run apply again with "n" to decline the overwrite prompt
	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNix(fakeStore)},
		"-c", configPath, "apply")
	cmd.Stdin = bytes.NewReader([]byte("n\n"))
	output, err := cmd.CombinedOutput()
	// The command exits cleanly on cancel (no error), but check just in case
	_ = err

	outputStr := string(output)
	if !strings.Contains(outputStr, "managed keys will be overwritten") {
		t.Errorf("expected overwrite prompt, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Cancelled") {
		t.Errorf("expected cancellation message, got:\n%s", outputStr)
	}
}

func TestCLIApplyNoPromptWithoutUserChanges(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.toml")
	userStore := filepath.Join(tmpDir, "Emulation")
	fakeStore := filepath.Join(tmpDir, "fake-store")

	// Init and first apply with GBA
	cmd := kyarabenCmd(t, tmpDir, "-c", configPath, "init", "-u", userStore, "-s", "gba")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}

	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNix(fakeStore)},
		"-c", configPath, "apply", "--no-show-diff")
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("first apply failed: %v\nOutput: %s", err, output)
	}

	// Apply again without modifying any files - should NOT prompt
	cmd = kyarabenCmdWith(t, tmpDir, []cmdOption{withFakeNix(fakeStore)},
		"-c", configPath, "apply", "--no-show-diff")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("second apply failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	if strings.Contains(outputStr, "managed keys will be overwritten") {
		t.Errorf("should not prompt when no user changes, got:\n%s", outputStr)
	}
	if !strings.Contains(outputStr, "Done!") {
		t.Errorf("expected completion, got:\n%s", outputStr)
	}
}
