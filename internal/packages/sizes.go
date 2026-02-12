package packages

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/hardware"
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

func RetroArchCoresInstalled(packagesDir string, coreNames []string, v *versions.Versions) bool {
	if len(coreNames) == 0 {
		return true
	}

	targetName := selectCoresTarget(v)
	if targetName == "" {
		return false
	}

	version := v.RetroArchCores.Default
	if version == "" {
		return false
	}
	build, ok := v.RetroArchCores.Versions[version]
	if !ok {
		return false
	}
	targetBuild, ok := build.Targets[targetName]
	if !ok {
		return false
	}

	pkgDir := filepath.Join(packagesDir, "retroarch-cores", version)
	coresDir := filepath.Join(pkgDir, "lib", "retroarch", "cores")

	for _, coreName := range coreNames {
		filename, ok := v.RetroArchCores.Files[coreName]
		if !ok {
			continue
		}
		if _, err := os.Stat(filepath.Join(coresDir, filename)); err != nil {
			return false
		}
	}

	markerPath := filepath.Join(pkgDir, ".sha256")
	data, err := os.ReadFile(markerPath)
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == targetBuild.SHA256
}

func selectCoresTarget(v *versions.Versions) string {
	detected := hardware.DetectTarget().Name
	version := v.RetroArchCores.Default
	if version == "" {
		return ""
	}
	build, ok := v.RetroArchCores.Versions[version]
	if !ok {
		return ""
	}

	if _, ok := build.Targets[detected]; ok {
		return detected
	}

	if fallback, ok := versions.TargetFallback[detected]; ok {
		if _, ok := build.Targets[fallback.String()]; ok {
			return fallback.String()
		}
	}

	return ""
}
