package packages

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/hardware"
	"github.com/fnune/kyaraben/internal/logging"
	"github.com/fnune/kyaraben/internal/versions"
)

var log = logging.New("packages")

type InstallProgress struct {
	PackageName     string
	Phase           string // "downloading", "extracting", "installed", "skipped"
	BytesDownloaded int64
	BytesTotal      int64
}

type InstalledBinary struct {
	Name string
	Path string
}

type InstalledCore struct {
	Filename string
	Path     string
}

type InstalledIcon struct {
	Name     string
	Filename string
	Path     string
}

type InstallResult struct {
	Binaries []InstalledBinary
	Cores    []InstalledCore
	Icons    []InstalledIcon
}

type Installer interface {
	InstallEmulator(ctx context.Context, name string, onProgress func(InstallProgress)) (*InstalledBinary, error)
	InstallCores(ctx context.Context, coreNames []string, onProgress func(InstallProgress)) ([]InstalledCore, error)
	InstallIcon(ctx context.Context, binaryName, url, sha256 string) (*InstalledIcon, error)
	IsEmulatorInstalled(name string) bool
	GarbageCollect(keep map[string]string) error
	PackagesDir() string
	ResolveVersion(name string) string
	SetVersionOverrides(overrides map[string]string)
}

type PackageInstaller struct {
	fs           vfs.FS
	packagesDir  string
	downloadsDir string
	downloader   Downloader
	extractor    Extractor
	overrides    map[string]string
}

func NewPackageInstaller(fs vfs.FS, stateDir string, downloader Downloader, extractor Extractor) *PackageInstaller {
	return &PackageInstaller{
		fs:           fs,
		packagesDir:  filepath.Join(stateDir, "packages"),
		downloadsDir: filepath.Join(stateDir, "downloads"),
		downloader:   downloader,
		extractor:    extractor,
		overrides:    make(map[string]string),
	}
}

func NewDefaultPackageInstaller(stateDir string, downloader Downloader, extractor Extractor) *PackageInstaller {
	return NewPackageInstaller(vfs.OSFS, stateDir, downloader, extractor)
}

func (i *PackageInstaller) SetVersionOverrides(overrides map[string]string) {
	i.overrides = overrides
}

func (i *PackageInstaller) PackagesDir() string {
	return i.packagesDir
}

func (i *PackageInstaller) resolveSpec(name string) (*versions.VersionEntry, *versions.PackageSpec, error) {
	v := versions.MustGet()

	spec, ok := v.GetPackage(name)
	if !ok {
		return nil, nil, fmt.Errorf("unknown package: %s", name)
	}

	if override, ok := i.overrides[name]; ok {
		entry := spec.GetVersion(override)
		if entry == nil {
			return nil, nil, fmt.Errorf("version %s not found for %s", override, name)
		}
		return entry, spec, nil
	}

	entry := spec.GetDefault()
	if entry == nil {
		return nil, nil, fmt.Errorf("no default version for %s", name)
	}
	return entry, spec, nil
}

func (i *PackageInstaller) ResolveVersion(name string) string {
	if name == "retroarch-cores" {
		return versions.MustGet().RetroArchCores.Default
	}
	entry, _, err := i.resolveSpec(name)
	if err != nil {
		return ""
	}
	return entry.Version
}

func (i *PackageInstaller) packageDir(name, version string) string {
	return filepath.Join(i.packagesDir, name, version)
}

func (i *PackageInstaller) IsEmulatorInstalled(name string) bool {
	entry, spec, err := i.resolveSpec(name)
	if err != nil {
		return false
	}

	target := i.selectTarget(entry)
	if target == "" {
		return false
	}

	binaryPath := entry.BinaryPathForTarget(target, spec)
	if binaryPath == "" {
		binaryPath = name
	}

	pkgDir := i.packageDir(name, entry.Version)
	fullPath := filepath.Join(pkgDir, "bin", filepath.Base(binaryPath))
	_, err = i.fs.Stat(fullPath)
	return err == nil
}

