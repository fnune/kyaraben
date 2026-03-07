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

	"github.com/gen2brain/go-unarr"
	"github.com/klauspost/compress/zstd"
	"github.com/twpayne/go-vfs/v5"
	"github.com/ulikunitz/xz"
)

type Extractor interface {
	Extract(archivePath, destDir, archiveType string) error
}

type OSExtractor struct {
	fs vfs.FS
}

func NewExtractor(fs vfs.FS) *OSExtractor {
	return &OSExtractor{fs: fs}
}

func NewDefaultExtractor() *OSExtractor {
	return NewExtractor(vfs.OSFS)
}

func (e *OSExtractor) Extract(archivePath, destDir, archiveType string) error {
	switch archiveType {
	case "tar.gz", "tgz":
		return e.extractTar(archivePath, destDir, openGzip)
	case "tar.xz":
		return e.extractTar(archivePath, destDir, openXZ)
	case "tar.zst":
		return e.extractTar(archivePath, destDir, openZstd)
	case "zip":
		return e.extractZip(archivePath, destDir)
	case "appimage":
		return e.installAppImage(archivePath, destDir)
	case "7z":
		return e.extract7z(archivePath, destDir)
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

func (e *OSExtractor) extractTar(archivePath, destDir string, decompress decompressor) error {
	f, err := e.fs.Open(archivePath)
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

		target, err := e.sanitizePath(destDir, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := vfs.MkdirAll(e.fs, target, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := e.writeFile(target, tr, header.FileInfo().Mode()); err != nil {
				return err
			}
		case tar.TypeSymlink:
			if err := e.createSafeSymlink(destDir, target, header.Linkname); err != nil {
				return err
			}
		}
	}
}

func (e *OSExtractor) extractZip(archivePath, destDir string) error {
	f, err := e.fs.Open(archivePath)
	if err != nil {
		return fmt.Errorf("opening zip: %w", err)
	}
	defer func() { _ = f.Close() }()

	readerAt, ok := f.(io.ReaderAt)
	if !ok {
		return fmt.Errorf("zip extraction requires ReaderAt support")
	}

	stat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("stat zip: %w", err)
	}

	r, err := zip.NewReader(readerAt, stat.Size())
	if err != nil {
		return fmt.Errorf("reading zip: %w", err)
	}

	for _, zf := range r.File {
		target, err := e.sanitizePath(destDir, zf.Name)
		if err != nil {
			return err
		}

		mode := zf.FileInfo().Mode()

		if mode&os.ModeSymlink != 0 {
			rc, err := zf.Open()
			if err != nil {
				return fmt.Errorf("opening zip symlink %s: %w", zf.Name, err)
			}
			linkTarget, err := io.ReadAll(rc)
			_ = rc.Close()
			if err != nil {
				return fmt.Errorf("reading zip symlink %s: %w", zf.Name, err)
			}
			if err := e.createSafeSymlink(destDir, target, string(linkTarget)); err != nil {
				return err
			}
			continue
		}

		if zf.FileInfo().IsDir() {
			if err := vfs.MkdirAll(e.fs, target, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}
			continue
		}

		if err := vfs.MkdirAll(e.fs, filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", target, err)
		}

		rc, err := zf.Open()
		if err != nil {
			return fmt.Errorf("opening zip entry %s: %w", zf.Name, err)
		}

		err = e.writeFile(target, rc, mode)
		_ = rc.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

func (e *OSExtractor) installAppImage(srcPath, destDir string) error {
	if err := vfs.MkdirAll(e.fs, destDir, 0755); err != nil {
		return fmt.Errorf("creating destination: %w", err)
	}

	src, err := e.fs.Open(srcPath)
	if err != nil {
		return fmt.Errorf("opening appimage: %w", err)
	}
	defer func() { _ = src.Close() }()

	destName := filepath.Base(srcPath)
	destPath := filepath.Join(destDir, destName)

	dst, err := e.fs.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("creating destination file: %w", err)
	}

	if _, err := io.Copy(dst, src); err != nil {
		_ = dst.Close()
		return fmt.Errorf("copying appimage: %w", err)
	}

	return dst.Close()
}

func (e *OSExtractor) createSafeSymlink(destDir, target, linkTarget string) error {
	if filepath.IsAbs(linkTarget) {
		return fmt.Errorf("absolute symlink target: %s -> %s", target, linkTarget)
	}
	resolvedTarget := filepath.Join(filepath.Dir(target), linkTarget)
	resolvedTarget = filepath.Clean(resolvedTarget)
	if !strings.HasPrefix(resolvedTarget, filepath.Clean(destDir)+string(filepath.Separator)) && resolvedTarget != filepath.Clean(destDir) {
		return fmt.Errorf("symlink escapes destination: %s -> %s", target, linkTarget)
	}
	if err := e.fs.Symlink(linkTarget, target); err != nil {
		return fmt.Errorf("creating symlink %s: %w", target, err)
	}
	return nil
}

func (e *OSExtractor) sanitizePath(destDir, name string) (string, error) {
	cleaned := filepath.Clean(name)
	if strings.HasPrefix(cleaned, "..") || filepath.IsAbs(cleaned) {
		return "", fmt.Errorf("path traversal in archive: %s", name)
	}

	target := filepath.Join(destDir, cleaned)
	if !strings.HasPrefix(target, filepath.Clean(destDir)+string(filepath.Separator)) && target != filepath.Clean(destDir) {
		return "", fmt.Errorf("path escapes destination: %s", name)
	}

	if err := vfs.MkdirAll(e.fs, filepath.Dir(target), 0755); err != nil {
		return "", fmt.Errorf("creating parent directory: %w", err)
	}

	return target, nil
}

func (e *OSExtractor) writeFile(path string, r io.Reader, mode os.FileMode) error {
	if mode == 0 {
		mode = 0644
	}
	f, err := e.fs.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return fmt.Errorf("creating file %s: %w", path, err)
	}

	if _, err := io.Copy(f, r); err != nil {
		_ = f.Close()
		return fmt.Errorf("writing file %s: %w", path, err)
	}

	return f.Close()
}

func (e *OSExtractor) extract7z(archivePath, destDir string) error {
	a, err := unarr.NewArchive(archivePath)
	if err != nil {
		return fmt.Errorf("opening 7z: %w", err)
	}
	defer func() { _ = a.Close() }()

	for {
		err := a.Entry()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("reading 7z entry: %w", err)
		}

		name := a.Name()
		target, err := e.sanitizePath(destDir, name)
		if err != nil {
			return err
		}

		if strings.HasSuffix(name, "/") {
			if err := vfs.MkdirAll(e.fs, target, 0755); err != nil {
				return fmt.Errorf("creating directory %s: %w", target, err)
			}
			continue
		}

		data, err := a.ReadAll()
		if err != nil {
			return fmt.Errorf("reading 7z entry %s: %w", name, err)
		}

		if err := vfs.MkdirAll(e.fs, filepath.Dir(target), 0755); err != nil {
			return fmt.Errorf("creating parent dir for %s: %w", target, err)
		}

		if err := e.fs.WriteFile(target, data, 0644); err != nil {
			return fmt.Errorf("writing file %s: %w", target, err)
		}
	}

	return nil
}

var _ Extractor = (*OSExtractor)(nil)
