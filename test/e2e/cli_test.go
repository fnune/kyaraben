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
	collection string
}

func TestMain(m *testing.M) {
	if testing.Short() {
		os.Exit(0)
	}

	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			panic("go.mod not found")
		}
		root = parent
	}

	buildCmd := exec.Command("go", "build", "-o", filepath.Join(root, "kyaraben"), "./cmd/kyaraben")
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	buildCmd.Dir = root
	if err := buildCmd.Run(); err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func newCLITest(t *testing.T) *cliTest {
	t.Helper()
	tmpDir := t.TempDir()
	return &cliTest{
		t:          t,
		tmpDir:     tmpDir,
		configPath: filepath.Join(tmpDir, "config.toml"),
		collection: filepath.Join(tmpDir, "Emulation"),
	}
}

func (c *cliTest) run(args ...string) (string, error) {
	c.t.Helper()
	cmd := c.cmd(args...)
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func (c *cliTest) cmd(args ...string) *exec.Cmd {
	c.t.Helper()
	fullArgs := append([]string{"-c", c.configPath}, args...)
	return kyarabenCmd(c.t, c.tmpDir, fullArgs...)
}

func (c *cliTest) init() string {
	c.t.Helper()
	output, err := c.run("init", "--collection", c.collection)
	if err != nil {
		c.t.Fatalf("init failed: %v\nOutput: %s", err, output)
	}
	return output
}

func (c *cliTest) apply() string {
	c.t.Helper()
	output, err := c.run("apply")
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

		output, err := c.run("init", "--collection", c.collection)
		if err == nil {
			t.Error("expected error when config already exists")
		}
		c.assertContains(output, "already exists")
	})

	t.Run("force overwrites existing config", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("init", "--collection", c.collection, "-f")
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
		c.assertFileExists(filepath.Join(c.collection, "roms", "snes"))
		c.assertFileExists(filepath.Join(c.collection, "roms", "gba"))
		c.assertFileExists(filepath.Join(c.collection, "roms", "psx"))
	})

	t.Run("dry run does not modify filesystem", func(t *testing.T) {
		c := newCLITest(t)
		c.init()

		output, err := c.run("apply", "--dry-run")
		if err != nil {
			t.Fatalf("apply --dry-run failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(strings.ToLower(output), "dry run")
		c.assertFileNotExists(c.collection)
	})

	t.Run("reapply is idempotent", func(t *testing.T) {
		c := newCLITest(t)
		c.init()
		c.apply()

		output := c.apply()
		c.assertContains(output, "Done!")
		c.assertNotContains(output, "managed settings will be overwritten")
	})

	t.Run("build failure is reported", func(t *testing.T) {
		t.Skip("installer failure simulation removed")
	})
}

func TestApplyUserModifiedConfig(t *testing.T) {
	c := newCLITest(t)
	c.init()
	c.apply()

	retroarchConfig := filepath.Join(c.tmpDir, "xdg-config", "retroarch", "retroarch.cfg")
	content := c.readFile("xdg-config/retroarch/retroarch.cfg")
	modified := strings.Replace(content, c.collection, "/user/modified/path", 1)
	if modified == content {
		t.Fatal("failed to modify config - no substitution made")
	}
	if err := os.WriteFile(retroarchConfig, []byte(modified), 0644); err != nil {
		t.Fatalf("writing modified config: %v", err)
	}

	cmd := c.cmd("apply")
	cmd.Stdin = bytes.NewReader([]byte("n\n"))
	output, _ := cmd.CombinedOutput()

	c.assertContains(string(output), "managed settings will be overwritten")
	c.assertContains(string(output), "Cancelled")
}

func TestApplyUserModifiedConfigXML(t *testing.T) {
	c := newCLITest(t)

	configContent := `[global]
collection = "` + c.collection + `"

[systems]
wiiu = ["cemu"]
`
	c.writeFile("config.toml", configContent)
	c.apply()

	cemuConfig := filepath.Join(c.tmpDir, "xdg-config", "Cemu", "settings.xml")
	content := c.readFile("xdg-config/Cemu/settings.xml")

	if !strings.Contains(content, "<check_update>false</check_update>") {
		t.Fatalf("expected check_update=false in config, got:\n%s", content)
	}

	modified := strings.Replace(content, "<check_update>false</check_update>", "<check_update>true</check_update>", 1)
	if modified == content {
		t.Fatal("failed to modify config - no substitution made")
	}
	if err := os.WriteFile(cemuConfig, []byte(modified), 0644); err != nil {
		t.Fatalf("writing modified config: %v", err)
	}

	cmd := c.cmd("apply")
	cmd.Stdin = bytes.NewReader([]byte("n\n"))
	output, _ := cmd.CombinedOutput()

	c.assertContains(string(output), "managed settings will be overwritten")
	c.assertContains(string(output), "check_update")
	c.assertContains(string(output), "Cancelled")
}

func TestApplyNintendoConfirmChangeIsNotConflict(t *testing.T) {
	c := newCLITest(t)

	configContent := `[global]
collection = "` + c.collection + `"

[controller]
nintendo_confirm = "east"

[systems]
gamecube = ["dolphin"]
`
	c.writeFile("config.toml", configContent)
	c.apply()

	dolphinGCPadConfig := filepath.Join(c.tmpDir, "xdg-config", "dolphin-emu", "GCPadNew.ini")
	c.assertFileExists(dolphinGCPadConfig)
	originalContent := c.readFile("xdg-config/dolphin-emu/GCPadNew.ini")
	if !strings.Contains(originalContent, "Buttons/A") {
		t.Fatalf("expected Buttons/A in config, got:\n%s", originalContent)
	}

	configContent = `[global]
collection = "` + c.collection + `"

[controller]
nintendo_confirm = "south"

[systems]
gamecube = ["dolphin"]
`
	c.writeFile("config.toml", configContent)

	output := c.apply()

	c.assertNotContains(output, "managed settings will be overwritten")
	c.assertNotContains(output, "Your changes")
	c.assertContains(output, "Done!")
}

func TestDoctor(t *testing.T) {
	t.Run("reports missing required provisions", func(t *testing.T) {
		c := newCLITest(t)
		c.init()
		_ = os.MkdirAll(filepath.Join(c.collection, "bios", "psx"), 0755)

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
		c.assertContains(output, "Collection:")
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
		c.assertContains(output, "pair")
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
		"KYARABEN_E2E_FAKE_INSTALLER=1",
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