func (i *PackageInstaller) InstallEmulator(ctx context.Context, name string, onProgress func(InstallProgress)) (*InstalledBinary, error) {
	entry, spec, err := i.resolveSpec(name)
	if err != nil {
		return nil, err
	}

	target := i.selectTarget(entry)
	if target == "" {
		return nil, fmt.Errorf("no build available for %s on current architecture", name)
	}

	build := entry.Target(target)
	archiveType := entry.ArchiveType(target, spec)
	url := entry.URL(target, spec)
	binaryPath := entry.BinaryPathForTarget(target, spec)

	pkgDir := i.packageDir(name, entry.Version)
	installedPath := filepath.Join(pkgDir, "bin", name)
	cache := &packageCache{fs: i.fs, pkgDir: pkgDir, expectedHash: build.SHA256}
	if cache.isValid(installedPath) {
		if onProgress != nil {
			onProgress(InstallProgress{PackageName: name, Phase: "skipped"})
		}
		return &InstalledBinary{Name: name, Path: installedPath}, nil
	}
	cache.invalidate()

	if err := vfs.MkdirAll(i.fs, i.downloadsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating downloads dir: %w", err)
	}

	log.Info("Downloading %s %s", name, entry.Version)

	downloadDest := filepath.Join(i.downloadsDir, name+"-"+entry.Version+".download")
	defer func() { _ = i.fs.Remove(downloadDest) }()

	if onProgress != nil {
		onProgress(InstallProgress{PackageName: name, Phase: "downloading"})
	}

	dlReq := DownloadRequest{
		URLs:     []string{url},
		SHA256:   build.SHA256,
		DestPath: downloadDest,
	}
	if onProgress != nil {
		dlReq.OnProgress = func(p DownloadProgress) {
			onProgress(InstallProgress{
				PackageName:     name,
				Phase:           "downloading",
				BytesDownloaded: p.BytesDownloaded,
				BytesTotal:      p.BytesTotal,
			})
		}
	}

	if err := i.downloader.Download(ctx, dlReq); err != nil {
		return nil, fmt.Errorf("downloading %s: %w", name, err)
	}

	if archiveType != "" && onProgress != nil {
		onProgress(InstallProgress{PackageName: name, Phase: "extracting"})
		log.Info("Extracting %s %s", name, entry.Version)
	}

	tmpDir := pkgDir + ".tmp"
	_ = i.fs.RemoveAll(tmpDir)

	if archiveType == "" {
		binDir := filepath.Join(tmpDir, "bin")
		if err := vfs.MkdirAll(i.fs, binDir, 0755); err != nil {
			return nil, fmt.Errorf("creating bin dir: %w", err)
		}
		destPath := filepath.Join(binDir, name)
		if err := i.copyFileExec(downloadDest, destPath); err != nil {
			return nil, fmt.Errorf("installing binary: %w", err)
		}
	} else {
		extractDir := filepath.Join(tmpDir, "extracted")
		if err := i.extractor.Extract(downloadDest, extractDir, archiveType); err != nil {
			_ = i.fs.RemoveAll(tmpDir)
			return nil, fmt.Errorf("extracting %s: %w", name, err)
		}

		binDir := filepath.Join(tmpDir, "bin")
		if err := vfs.MkdirAll(i.fs, binDir, 0755); err != nil {
			_ = i.fs.RemoveAll(tmpDir)
			return nil, fmt.Errorf("creating bin dir: %w", err)
		}

		srcPath := filepath.Join(extractDir, binaryPath)
		destPath := filepath.Join(binDir, name)
		if err := i.copyFileExec(srcPath, destPath); err != nil {
			_ = i.fs.RemoveAll(tmpDir)
			return nil, fmt.Errorf("installing extracted binary: %w", err)
		}

		_ = i.fs.RemoveAll(extractDir)
	}

	_ = i.fs.RemoveAll(pkgDir)
	if err := vfs.MkdirAll(i.fs, filepath.Dir(pkgDir), 0755); err != nil {
		_ = i.fs.RemoveAll(tmpDir)
		return nil, fmt.Errorf("creating package parent dir: %w", err)
	}
	if err := i.fs.Rename(tmpDir, pkgDir); err != nil {
		_ = i.fs.RemoveAll(tmpDir)
		return nil, fmt.Errorf("finalizing package install: %w", err)
	}

	cache.markValid()

	installedPath = filepath.Join(pkgDir, "bin", name)
	if onProgress != nil {
		onProgress(InstallProgress{PackageName: name, Phase: "installed"})
	}

	log.Info("Installed %s %s to %s", name, entry.Version, pkgDir)
	return &InstalledBinary{Name: name, Path: installedPath}, nil
}

