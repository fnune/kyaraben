package packages

import (
	"os"
	"path/filepath"

	"github.com/fnune/kyaraben/internal/versions"
)

type ChangeSummary struct {
	TotalDownloadBytes int64
	TotalFreeBytes     int64
	PackagesToDownload []string
}

func CalculateChangeSummary(toInstall []string, toRemove []string, installed map[string]bool, getSizeFunc func(pkgName string) int64) ChangeSummary {
	summary := ChangeSummary{}
	seenInstall := make(map[string]bool)
	for _, pkgName := range toInstall {
		if seenInstall[pkgName] {
			continue
		}
		seenInstall[pkgName] = true
		if installed[pkgName] {
			continue
		}
		size := getSizeFunc(pkgName)
		if size > 0 {
			summary.TotalDownloadBytes += size
		}
		summary.PackagesToDownload = append(summary.PackagesToDownload, pkgName)
	}

	seenRemove := make(map[string]bool)
	for _, pkgName := range toRemove {
		if seenRemove[pkgName] {
			continue
		}
		seenRemove[pkgName] = true
		if !installed[pkgName] {
			continue
		}
		size := getSizeFunc(pkgName)
		if size > 0 {
			summary.TotalFreeBytes += size
		}
	}

	return summary
}

func RetroArchCoresInstalled(installer Installer, coreNames []string, v *versions.Versions) bool {
	if len(coreNames) == 0 {
		return true
	}

	packagesDir := installer.PackagesDir()

	for _, coreName := range coreNames {
		version := installer.ResolveVersion(coreName)
		if version == "" {
			return false
		}

		pkgDir := filepath.Join(packagesDir, "retroarch-cores", version)
		coresDir := filepath.Join(pkgDir, "lib", "retroarch", "cores")

		var filename string
		if standalone, ok := v.RetroArchCores.Standalone[coreName]; ok {
			filename = standalone.Filename
		} else if f, ok := v.RetroArchCores.Files[coreName]; ok {
			filename = f
		} else {
			continue
		}
		if _, err := os.Stat(filepath.Join(coresDir, filename)); err != nil {
			return false
		}
	}

	return true
}
