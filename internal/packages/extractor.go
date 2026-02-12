package packages

import (
	"archive/tar"
	"archive/zip"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/bodgit/sevenzip"
	"github.com/klauspost/compress/zstd"
	"github.com/ulikunitz/xz"
)

type Extractor interface {
	Extract(archivePath, destDir, archiveType string) error
}

type OSExtractor struct{}

func (e OSExtractor) Extract(archivePath, destDir, archiveType string) error {
	switch archiveType {
	case "tar.gz", "tgz":
		return extractTar(archivePath, destDir, openGzip)
	case "tar.xz":
		return extractTar(archivePath, destDir, openXZ)
	case "tar.zst":
		return extractTar(archivePath, destDir, openZstd)
	case "zip":
		return extractZip(archivePath, destDir)
	case "appimage":
		return installAppImage(archivePath, destDir)
	case "7z":
		return extract7z(archivePath, destDir)
	default:
		return fmt.Errorf("unsupported archive type: %s", archiveType)
	}
}

type decompressor func(io.Reader) (io.Reader, error)

func openGzip(r io.Reader) (io.Reader, error) {
	return gzip.NewReader(r)
}

func openXZ(r io.Reader) (io.Reader, error) {
	return xz.NewReader(r)
}

func openZstd(r io.Reader) (io.Reader, error) {
	return zstd.NewReader(r)
}

func extractTar(archivePath, destDir string, decompress decompressor) error {
	f, err := os.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening archive: %w", err)
	}
	defer func() { _ = f.Close() }()

	reader, err := decompress(f)
	if err != nil {
		return fmt.Errorf("decompressing: %w", err)
	}
	if closer, ok := reader.(io.Closer); ok {
		defer func() { _ = closer.Close() }()
	}

	tr := tar.NewReader(reader)
	for {
		header, err := tr.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("reading tar entry: %w", err)
		}

		target, err := sanitizePath(destDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := writeFile(target, tr, header.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := createSafeSymlink(destDir, target, header.Linkname); err != nil {
				return err
			}
		}
	}
}

func extractZip(archivePath, destDir string) error {
	r, err := zip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		target, err := sanitizePath(destDir, f.Name)
		if err != nil {
			return err
		}

		mode := f.FileInfo().Mode()

		if mode&os.ModeSymlink != 0 {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("opening zip symlink %s: %w", f.Name, err)
			}
			linkTarget, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				return fmt.Errorf("reading zip symlink %s: %w", f.Name, err)
			}
			if err := createSafeSymlink(destDir, target, string(linkTarget)); err != nil {
				return err
			}
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", target, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening zip entry %s: %w", f.Name, err)
		}

		err = writeFile(target, rc, mode)
		_ = rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func installAppImage(srcPath, destDir string) error {
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}

	src, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("opening appimage: %w", err)
	}
	defer func() { _ = src.Close() }()

	destName := filepath.Base(srcPath)
	destPath := filepath.Join(destDir, destName)

	dst, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return fmt.Errorf("copying appimage: %w", err)
	}

	return dst.Close()
}

func createSafeSymlink(destDir, target, linkTarget string) error {
	if filepath.IsAbs(linkTarget) {
		return fmt.Errorf("absolute symlink target: %s -> %s", target, linkTarget)
	}
	resolvedTarget := filepath.Join(filepath.Dir(target), linkTarget)
	resolvedTarget = filepath.Clean(resolvedTarget)
	if !strings.HasPrefix(resolvedTarget, filepath.Clean(destDir)+string(filepath.Separator)) && resolvedTarget != filepath.Clean(destDir) {
		return fmt.Errorf("symlink escapes destination: %s -> %s", target, linkTarget)
	}
	if err := os.Symlink(linkTarget, target); err != nil {
		return fmt.Errorf("creating symlink %s: %w", target, err)
	}
	return nil
}

func sanitizePath(destDir, name string) (string, error) {
	cleaned := filepath.Clean(name)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path traversal in archive: %s", name)
	}

	target := filepath.Join(destDir, cleaned)
	if !strings.HasPrefix(target, filepath.Clean(destDir)+string(filepath.Separator)) && target != filepath.Clean(destDir) {
		return "", fmt.Errorf("path escapes destination: %s", name)
	}

	if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
		return "", fmt.Errorf("creating parent directory: %w", err)
	}

	return target, nil
}

func writeFile(path string, r io.Reader, mode os.FileMode) error {
	if mode == 0 {
		mode = 0644
	}
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}

	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return fmt.Errorf("writing file %s: %w", path, err)
	}

	return f.Close()
}

func extract7z(archivePath, destDir string) error {
	r, err := sevenzip.OpenReader(archivePath)
	if err != nil {
		return fmt.Errorf("opening 7z: %w", err)
	}
	defer func() { _ = r.Close() }()

	for _, f := range r.File {
		target, err := sanitizePath(destDir, f.Name)
		if err != nil {
			return err
		}

		mode := f.FileInfo().Mode()

		if mode&os.ModeSymlink != 0 {
			rc, err := f.Open()
			if err != nil {
				return fmt.Errorf("opening 7z symlink %s: %w", f.Name, err)
			}
			linkTarget, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				return fmt.Errorf("reading 7z symlink %s: %w", f.Name, err)
			}
			if err := createSafeSymlink(destDir, target, string(linkTarget)); err != nil {
				return err
			}
			continue
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", target, err)
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("opening 7z entry %s: %w", f.Name, err)
		}

		err = writeFile(target, rc, mode)
		_ = rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

var _ Extractor = OSExtractor{}
