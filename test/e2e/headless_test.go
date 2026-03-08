package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func (c *cliTest) initHeadlessNoSync() {
	c.t.Helper()
	_, err := c.run("init", "--headless", "--collection", c.collection)
	if err != nil {
		c.t.Fatalf("init --headless failed: %v", err)
	}
	config := c.readFile("config.toml")
	config = strings.Replace(config, "enabled = true", "enabled = false", 1)
	c.writeFile("config.toml", config)
}

func TestHeadlessInit(t *testing.T) {
	t.Run("creates headless config with sync enabled", func(t *testing.T) {
		c := newCLITest(t)

		output, err := c.run("init", "--headless", "--collection", c.collection)
		if err != nil {
			t.Fatalf("init --headless failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Headless mode")
		c.assertContains(output, "sync all systems")

		config := c.readFile("config.toml")
		c.assertContains(config, "headless = true")
		c.assertContains(config, "enabled = true")
	})

	t.Run("headless config has no systems section", func(t *testing.T) {
		c := newCLITest(t)

		_, err := c.run("init", "--headless", "--collection", c.collection)
		if err != nil {
			t.Fatalf("init --headless failed: %v", err)
		}

		config := c.readFile("config.toml")
		if strings.Contains(config, "[systems]") {
			t.Error("headless config should not have [systems] section")
		}
	})
}

func TestHeadlessApply(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping headless apply test in short mode")
	}

	t.Run("dry run does not create directories", func(t *testing.T) {
		c := newCLITest(t)
		c.initHeadlessNoSync()

		output, err := c.run("apply", "--dry-run")
		if err != nil {
			t.Fatalf("apply --dry-run failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Headless mode")
		c.assertContains(output, "Dry run")

		if _, err := os.Stat(c.collection); !os.IsNotExist(err) {
			t.Error("dry run should not create collection directory")
		}
	})

	t.Run("apply is idempotent", func(t *testing.T) {
		c := newCLITest(t)
		c.initHeadlessNoSync()

		_, err := c.run("apply")
		if err != nil {
			t.Fatalf("first apply failed: %v", err)
		}

		output, err := c.run("apply")
		if err != nil {
			t.Fatalf("second apply failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Done!")
	})

	t.Run("creates directories for all systems without installing emulators", func(t *testing.T) {
		c := newCLITest(t)
		c.initHeadlessNoSync()

		output, err := c.run("apply")
		if err != nil {
			t.Fatalf("apply failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "headless")

		romsDir := filepath.Join(c.collection, "roms")
		entries, err := os.ReadDir(romsDir)
		if err != nil {
			t.Fatalf("reading roms dir: %v", err)
		}

		if len(entries) < 10 {
			t.Errorf("expected at least 10 system directories in roms, got %d", len(entries))
		}

		expectedSystems := []string{"snes", "psx", "n64", "gba", "nds", "gamecube", "wii", "ps2", "psp"}
		for _, sys := range expectedSystems {
			sysPath := filepath.Join(romsDir, sys)
			if _, err := os.Stat(sysPath); os.IsNotExist(err) {
				t.Errorf("expected %s directory to exist", sysPath)
			}
		}

		savesDir := filepath.Join(c.collection, "saves")
		if _, err := os.Stat(savesDir); os.IsNotExist(err) {
			t.Error("expected saves directory to exist")
		}

		statesDir := filepath.Join(c.collection, "states")
		if _, err := os.Stat(statesDir); os.IsNotExist(err) {
			t.Error("expected states directory to exist")
		}
	})

	t.Run("does not install emulators in headless mode", func(t *testing.T) {
		c := newCLITest(t)
		c.initHeadlessNoSync()

		output, err := c.run("apply")
		if err != nil {
			t.Fatalf("apply failed: %v\nOutput: %s", err, output)
		}

		if strings.Contains(output, "Installing") || strings.Contains(output, "Downloading") {
			t.Error("headless mode should not install emulators")
		}

		statusOutput, err := c.run("status")
		if err != nil {
			t.Fatalf("status failed: %v\nOutput: %s", err, statusOutput)
		}

		c.assertContains(statusOutput, "Managed emulators: none")
	})
}

func TestHeadlessStatus(t *testing.T) {
	t.Run("shows headless mode in status", func(t *testing.T) {
		c := newCLITest(t)

		_, err := c.run("init", "--headless", "--collection", c.collection)
		if err != nil {
			t.Fatalf("init --headless failed: %v", err)
		}

		output, err := c.run("status")
		if err != nil {
			t.Fatalf("status failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "Enabled systems: none")
	})
}

func TestHeadlessDoctor(t *testing.T) {
	t.Run("doctor reports all provisions satisfied in headless mode", func(t *testing.T) {
		c := newCLITest(t)
		c.initHeadlessNoSync()

		output, err := c.run("doctor")
		if err != nil {
			t.Fatalf("doctor failed: %v\nOutput: %s", err, output)
		}

		c.assertContains(output, "All provisions satisfied")
		c.assertNotContains(output, "MISSING")
	})
}
