package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type cliTest struct {
	t          *testing.T
	tmpDir     string
	configPath string
	userStore  string
	fakeStore  string
}

func newCLITest(t *testing.T) *cliTest {
	t.Helper()
	tmpDir := t.TempDir()
	return &cliTest{
		t:          t,
		tmpDir:     tmpDir,
		configPath: filepath.Join(tmpDir, "config.toml"),
		userStore:  filepath.Join(tmpDir, "Emulation"),
		fakeStore:  filepath.Join(tmpDir, "fake-store"),
	}
}

func (c *cliTest) run(args ...string) (string, error) {
	c.t.Helper()
	cmd := c.cmd(args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (c *cliTest) runWithFakeNix(args ...string) (string, error) {
	c.t.Helper()
	cmd := c.cmdWithFakeNix(args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (c *cliTest) runWithFakeNixFail(args ...string) (string, error) {
	c.t.Helper()
	cmd := c.cmdWithFakeNixFail(args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (c *cliTest) cmd(args ...string) *exec.Cmd {
	c.t.Helper()
	fullArgs := append([]string{"-c", c.configPath}, args...)
	return kyarabenCmd(c.t, c.tmpDir, fullArgs...)
}

func (c *cliTest) cmdWithFakeNix(args ...string) *exec.Cmd {
	c.t.Helper()
	fullArgs := append([]string{"-c", c.configPath}, args...)
	return kyarabenCmdWith(c.t, c.tmpDir, []cmdOption{withFakeNix(c.fakeStore)}, fullArgs...)
}

func (c *cliTest) cmdWithFakeNixFail(args ...string) *exec.Cmd {
	c.t.Helper()
	fullArgs := append([]string{"-c", c.configPath}, args...)
	return kyarabenCmdWith(c.t, c.tmpDir, []cmdOption{withFakeNixFail(c.fakeStore)}, fullArgs...)
}

func (c *cliTest) init() string {
	c.t.Helper()
	output, err := c.run("init", "-u", c.userStore)
	if err != nil {
		c.t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}
	return output
}

func (c *cliTest) apply() string {
	c.t.Helper()
	output, err := c.runWithFakeNix("apply")
	if err != nil {
		c.t.Fatalf("apply failed: %v\nOutput: %s", err, output)
	}
	return output
}

func (c *cliTest) assertFileExists(path string) {
	c.t.Helper()
	if _, err := os.Stat(path); err != nil {
		c.t.Errorf("expected file to exist: %s", path)
	}
}

func (c *cliTest) assertFileNotExists(path string) {
	c.t.Helper()
	if _, err := os.Stat(path); err == nil {
		c.t.Errorf("expected file to not exist: %s", path)
	}
}

func (c *cliTest) assertContains(output, substr string) {
	c.t.Helper()
	if !strings.Contains(output, substr) {
		c.t.Errorf("expected output to contain %q, got:\n%s", substr, output)
	}
}

func (c *cliTest) assertNotContains(output, substr string) {
	c.t.Helper()
	if strings.Contains(output, substr) {
		c.t.Errorf("expected output to not contain %q, got:\n%s", substr, output)
	}
}

func (c *cliTest) writeFile(relPath, content string) {
	c.t.Helper()
	path := filepath.Join(c.tmpDir, relPath)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		c.t.Fatalf("creating directory: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		c.t.Fatalf("writing file: %v", err)
	}
}

func (c *cliTest) readFile(relPath string) string {
	c.t.Helper()
	path := filepath.Join(c.tmpDir, relPath)
	data, err := os.ReadFile(path)
	if err != nil {
		c.t.Fatalf("reading file: %v", err)
	}
	return string(data)
}

func TestInit(t *testing.T) {
	t.Run("creates config with default systems", func(t *testing.T) {
		c := newCLITest(t)
		output := c.init()

		c.assertFileExists(c.configPath)
		c.assertContains(output, "Enabled")
	})

	t.Run("fails without force when config exists", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("init", "-u", c.userStore)
		if err == nil {
			t.Error("expected error when config already exists")
		}
		c.assertContains(output, "already exists")
	})

	t.Run("force overwrites existing config", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("init", "-u", c.userStore, "-f")
		if err != nil {
			t.Fatalf("init --force failed: %v\nOutput: %s", err, output)
		}
	})
}

func TestApplyWorkflow(t *testing.T) {
	t.Run("creates directories and completes", func(t *testing.T) {
		c := newCLITest(t)
		c.init()
		output := c.apply()

		c.assertContains(output, "Applying kyaraben configuration")
		c.assertContains(output, "Done!")
		c.assertFileExists(filepath.Join(c.userStore, "roms", "snes"))
		c.assertFileExists(filepath.Join(c.userStore, "roms", "gba"))
		c.assertFileExists(filepath.Join(c.userStore, "roms", "psx"))
	})

	t.Run("dry run does not modify filesystem", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("apply", "--dry-run")
		if err != nil {
			t.Fatalf("apply --dry-run failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(strings.ToLower(output), "dry run")
		c.assertFileNotExists(c.userStore)
	})

	t.Run("reapply is idempotent", func(t *testing.T) {
		c := newCLITest(t)
		c.init()
		c.apply()

		output := c.apply()
		c.assertContains(output, "Done!")
		c.assertNotContains(output, "managed keys will be overwritten")
	})

	t.Run("build failure is reported", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.runWithFakeNixFail("apply")
		if err == nil {
			t.Error("expected error when nix build fails")
		}
		c.assertContains(output, "failed")
	})
}

func TestApplyUserModifiedConfig(t *testing.T) {
	c := newCLITest(t)
	c.init()
	c.apply()

	mgbaConfig := filepath.Join(c.tmpDir, "xdg-config", "mgba", "config.ini")
	content := c.readFile("xdg-config/mgba/config.ini")
	modified := strings.Replace(content, c.userStore, "/user/modified/path", 1)
	if modified == content {
		t.Fatal("failed to modify config - no substitution made")
	}
	if err := os.WriteFile(mgbaConfig, []byte(modified), 0644); err != nil {
		t.Fatalf("writing modified config: %v", err)
	}

	cmd := c.cmdWithFakeNix("apply")
	cmd.Stdin = bytes.NewReader([]byte("n\n"))
	output, _ := cmd.CombinedOutput()

	c.assertContains(string(output), "managed keys will be overwritten")
	c.assertContains(string(output), "Cancelled")
}

func TestDoctor(t *testing.T) {
	t.Run("reports missing required provisions", func(t *testing.T) {
		c := newCLITest(t)
		c.init()
		_ = os.MkdirAll(filepath.Join(c.userStore, "bios", "psx"), 0755)

		output, _ := c.run("doctor")
		c.assertContains(output, "Checking provisions")
		c.assertContains(output, "scph5501.bin")
		c.assertContains(output, "MISSING")
	})

	t.Run("fails without config", func(t *testing.T) {
		c := newCLITest(t)
		_, err := c.run("doctor")
		if err == nil {
			t.Error("expected error when no config exists")
		}
	})
}

func TestStatus(t *testing.T) {
	t.Run("shows config info", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("status")
		if err != nil {
			t.Fatalf("status failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Config:")
		c.assertContains(output, "Emulation folder:")
	})

	t.Run("suggests init when no config", func(t *testing.T) {
		c := newCLITest(t)
		output, err := c.run("status")
		if err != nil {
			t.Fatalf("status failed unexpectedly: %v", err)
		}

		c.assertContains(output, "not found or invalid")
		c.assertContains(output, "kyaraben init")
	})
}

func TestUninstall(t *testing.T) {
	t.Run("shows what will be removed", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("uninstall", "-f")
		if err != nil {
			t.Fatalf("uninstall failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "This will remove")
		c.assertContains(output, "This will NOT remove")
	})

	t.Run("handles corrupted manifest gracefully", func(t *testing.T) {
		c := newCLITest(t)
		c.init()
		c.writeFile("xdg-state/kyaraben/build/manifest.json", "not valid json {{{")

		output, err := c.run("uninstall", "-n")
		if err != nil {
			t.Fatalf("uninstall failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Warning: could not load manifest")
		c.assertContains(output, "Some files may not be listed")
	})
}

func TestSync(t *testing.T) {
	t.Run("status shows disabled by default", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("sync", "status")
		if err != nil {
			t.Fatalf("sync status failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "disabled")
		c.assertContains(output, "Enable sync")
	})

	t.Run("help shows subcommands", func(t *testing.T) {
		c := newCLITest(t)
		output, err := c.run("sync", "--help")
		if err != nil {
			t.Fatalf("sync help failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "status")
		c.assertContains(output, "add-device")
		c.assertContains(output, "remove-device")
	})
}

func TestHelp(t *testing.T) {
	c := newCLITest(t)
	cmd := kyarabenCmd(c.t, c.tmpDir, "--help")
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("help failed: %v\nOutput: %s", err, output)
	}

	outputStr := string(output)
	c.assertContains(outputStr, "apply")
	c.assertContains(outputStr, "doctor")
	c.assertContains(outputStr, "status")
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
