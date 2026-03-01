package packages

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/versions"
)

type FakeDownloader struct {
	fs      vfs.FS
	Content []byte
	Error   error
	Calls   []DownloadRequest
	mu      sync.Mutex
}

func NewFakeDownloader(fs vfs.FS, content []byte) *FakeDownloader {
	return &FakeDownloader{fs: fs, Content: content}
}

func (f *FakeDownloader) Download(ctx context.Context, req DownloadRequest) error {
	f.mu.Lock()
	f.Calls = append(f.Calls, req)
	f.mu.Unlock()
	if f.Error != nil {
		return f.Error
	}
	if err := vfs.MkdirAll(f.fs, filepath.Dir(req.DestPath), 0755); err != nil {
		return err
	}
	return f.fs.WriteFile(req.DestPath, f.Content, 0644)
}

var _ Downloader = (*FakeDownloader)(nil)

type FakeExtractor struct {
	fs    vfs.FS
	Files map[string]string // relative path -> content
	Error error
	Calls []fakeExtractCall
}

type fakeExtractCall struct {
	ArchivePath string
	DestDir     string
	ArchiveType string
}

func NewFakeExtractor(fs vfs.FS, files map[string]string) *FakeExtractor {
	return &FakeExtractor{fs: fs, Files: files}
}

func (f *FakeExtractor) Extract(archivePath, destDir, archiveType string) error {
	f.Calls = append(f.Calls, fakeExtractCall{archivePath, destDir, archiveType})
	if f.Error != nil {
		return f.Error
	}
	for relPath, content := range f.Files {
		fullPath := filepath.Join(destDir, relPath)
		if err := vfs.MkdirAll(f.fs, filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := f.fs.WriteFile(fullPath, []byte(content), 0755); err != nil {
			return err
		}
	}
	return nil
}

var _ Extractor = (*FakeExtractor)(nil)

type FakeInstaller struct {
	fs           vfs.FS
	Binaries     map[string]*InstalledBinary
	Cores        []InstalledCore
	Icons        map[string]*InstalledIcon
	Installed    map[string]bool
	Versions     map[string]string
	overrides    map[string]string
	GCCalls      []map[string]string
	InstallError error
	CoresError   error
	IconError    error
	packagesDir  string
}

func NewFakeInstaller(fs vfs.FS, dir string) *FakeInstaller {
	return &FakeInstaller{
		fs:          fs,
		Binaries:    make(map[string]*InstalledBinary),
		Icons:       make(map[string]*InstalledIcon),
		Installed:   make(map[string]bool),
		Versions:    make(map[string]string),
		packagesDir: dir,
	}
}

func (f *FakeInstaller) InstallEmulator(ctx context.Context, name string, onProgress func(InstallProgress)) (*InstalledBinary, error) {
	if f.InstallError != nil {
		return nil, f.InstallError
	}
	if b, ok := f.Binaries[name]; ok {
		return b, nil
	}
	binPath := filepath.Join(f.packagesDir, name, "bin", name)
	if err := vfs.MkdirAll(f.fs, filepath.Dir(binPath), 0755); err != nil {
		return nil, err
	}
	if err := f.fs.WriteFile(binPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		return nil, err
	}
	if f.Installed != nil {
		f.Installed[name] = true
	}
	return &InstalledBinary{Name: name, Path: binPath}, nil
}

func (f *FakeInstaller) InstallCores(ctx context.Context, coreNames []string, onProgress func(InstallProgress)) ([]InstalledCore, error) {
	if f.CoresError != nil {
		return nil, f.CoresError
	}
	if f.Cores != nil {
		return f.Cores, nil
	}
	v := versions.MustGet()
	installed := make([]InstalledCore, 0, len(coreNames))
	for _, name := range coreNames {
		spec, ok := v.GetPackage(name)
		if !ok {
			continue
		}
		filename := spec.BinaryPath
		if filename == "" {
			filename = name + "_libretro.so"
		}
		version := f.ResolveVersion(name)
		coresDir := filepath.Join(f.packagesDir, name, version, "lib", "retroarch", "cores")
		if err := vfs.MkdirAll(f.fs, coresDir, 0755); err != nil {
			return nil, err
		}
		corePath := filepath.Join(coresDir, filename)
		if err := f.fs.WriteFile(corePath, []byte("fake core"), 0644); err != nil {
			return nil, err
		}
		installed = append(installed, InstalledCore{Filename: filename, Path: corePath})
	}
	return installed, nil
}

func (f *FakeInstaller) InstallIcon(ctx context.Context, binaryName, url, sha256 string) (*InstalledIcon, error) {
	if f.IconError != nil {
		return nil, f.IconError
	}
	if icon, ok := f.Icons[binaryName]; ok {
		return icon, nil
	}
	iconPath := filepath.Join(f.packagesDir, "icons", binaryName+".png")
	if err := vfs.MkdirAll(f.fs, filepath.Dir(iconPath), 0755); err != nil {
		return nil, err
	}
	if err := f.fs.WriteFile(iconPath, []byte("fake icon"), 0644); err != nil {
		return nil, err
	}
	return &InstalledIcon{Name: binaryName, Filename: binaryName + ".png", Path: iconPath}, nil
}

func (f *FakeInstaller) IsEmulatorInstalled(name string) bool {
	return f.Installed[name]
}

func (f *FakeInstaller) GarbageCollect(keep map[string]string) error {
	f.GCCalls = append(f.GCCalls, keep)
	return nil
}

func (f *FakeInstaller) PackagesDir() string {
	return f.packagesDir
}

func (f *FakeInstaller) ResolveVersion(name string) string {
	if f.overrides != nil {
		if v, ok := f.overrides[name]; ok {
			return v
		}
	}
	if v, ok := f.Versions[name]; ok {
		return v
	}
	v := versions.MustGet()
	if spec, ok := v.GetPackage(name); ok {
		if entry := spec.GetDefault(); entry != nil {
			return entry.Version
		}
	}
	return ""
}

func (f *FakeInstaller) SetVersionOverrides(overrides map[string]string) {
	f.overrides = overrides
}

var _ Installer = (*FakeInstaller)(nil)
