package collections

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/twpayne/go-vfs/v5"
)

const maxManifestBytes = 8 << 20

type Fetcher interface {
	Fetch(ctx context.Context, url string) ([]byte, error)
}

type HTTPFetcher struct {
	Client *http.Client
}

func NewHTTPFetcher() *HTTPFetcher {
	return &HTTPFetcher{Client: &http.Client{Timeout: 15 * time.Second}}
}

func (f *HTTPFetcher) Fetch(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	resp, err := f.Client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetching %s: %s", url, resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxManifestBytes+1))
	if err != nil {
		return nil, err
	}
	if len(data) > maxManifestBytes {
		return nil, fmt.Errorf("fetching %s: manifest larger than %d bytes", url, maxManifestBytes)
	}
	return data, nil
}

// Source fetches manifests and keeps the last valid copy so games stay
// launchable while the publishing server is unreachable.
type Source struct {
	fs       vfs.FS
	fetcher  Fetcher
	cacheDir string
}

func NewSource(fs vfs.FS, fetcher Fetcher, cacheDir string) *Source {
	return &Source{fs: fs, fetcher: fetcher, cacheDir: cacheDir}
}

type Loaded struct {
	Manifest  *Manifest
	FromCache bool
	// FetchErr is set when the cached copy is served because fetching failed.
	FetchErr error
}

// Load returns the cached manifest when it is younger than maxAge, and
// otherwise fetches it, falling back to the cache if fetching fails.
func (s *Source) Load(ctx context.Context, collectionID, url string, maxAge time.Duration) (*Loaded, error) {
	if maxAge > 0 {
		if info, err := s.fs.Stat(s.cachePath(collectionID)); err == nil && time.Since(info.ModTime()) < maxAge {
			if m, err := s.Cached(collectionID); err == nil {
				return &Loaded{Manifest: m, FromCache: true}, nil
			}
		}
	}

	data, fetchErr := s.fetcher.Fetch(ctx, url)
	if fetchErr == nil {
		m, err := s.parse(collectionID, data)
		if err != nil {
			fetchErr = err
		} else {
			if err := s.writeCache(collectionID, data); err != nil {
				return nil, err
			}
			return &Loaded{Manifest: m}, nil
		}
	}

	m, err := s.Cached(collectionID)
	if err != nil {
		return nil, fmt.Errorf("%w (no usable cached copy: %v)", fetchErr, err)
	}
	return &Loaded{Manifest: m, FromCache: true, FetchErr: fetchErr}, nil
}

func (s *Source) Cached(collectionID string) (*Manifest, error) {
	data, err := s.fs.ReadFile(s.cachePath(collectionID))
	if err != nil {
		return nil, err
	}
	return s.parse(collectionID, data)
}

func (s *Source) parse(collectionID string, data []byte) (*Manifest, error) {
	m, err := ParseManifest(data)
	if err != nil {
		return nil, err
	}
	if m.ID != collectionID {
		return nil, fmt.Errorf("manifest is for collection %q, configured as %q", m.ID, collectionID)
	}
	return m, nil
}

func (s *Source) cachePath(collectionID string) string {
	return filepath.Join(s.cacheDir, collectionID+".json")
}

func (s *Source) writeCache(collectionID string, data []byte) error {
	if err := vfs.MkdirAll(s.fs, s.cacheDir, 0o755); err != nil {
		return fmt.Errorf("creating collections cache: %w", err)
	}
	tmp := s.cachePath(collectionID) + ".tmp"
	if err := s.fs.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("caching manifest: %w", err)
	}
	if err := s.fs.Rename(tmp, s.cachePath(collectionID)); err != nil {
		_ = s.fs.Remove(tmp)
		return fmt.Errorf("caching manifest: %w", err)
	}
	return nil
}

func isNotExist(err error) bool {
	return err != nil && os.IsNotExist(err)
}