func (i *PackageInstaller) InstallCores(ctx context.Context, coreNames []string, onProgress func(InstallProgress)) ([]InstalledCore, error) {
	v := versions.MustGet()
	targetName := i.selectCoresTarget(v)
	if targetName == "" {
		return nil, fmt.Errorf("no RetroArch cores bundle for this system")
	}
	url, sha256, ok := v.RetroArchCores.GetCoresURL(targetName)
	if !ok {
		return nil, fmt.Errorf("no RetroArch cores bundle for target %s", targetName)
	}

	version := v.RetroArchCores.Default
	pkgDir := i.packageDir("retroarch-cores", version)
	coresDir := filepath.Join(pkgDir, "lib", "retroarch", "cores")
	cache := &packageCache{fs: i.fs, pkgDir: pkgDir, expectedHash: sha256}

	allPresent := true
	for _, coreName := range coreNames {
		filename, ok := v.RetroArchCores.Files[coreName]
		if !ok {
			continue
		}
		if _, err := i.fs.Stat(filepath.Join(coresDir, filename)); err != nil {
			allPresent = false
			break
		}
	}
	if allPresent && cache.isValid(coresDir) {
		if onProgress != nil {
			onProgress(InstallProgress{PackageName: "retroarch-cores", Phase: "skipped"})
		}
		return i.buildCoresList(coreNames, coresDir, v), nil
	}
	cache.invalidate()

	if err := vfs.MkdirAll(i.fs, i.downloadsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating downloads dir: %w", err)
	}

	log.Info("Downloading RetroArch cores bundle %s", version)

	downloadDest := filepath.Join(i.downloadsDir, "retroarch-cores-"+version+".download")
	defer func() { _ = i.fs.Remove(downloadDest) }()

	if onProgress != nil {
		onProgress(InstallProgress{PackageName: "retroarch-cores", Phase: "downloading"})
	}

	dlReq := DownloadRequest{
		URLs:     []string{url},
		SHA256:   sha256,
		DestPath: downloadDest,
	}
	if onProgress != nil {
		dlReq.OnProgress = func(p DownloadProgress) {
			onProgress(InstallProgress{
				PackageName:     "retroarch-cores",
				Phase:           "downloading",
				BytesDownloaded: p.BytesDownloaded,
				BytesTotal:      p.BytesTotal,
			})
		}
	}

	if err := i.downloader.Download(ctx, dlReq); err != nil {
		return nil, fmt.Errorf("downloading cores bundle: %w", err)
	}

	if onProgress != nil {
		onProgress(InstallProgress{PackageName: "retroarch-cores", Phase: "extracting"})
	}

	log.Info("Extracting RetroArch cores bundle %s", version)

	extractDir := filepath.Join(i.downloadsDir, "retroarch-cores-extract")
	_ = i.fs.RemoveAll(extractDir)
	defer func() { _ = i.fs.RemoveAll(extractDir) }()

	archiveType := detectArchiveType(url)
	if err := i.extractor.Extract(downloadDest, extractDir, archiveType); err != nil {
		return nil, fmt.Errorf("extracting cores bundle: %w", err)
	}

	tmpDir := pkgDir + ".tmp"
	_ = i.fs.RemoveAll(tmpDir)
	tmpCoresDir := filepath.Join(tmpDir, "lib", "retroarch", "cores")
	if err := vfs.MkdirAll(i.fs, tmpCoresDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cores dir: %w", err)
	}

	for _, coreName := range coreNames {
		filename, ok := v.RetroArchCores.Files[coreName]
		if !ok {
			log.Info("Unknown core: %s, skipping", coreName)
			continue
		}

		src, err := i.findFile(extractDir, filename)
		if err != nil {
			_ = i.fs.RemoveAll(tmpDir)
			return nil, fmt.Errorf("finding core %s in archive: %w", coreName, err)
		}

		dest := filepath.Join(tmpCoresDir, filename)
		if err := i.copyFileMode(src, dest, 0644); err != nil {
			_ = i.fs.RemoveAll(tmpDir)
			return nil, fmt.Errorf("installing core %s: %w", coreName, err)
		}
	}

	_ = i.fs.RemoveAll(pkgDir)
	if err := vfs.MkdirAll(i.fs, filepath.Dir(pkgDir), 0755); err != nil {
		_ = i.fs.RemoveAll(tmpDir)
		return nil, fmt.Errorf("creating package parent dir: %w", err)
	}
	if err := i.fs.Rename(tmpDir, pkgDir); err != nil {
		_ = i.fs.RemoveAll(tmpDir)
		return nil, fmt.Errorf("finalizing cores install: %w", err)
	}

	cache.markValid()

	coresDir = filepath.Join(pkgDir, "lib", "retroarch", "cores")
	if onProgress != nil {
		onProgress(InstallProgress{PackageName: "retroarch-cores", Phase: "installed"})
	}

	log.Info("Installed %d RetroArch cores to %s", len(coreNames), pkgDir)
	return i.buildCoresList(coreNames, coresDir, v), nil
}

