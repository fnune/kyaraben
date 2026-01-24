package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

type UninstallCmd struct {
	Force bool `short:"f" help:"Skip confirmation prompt."`
}

func (cmd *UninstallCmd) Run(ctx *Context) error {
	kyarabenDataDir, err := paths.KyarabenDataDir()
	if err != nil {
		return err
	}

	kyarabenStateDir, err := paths.KyarabenStateDir()
	if err != nil {
		return err
	}

	configPath, err := ctx.GetConfigPath()
	if err != nil {
		return err
	}
	configDir := filepath.Dir(configPath)

	manifestPath, err := model.DefaultManifestPath()
	if err != nil {
		return err
	}
	manifest, _ := model.LoadManifest(manifestPath)

	cfg, _ := ctx.LoadConfig()
	userStore := "~/Emulation"
	if cfg != nil {
		userStore = cfg.Global.UserStore
	}

	fmt.Println("This will remove:")
	fmt.Println()

	if dirExists(kyarabenDataDir) {
		fmt.Printf("  %s (emulator store, flake)\n", kyarabenDataDir)
	}
	if dirExists(kyarabenStateDir) {
		fmt.Printf("  %s (manifest, state)\n", kyarabenStateDir)
	}

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
	fmt.Printf("  %s (your ROMs, saves, BIOS)\n", userStore)
	fmt.Printf("  %s (your kyaraben config)\n", configDir)
	fmt.Println()
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

	for _, cfg := range manifest.ManagedConfigs {
		if fileExists(cfg.Path) {
			if err := os.Remove(cfg.Path); err != nil {
				fmt.Printf("  Warning: could not remove %s: %v\n", cfg.Path, err)
			} else {
				fmt.Printf("  Removed: %s\n", cfg.Path)
			}
		}
	}

	for _, dir := range []string{kyarabenDataDir, kyarabenStateDir} {
		if dirExists(dir) {
			if err := os.RemoveAll(dir); err != nil {
				fmt.Printf("  Warning: could not remove %s: %v\n", dir, err)
			} else {
				fmt.Printf("  Removed: %s\n", dir)
			}
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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
