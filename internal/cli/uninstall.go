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
	Force  bool `short:"f" help:"Skip confirmation prompt."`
	DryRun bool `short:"n" help:"Show what would be removed without doing anything."`
}

func (cmd *UninstallCmd) Run(ctx *Context) error {
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

	if dirExists(kyarabenStateDir) {
		fmt.Printf("  %s (nix store, manifest, state)\n", kyarabenStateDir)
	}

	if len(manifest.DesktopFiles) > 0 {
		fmt.Println()
		fmt.Println("  Desktop entries:")
		for _, f := range manifest.DesktopFiles {
			if fileExists(f) {
				fmt.Printf("    %s\n", f)
			}
		}
	}

	if len(manifest.IconFiles) > 0 {
		fmt.Println()
		fmt.Println("  Icons:")
		for _, f := range manifest.IconFiles {
			if fileExists(f) {
				fmt.Printf("    %s\n", f)
			}
		}
	}

	if len(manifest.ManagedConfigs) > 0 {
		fmt.Println()
		fmt.Println("  Managed config files:")
		for _, cfg := range manifest.ManagedConfigs {
			path, err := cfg.Target.Resolve()
			if err == nil && fileExists(path) {
				fmt.Printf("    %s\n", path)
			}
		}
	}

	fmt.Println()
	fmt.Println("This will NOT remove:")
	fmt.Printf("  %s (your ROMs, saves, BIOS)\n", userStore)
	fmt.Printf("  %s (your kyaraben config)\n", configDir)
	fmt.Println()
	if cmd.DryRun {
		return nil
	}
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
		path, err := cfg.Target.Resolve()
		if err != nil {
			continue
		}
		if fileExists(path) {
			if err := os.Remove(path); err != nil {
				fmt.Printf("  Warning: could not remove %s: %v\n", path, err)
			} else {
				fmt.Printf("  Removed: %s\n", path)
			}
		}
	}

	for _, f := range manifest.DesktopFiles {
		if fileExists(f) {
			if err := os.Remove(f); err != nil {
				fmt.Printf("  Warning: could not remove %s: %v\n", f, err)
			} else {
				fmt.Printf("  Removed: %s\n", f)
			}
		}
	}

	for _, f := range manifest.IconFiles {
		if fileExists(f) {
			if err := os.Remove(f); err != nil {
				fmt.Printf("  Warning: could not remove %s: %v\n", f, err)
			} else {
				fmt.Printf("  Removed: %s\n", f)
			}
		}
	}

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

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && !info.IsDir()
}
