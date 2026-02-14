package versions

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

// FetchSize fetches the download size for a URL using an HTTP HEAD request.
func FetchSize(ctx context.Context, url string) (int64, error) {
	client := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return fmt.Errorf("too many redirects")
			}
			return nil
		},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodHead, url, nil)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("HEAD request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	if resp.ContentLength < 0 {
		return 0, fmt.Errorf("server did not provide Content-Length")
	}

	return resp.ContentLength, nil
}

// SizeInfo contains size information for an emulator.
type SizeInfo struct {
	Name    string
	Version string
	Target  string
	URL     string
	Size    int64
	Error   error
}

// FetchAllSizes fetches sizes for all emulators in versions.toml.
// Returns a channel of SizeInfo results as they complete.
func FetchAllSizes(ctx context.Context) <-chan SizeInfo {
	results := make(chan SizeInfo)

	go func() {
		defer close(results)

		v, err := Get()
		if err != nil {
			results <- SizeInfo{Error: err}
			return
		}

		for name, spec := range v.Packages {
			for version, entry := range spec.Versions {
				for targetName := range entry.Targets {
					select {
					case <-ctx.Done():
						return
					default:
					}

					url := entry.URL(targetName, &spec)
					size, fetchErr := FetchSize(ctx, url)

					results <- SizeInfo{
						Name:    name,
						Version: version,
						Target:  targetName,
						URL:     url,
						Size:    size,
						Error:   fetchErr,
					}
				}
			}
		}
	}()

	return results
}

// FormatSize formats bytes as a human-readable string.
func FormatSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
