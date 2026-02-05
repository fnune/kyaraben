package cli

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/fnune/kyaraben/internal/launcher"
	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
)

type UninstallCmd struct {
	Force   bool `short:"f" help:"Skip confirmation prompt."`
	DryRun  bool `short:"n" help:"Show what would be removed without doing anything."`
	WaitPID int  `help:"Wait for this process to exit before uninstalling."`
	Notify  bool `help:"Show a desktop notification when done."`
}

func (cmd *UninstallCmd) Run(ctx *Context) error {
	if cmd.WaitPID > 0 {
		waitForProcess(cmd.WaitPID)
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
	manifest, err := model.LoadManifest(manifestPath)
	if err != nil {
		fmt.Printf("Warning: could not load manifest: %v\n", err)
		fmt.Println("Some files may not be listed for removal.")
		fmt.Println()
		fmt.Printf("Manifest path: %s\n", manifestPath)
		if data, readErr := os.ReadFile(manifestPath); readErr == nil {
			fmt.Println("Manifest contents:")
			fmt.Println(string(data))
		}
		fmt.Println()
		manifest = model.NewManifest()
	}

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
	iconsDir := filepath.Join(homeDir, ".local", "share", "icons", "hicolor")
	launcher.UpdateIconCaches(iconsDir)

	fmt.Println()
	fmt.Println("Done. Kyaraben files have been removed.")
	fmt.Println()
	fmt.Printf("To fully uninstall, also remove:\n")
	fmt.Printf("  %s (your config)\n", configDir)

	if cmd.Notify {
		pathEnv := os.Getenv("PATH")
		if err := sendNotification("Kyaraben uninstalled", "All managed files have been removed."); err != nil {
			_ = os.WriteFile("/tmp/kyaraben-uninstall-debug.txt", []byte(fmt.Sprintf("notification failed: %v\nPATH=%s\n", err, pathEnv)), 0644)
		} else {
			_ = os.WriteFile("/tmp/kyaraben-uninstall-debug.txt", []byte("notification sent successfully\n"), 0644)
		}
	}

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
// (like nix store paths). It works by recursively fixing permissions before
// descending into directories, which is necessary because WalkDir can't enter
// directories without execute permission.
func forceRemoveAll(path string) error {
	if err := forceChmodRecursive(path); err != nil {
		return err
	}
	return os.RemoveAll(path)
}

func forceChmodRecursive(path string) error {
	info, err := os.Lstat(path)
	if err != nil {
		return err
	}

	if info.Mode()&os.ModeSymlink != 0 {
		return nil
	}

	if !info.IsDir() {
		return os.Chmod(path, 0644)
	}

	if err := os.Chmod(path, 0755); err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := forceChmodRecursive(filepath.Join(path, entry.Name())); err != nil {
			return err
		}
	}

	return nil
}

func waitForProcess(pid int) {
	for {
		if err := syscall.Kill(pid, 0); err != nil {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}
}

func sendNotification(title, body string) error {
	return exec.Command("/usr/bin/env", "notify-send", title, body).Run()
}
