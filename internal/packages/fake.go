package packages

import (
	"context"
	"os"
	"path/filepath"
)

type FakeDownloader struct {
	Content []byte
	Error   error
	Calls   []DownloadRequest
}

func (f *FakeDownloader) Download(ctx context.Context, req DownloadRequest) error {
	f.Calls = append(f.Calls, req)
	if f.Error != nil {
		return f.Error
	}
	if err := os.MkdirAll(filepath.Dir(req.DestPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(req.DestPath, f.Content, 0644)
}

var _ Downloader = (*FakeDownloader)(nil)

type FakeExtractor struct {
	Files map[string]string // relative path -> content
	Error error
	Calls []fakeExtractCall
}

type fakeExtractCall struct {
	ArchivePath string
	DestDir     string
	ArchiveType string
}

func (f *FakeExtractor) Extract(archivePath, destDir, archiveType string) error {
	f.Calls = append(f.Calls, fakeExtractCall{archivePath, destDir, archiveType})
	if f.Error != nil {
		return f.Error
	}
	for relPath, content := range f.Files {
		fullPath := filepath.Join(destDir, relPath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(fullPath, []byte(content), 0755); err != nil {
			return err
		}
	}
	return nil
}

var _ Extractor = (*FakeExtractor)(nil)

type FakeInstaller struct {
	Binaries     map[string]*InstalledBinary
	Cores        []InstalledCore
	Icons        map[string]*InstalledIcon
	Installed    map[string]bool
	Versions     map[string]string
	GCCalls      []map[string]string
	InstallError error
	CoresError   error
	IconError    error
	packagesDir  string
}

func NewFakeInstaller(dir string) *FakeInstaller {
	return &FakeInstaller{
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
	if err := os.MkdirAll(filepath.Dir(binPath), 0755); err != nil {
		return nil, err
	}
	if err := os.WriteFile(binPath, []byte("#!/bin/sh\n"), 0755); err != nil {
		return nil, err
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
	return nil, nil
}

func (f *FakeInstaller) InstallIcon(ctx context.Context, binaryName, url, sha256 string) (*InstalledIcon, error) {
	if f.IconError != nil {
		return nil, f.IconError
	}
	if icon, ok := f.Icons[binaryName]; ok {
		return icon, nil
	}
	return &InstalledIcon{Name: binaryName, Filename: binaryName + ".png", Path: filepath.Join(f.packagesDir, "icons", binaryName+".png")}, nil
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
	if v, ok := f.Versions[name]; ok {
		return v
	}
	return "fake-version"
}

func (f *FakeInstaller) SetVersionOverrides(overrides map[string]string) {
	for k, v := range overrides {
		f.Versions[k] = v
	}
}

var _ Installer = (*FakeInstaller)(nil)
