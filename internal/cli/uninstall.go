package cli

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

	if manifest.KyarabenInstall != nil {
		ki := manifest.KyarabenInstall
		if fileExists(ki.AppPath) || fileExists(ki.CLIPath) || fileExists(ki.DesktopPath) {
			fmt.Println()
			fmt.Println("  Kyaraben installation:")
			if fileExists(ki.AppPath) {
				fmt.Printf("    %s\n", ki.AppPath)
			}
			if fileExists(ki.CLIPath) {
				fmt.Printf("    %s\n", ki.CLIPath)
			}
			if fileExists(ki.DesktopPath) {
				fmt.Printf("    %s\n", ki.DesktopPath)
			}
		}
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

	if manifest.KyarabenInstall != nil {
		ki := manifest.KyarabenInstall
		for _, path := range []string{ki.AppPath, ki.CLIPath, ki.DesktopPath} {
			if path != "" && fileExists(path) {
				if err := os.Remove(path); err != nil {
					fmt.Printf("  Warning: could not remove %s: %v\n", path, err)
				} else {
					fmt.Printf("  Removed: %s\n", path)
				}
			}
		}
	}

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
		if err := forceRemoveAll(kyarabenStateDir); err != nil {
			fmt.Printf("  Warning: could not remove %s: %v\n", kyarabenStateDir, err)
		} else {
			fmt.Printf("  Removed: %s\n", kyarabenStateDir)
		}
	}

	homeDir, _ := os.UserHomeDir()
	refreshIconCaches(homeDir)

	fmt.Println()
	fmt.Println("Done. Kyaraben files have been removed.")
	fmt.Println()
	fmt.Printf("To fully uninstall, also remove:\n")
	fmt.Printf("  %s (your config)\n", configDir)

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

// forceRemoveAll removes a directory tree, even if it contains read-only files
// (like nix store paths). It works by making all files and directories writable.
func forceRemoveAll(path string) error {
	err := filepath.WalkDir(path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			_ = os.Chmod(p, 0755)
		} else {
			_ = os.Chmod(p, 0644)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func refreshIconCaches(homeDir string) {
	iconsDir := filepath.Join(homeDir, ".local", "share", "icons", "hicolor")

	runWithTimeout := func(name string, args ...string) {
		cmd := exec.Command(name, args...)
		if err := cmd.Start(); err != nil {
			return
		}
		done := make(chan error)
		go func() { done <- cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = cmd.Process.Kill()
		}
	}

	runWithTimeout("gtk-update-icon-cache", "-f", "-t", iconsDir)
	runWithTimeout("kbuildsycoca6")
	runWithTimeout("kbuildsycoca5")
}