func (i *PackageInstaller) buildCoresList(coreNames []string, coresDir string, v *versions.Versions) []InstalledCore {
	var cores []InstalledCore
	for _, coreName := range coreNames {
		filename, ok := v.RetroArchCores.Files[coreName]
		if !ok {
			continue
		}
		cores = append(cores, InstalledCore{
			Filename: filename,
			Path:     filepath.Join(coresDir, filename),
		})
	}
	return cores
}

func (i *PackageInstaller) InstallIcon(ctx context.Context, binaryName, url, sha256Hash string) (*InstalledIcon, error) {
	ext := filepath.Ext(url)
	filename := binaryName + ext
	iconsDir := filepath.Join(i.packagesDir, "icons")
	destPath := filepath.Join(iconsDir, filename)

	if _, err := i.fs.Stat(destPath); err == nil {
		if sha256Hash != "" && i.verifyFileHash(destPath, sha256Hash) {
			return &InstalledIcon{Name: binaryName, Filename: filename, Path: destPath}, nil
		}
	}

	if err := vfs.MkdirAll(i.fs, iconsDir, 0755); err != nil {
		return nil, fmt.Errorf("creating icons dir: %w", err)
	}

	if err := i.downloader.Download(ctx, DownloadRequest{
		URLs:     []string{url},
		SHA256:   sha256Hash,
		DestPath: destPath,
	}); err != nil {
		return nil, fmt.Errorf("downloading icon for %s: %w", binaryName, err)
	}

	return &InstalledIcon{Name: binaryName, Filename: filename, Path: destPath}, nil
}

type packageCache struct {
	fs           vfs.FS
	pkgDir       string
	expectedHash string
}

func (c *packageCache) markerPath() string {
	return filepath.Join(c.pkgDir, ".sha256")
}

func (c *packageCache) isValid(installedPath string) bool {
	if _, err := c.fs.Stat(installedPath); err != nil {
		return false
	}
	data, err := c.fs.ReadFile(c.markerPath())
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(data)) == c.expectedHash
}

func (c *packageCache) invalidate() {
	_ = c.fs.RemoveAll(c.pkgDir)
}

func (c *packageCache) markValid() {
	_ = c.fs.WriteFile(c.markerPath(), []byte(c.expectedHash), 0644)
}

func (i *PackageInstaller) verifyFileHash(path, expectedHash string) bool {
	expected, err := parseSHA256(expectedHash)
	if err != nil {
		return false
	}

	f, err := i.fs.Open(path)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false
	}

	return bytes.Equal(h.Sum(nil), expected)
}

