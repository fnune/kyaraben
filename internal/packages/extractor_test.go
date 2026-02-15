package packages

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"path/filepath"
	"testing"

	"github.com/klauspost/compress/zstd"
	"github.com/twpayne/go-vfs/v5"
	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/testutil"
)

func createTarGzVFS(t *testing.T, fs vfs.FS, files map[string]string) string {
	t.Helper()
	path := "/archive.tar.gz"
	f, err := fs.Create(path)
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

func createTarZstVFS(t *testing.T, fs vfs.FS, files map[string]string) string {
	t.Helper()
	path := "/archive.tar.zst"
	f, err := fs.Create(path)
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

func createZipVFS(t *testing.T, fs vfs.FS, files map[string]string) string {
	t.Helper()
	path := "/archive.zip"
	f, err := fs.Create(path)
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
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{})

	archive := createTarGzVFS(t, fs, map[string]string{
		"bin/program":    "#!/bin/sh\necho hello",
		"lib/library.so": "fake-library",
	})

	dest := "/extracted"
	ext := NewExtractor(fs)

	if err := ext.Extract(archive, dest, "tar.gz"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	content, err := fs.ReadFile(filepath.Join(dest, "bin", "program"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "#!/bin/sh\necho hello" {
		t.Errorf("content = %q", content)
	}

	info, err := fs.Stat(filepath.Join(dest, "bin", "program"))
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("extracted file should be executable")
	}
}

func TestExtractTarZst(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{})

	archive := createTarZstVFS(t, fs, map[string]string{
		"cores/bsnes_libretro.so": "fake-core",
	})

	dest := "/extracted"
	ext := NewExtractor(fs)

	if err := ext.Extract(archive, dest, "tar.zst"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	content, err := fs.ReadFile(filepath.Join(dest, "cores", "bsnes_libretro.so"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "fake-core" {
		t.Errorf("content = %q", content)
	}
}

func TestExtractZip(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{})

	archive := createZipVFS(t, fs, map[string]string{
		"melonDS-x86_64.AppImage": "fake-appimage",
	})

	dest := "/extracted"
	ext := NewExtractor(fs)

	if err := ext.Extract(archive, dest, "zip"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	content, err := fs.ReadFile(filepath.Join(dest, "melonDS-x86_64.AppImage"))
	if err != nil {
		t.Fatalf("reading extracted file: %v", err)
	}
	if string(content) != "fake-appimage" {
		t.Errorf("content = %q", content)
	}
}

func TestExtractAppImage(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/src/eden.AppImage": "fake-appimage-binary",
	})

	dest := "/extracted"
	ext := NewExtractor(fs)

	if err := ext.Extract("/src/eden.AppImage", dest, "appimage"); err != nil {
		t.Fatalf("Extract: %v", err)
	}

	destPath := filepath.Join(dest, "eden.AppImage")
	info, err := fs.Stat(destPath)
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Error("appimage should be executable")
	}
}

func TestExtractPathTraversal(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{})

	path := "/malicious.tar.gz"
	f, err := fs.Create(path)
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

	dest := "/extracted"
	ext := NewExtractor(fs)

	err = ext.Extract(path, dest, "tar.gz")
	if err == nil {
		t.Fatal("expected path traversal error")
	}
}

func TestExtractAbsoluteSymlink(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{})

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

	path := "/archive.tar.gz"
	if err := fs.WriteFile(path, buf.Bytes(), 0644); err != nil {
		t.Fatal(err)
	}

	dest := "/extracted"
	ext := NewExtractor(fs)

	err := ext.Extract(path, dest, "tar.gz")
	if err == nil {
		t.Fatal("expected error for absolute symlink")
	}
}

func TestExtractUnsupportedType(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/dest": &vfst.Dir{Perm: 0755},
	})

	ext := NewExtractor(fs)
	err := ext.Extract("/nonexistent", "/dest", "rar")
	if err == nil {
		t.Fatal("expected error for unsupported type")
	}
}
