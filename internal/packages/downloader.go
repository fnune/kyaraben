package packages

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/version"
)

type DownloadProgress struct {
	BytesDownloaded int64
	BytesTotal      int64
}

type DownloadRequest struct {
	URLs       []string
	SHA256     string // SRI format: "sha256-<base64>" or raw hex
	DestPath   string
	OnProgress func(DownloadProgress)
}

type Downloader interface {
	Download(ctx context.Context, req DownloadRequest) error
}

type HTTPDownloader struct {
	Client *http.Client
}

func NewHTTPDownloader() *HTTPDownloader {
	return &HTTPDownloader{
		Client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("too many redirects")
				}
				return nil
			},
		},
	}
}

func (d *HTTPDownloader) Download(ctx context.Context, req DownloadRequest) error {
	if len(req.URLs) == 0 {
		return fmt.Errorf("no URLs provided")
	}

	var lastErr error
	for _, url := range req.URLs {
		lastErr = d.downloadFromURL(ctx, url, req)
		if lastErr == nil {
			return nil
		}
	}
	return fmt.Errorf("all URLs failed, last error: %w", lastErr)
}

func (d *HTTPDownloader) downloadFromURL(ctx context.Context, url string, req DownloadRequest) error {
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	httpReq.Header.Set("User-Agent", "kyaraben/"+version.Get())

	resp, err := d.Client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("downloading %s: %w", url, err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("downloading %s: status %d", url, resp.StatusCode)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(req.DestPath), "kyaraben-download-*")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpPath := tmpFile.Name()
	defer func() { _ = os.Remove(tmpPath) }()

	hasher := sha256.New()
	var written int64

	var reader io.Reader = resp.Body
	if req.OnProgress != nil {
		reader = &progressReader{
			reader: resp.Body,
			total:  resp.ContentLength,
			onProgress: func(n int64) {
				written += n
				req.OnProgress(DownloadProgress{
					BytesDownloaded: written,
					BytesTotal:      resp.ContentLength,
				})
			},
		}
	}

	writer := io.MultiWriter(tmpFile, hasher)
	if _, err := io.Copy(writer, reader); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("writing download: %w", err)
	}

	if err := tmpFile.Sync(); err != nil {
		_ = tmpFile.Close()
		return fmt.Errorf("syncing download: %w", err)
	}
	if err := tmpFile.Close(); err != nil {
		return fmt.Errorf("closing download: %w", err)
	}

	if req.SHA256 != "" {
		actual := hasher.Sum(nil)
		expected, err := parseSHA256(req.SHA256)
		if err != nil {
			return err
		}
		if string(actual) != string(expected) {
			return fmt.Errorf("sha256 mismatch for %s: expected %x, got %x", url, expected, actual)
		}
	}

	if err := os.Rename(tmpPath, req.DestPath); err != nil {
		return fmt.Errorf("moving download to %s: %w", req.DestPath, err)
	}

	return nil
}

func parseSHA256(hash string) ([]byte, error) {
	if strings.HasPrefix(hash, "sha256-") {
		b64 := strings.TrimPrefix(hash, "sha256-")
		decoded, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return nil, fmt.Errorf("decoding SRI hash: %w", err)
		}
		return decoded, nil
	}
	if len(hash) == 64 {
		decoded, err := hex.DecodeString(hash)
		if err != nil {
			return nil, fmt.Errorf("decoding hex hash: %w", err)
		}
		return decoded, nil
	}
	return nil, fmt.Errorf("invalid sha256 hash format: %s", hash)
}

type progressReader struct {
	reader     io.Reader
	total      int64
	onProgress func(n int64)
}

func (r *progressReader) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if n > 0 {
		r.onProgress(int64(n))
	}
	return n, err
}

var _ Downloader = (*HTTPDownloader)(nil)