func (i *PackageInstaller) GarbageCollect(keep map[string]string) error {
	entries, err := i.fs.ReadDir(i.packagesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("reading packages dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		if name == "icons" {
			continue
		}

		keepVersion, wantKeep := keep[name]
		versionDirs, err := i.fs.ReadDir(filepath.Join(i.packagesDir, name))
		if err != nil {
			continue
		}

		for _, vd := range versionDirs {
			if !vd.IsDir() {
				continue
			}
			if wantKeep && vd.Name() == keepVersion {
				continue
			}
			toRemove := filepath.Join(i.packagesDir, name, vd.Name())
			log.Info("Removing unused package: %s", toRemove)
			_ = i.fs.RemoveAll(toRemove)
		}

		remaining, _ := i.fs.ReadDir(filepath.Join(i.packagesDir, name))
		if len(remaining) == 0 {
			_ = i.fs.Remove(filepath.Join(i.packagesDir, name))
		}
	}

	return nil
}

func (i *PackageInstaller) selectTarget(entry *versions.VersionEntry) string {
	return entry.SelectTarget(hardware.DetectTarget().Name)
}

func (i *PackageInstaller) selectCoresTarget(v *versions.Versions) string {
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

func detectArchiveType(url string) string {
	switch {
	case strings.HasSuffix(url, ".tar.zst"):
		return "tar.zst"
	case strings.HasSuffix(url, ".tar.gz"), strings.HasSuffix(url, ".tgz"):
		return "tar.gz"
	case strings.HasSuffix(url, ".tar.xz"):
		return "tar.xz"
	case strings.HasSuffix(url, ".zip"):
		return "zip"
	case strings.HasSuffix(url, ".7z"):
		return "7z"
	default:
		return ""
	}
}

func (i *PackageInstaller) findFile(root, filename string) (string, error) {
	var found string
	err := vfs.Walk(i.fs, root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.Name() == filename && !info.IsDir() {
			found = path
			return filepath.SkipAll
		}
		return nil
	})
	if err != nil && err != filepath.SkipAll {
		return "", err
	}
	if found == "" {
		return "", fmt.Errorf("file %s not found in archive", filename)
	}
	return found, nil
}

func (i *PackageInstaller) copyFileExec(src, dst string) error {
	return i.copyFileMode(src, dst, 0755)
}

func (i *PackageInstaller) copyFileMode(src, dst string, mode os.FileMode) error {
	data, err := i.fs.ReadFile(src)
	if err != nil {
		return err
	}
	return i.fs.WriteFile(dst, data, mode)
}

type ConcurrentInstaller struct {
	installer   Installer
	concurrency int
}

func NewConcurrentInstaller(installer Installer, concurrency int) *ConcurrentInstaller {
	return &ConcurrentInstaller{
		installer:   installer,
		concurrency: concurrency,
	}
}

func (c *ConcurrentInstaller) InstallAll(ctx context.Context, names []string, onProgress func(InstallProgress)) ([]InstalledBinary, error) {
	type result struct {
		binary *InstalledBinary
		err    error
	}

	results := make([]result, len(names))
	sem := make(chan struct{}, c.concurrency)
	var wg sync.WaitGroup

	var mu sync.Mutex
	safeProgress := func(p InstallProgress) {
		if onProgress != nil {
			mu.Lock()
			onProgress(p)
			mu.Unlock()
		}
	}

	for idx, name := range names {
		wg.Add(1)
		go func(idx int, name string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			binary, err := c.installer.InstallEmulator(ctx, name, safeProgress)
			results[idx] = result{binary: binary, err: err}
		}(idx, name)
	}

	wg.Wait()

	var binaries []InstalledBinary
	for idx, r := range results {
		if r.err != nil {
			return nil, fmt.Errorf("installing %s: %w", names[idx], r.err)
		}
		if r.binary != nil {
			binaries = append(binaries, *r.binary)
		}
	}

	return binaries, nil
}

var _ Installer = (*PackageInstaller)(nil)
