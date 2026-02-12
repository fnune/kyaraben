package packages

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"os"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
)

func createTarGz(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := gw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func createTarZst(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "archive.tar.zst")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	zw, err := zstd.NewWriter(f)
	if err != nil {
		t.Fatal(err)
	}
	tw := tar.NewWriter(zw)

	for name, content := range files {
		if err := tw.WriteHeader(&tar.Header{
			Name: name,
			Mode: 0755,
			Size: int64(len(content)),
		}); err != nil {
			t.Fatal(err)
		}
		if _, err := tw.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}

	if err := tw.Close(); err != nil {
		t.Fatal(err)
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func createZip(t *testing.T, files map[string]string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "archive.zip")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = f.Close() }()

	zw := zip.NewWriter(f)
	for name, content := range files {
		w, err := zw.Create(name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := w.Write([]byte(content)); err != nil {
			t.Fatal(err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestExtractTarGz(t *testing.T) {
	archive := createTarGz(t, map[string]string{
		"bin/program":    "#!/bin/sh\necho hello",
		"lib/library.so": "fake-library",
	})

	dest := filepath.Join(t.TempDir(), "extracted")
	ext := OSExtractor{}

	if err := ext.Extract(archive, dest, "tar.gz"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dest, "bin", "program"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "#!/bin/sh\necho hello" {
		t.Errorf("content = %q", content)
	}

	info, err := os.Stat(filepath.Join(dest, "bin", "program"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("extracted file should be executable")
	}
}

func TestExtractTarZst(t *testing.T) {
	archive := createTarZst(t, map[string]string{
		"cores/bsnes_libretro.so": "fake-core",
	})

	dest := filepath.Join(t.TempDir(), "extracted")
	ext := OSExtractor{}

	if err := ext.Extract(archive, dest, "tar.zst"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dest, "cores", "bsnes_libretro.so"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "fake-core" {
		t.Errorf("content = %q", content)
	}
}

func TestExtractZip(t *testing.T) {
	archive := createZip(t, map[string]string{
		"melonDS-x86_64.AppImage": "fake-appimage",
	})

	dest := filepath.Join(t.TempDir(), "extracted")
	ext := OSExtractor{}

	if err := ext.Extract(archive, dest, "zip"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dest, "melonDS-x86_64.AppImage"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "fake-appimage" {
		t.Errorf("content = %q", content)
	}
}

func TestExtractAppImage(t *testing.T) {
	srcDir := t.TempDir()
	srcPath := filepath.Join(srcDir, "eden.AppImage")
	if err := os.WriteFile(srcPath, []byte("fake-appimage-binary"), 0644); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "extracted")
	ext := OSExtractor{}

	if err := ext.Extract(srcPath, dest, "appimage"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	destPath := filepath.Join(dest, "eden.AppImage")
	info, err := os.Stat(destPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("appimage should be executable")
	}
}

func TestExtractPathTraversal(t *testing.T) {
	path := filepath.Join(t.TempDir(), "malicious.tar.gz")
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}

	gw := gzip.NewWriter(f)
	tw := tar.NewWriter(gw)

	content := []byte("pwned")
	_ = tw.WriteHeader(&tar.Header{
		Name: "../../../etc/passwd",
		Mode: 0644,
		Size: int64(len(content)),
	})
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gw.Close()
	_ = f.Close()

	dest := filepath.Join(t.TempDir(), "extracted")
	ext := OSExtractor{}

	err = ext.Extract(path, dest, "tar.gz")
	if err == nil {
		t.Fatal("expected path traversal error")
	}
}

func TestExtractAbsoluteSymlink(t *testing.T) {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	_ = tw.WriteHeader(&tar.Header{
		Name:     "link",
		Typeflag: tar.TypeSymlink,
		Linkname: "/etc/passwd",
	})
	_ = tw.Close()
	_ = gw.Close()

	path := filepath.Join(t.TempDir(), "archive.tar.gz")
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	dest := filepath.Join(t.TempDir(), "extracted")
	ext := OSExtractor{}

	err := ext.Extract(path, dest, "tar.gz")
	if err == nil {
		t.Fatal("expected error for absolute symlink")
	}
}

func TestExtractUnsupportedType(t *testing.T) {
	ext := OSExtractor{}
	err := ext.Extract("/dev/null", t.TempDir(), "rar")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
