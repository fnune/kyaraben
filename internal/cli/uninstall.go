package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
)

// UninstallCmd removes kyaraben-managed files.
type UninstallCmd struct {
	Force bool `short:"f" help:"Skip confirmation prompt."`
}

// Run executes the uninstall command.
func (cmd *UninstallCmd) Run(ctx *Context) error {
	// Get paths
	dataDir, err := userDataDir()
	if err != nil {
		return err
	}
	kyarabenDataDir := filepath.Join(dataDir, "kyaraben")

	stateDir, err := userStateDir()
	if err != nil {
		return err
	}
	kyarabenStateDir := filepath.Join(stateDir, "kyaraben")

	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}
	configDir := filepath.Dir(configPath)

	// Load manifest to find managed configs
	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return err
	}
	manifest, _ := model.LoadManifest(manifestPath)

	// Show what will be removed
	fmt.Println("This will remove:")
	fmt.Println()

	if dirExists(kyarabenDataDir) {
		fmt.Printf("  %s (emulator store, flake)\n", kyarabenDataDir)
	}
	if dirExists(kyarabenStateDir) {
		fmt.Printf("  %s (manifest, state)\n", kyarabenStateDir)
	}

	// List managed config files
	if len(manifest.ManagedConfigs) > 0 {
		fmt.Println()
		fmt.Println("  Managed config files:")
		for _, cfg := range manifest.ManagedConfigs {
			if fileExists(cfg.Path) {
				fmt.Printf("    %s\n", cfg.Path)
			}
		}
	}

	fmt.Println()
	fmt.Println("This will NOT remove:")
	fmt.Printf("  %s (your ROMs, saves, BIOS)\n", "~/Emulation")
	fmt.Printf("  %s (your kyaraben config)\n", configDir)
	fmt.Println()

	// Confirm unless --force
	if !cmd.Force {
		fmt.Print("Proceed? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return fmt.Errorf("reading response: %w", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println("Cancelled.")
			return nil
		}
	}

	fmt.Println()
	fmt.Println("Removing kyaraben files...")

	// Remove managed config files
	for _, cfg := range manifest.ManagedConfigs {
		if fileExists(cfg.Path) {
			if err := os.Remove(cfg.Path); err != nil {
				fmt.Printf("  Warning: could not remove %s: %v\n", cfg.Path, err)
			} else {
				fmt.Printf("  Removed: %s\n", cfg.Path)
			}
		}
	}

	// Remove data directory
	if dirExists(kyarabenDataDir) {
		if err := os.RemoveAll(kyarabenDataDir); err != nil {
			fmt.Printf("  Warning: could not remove %s: %v\n", kyarabenDataDir, err)
		} else {
			fmt.Printf("  Removed: %s\n", kyarabenDataDir)
		}
	}

	// Remove state directory
	if dirExists(kyarabenStateDir) {
		if err := os.RemoveAll(kyarabenStateDir); err != nil {
			fmt.Printf("  Warning: could not remove %s: %v\n", kyarabenStateDir, err)
		} else {
			fmt.Printf("  Removed: %s\n", kyarabenStateDir)
		}
	}

	fmt.Println()
	fmt.Println("Done. Kyaraben files have been removed.")
	fmt.Println()
	fmt.Printf("To fully uninstall, also remove:\n")
	fmt.Printf("  %s (your config)\n", configDir)
	fmt.Println("  The kyaraben binary itself")

	return nil
}

func userDataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".local", "share"), nil
}

func userStateDir() (string, error) {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}
	return filepath.Join(home, ".local", "state"), nil
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
