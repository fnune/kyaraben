package packages

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/testutil"
)

func TestHTTPDownloaderSuccess(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	content := []byte("hello world")
	hash := sha256.Sum256(content)
	sri := "sha256-" + base64.StdEncoding.EncodeToString(hash[:])

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", "11")
		_, _ = w.Write(content)
	}))
	defer server.Close()

	dest := "/downloads/output"
	dl := NewDownloader(fs)

	err := dl.Download(context.Background(), DownloadRequest{
		URLs:     []string{server.URL},
		SHA256:   sri,
		DestPath: dest,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := fs.ReadFile(dest)
	if err != nil {
		t.Fatalf("reading output: %v", err)
	}
	if string(got) != string(content) {
		t.Errorf("content = %q, want %q", got, content)
	}
}

func TestHTTPDownloaderSHA256Mismatch(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("wrong content"))
	}))
	defer server.Close()

	hash := sha256.Sum256([]byte("expected content"))
	sri := "sha256-" + base64.StdEncoding.EncodeToString(hash[:])

	dest := "/downloads/output"
	dl := NewDownloader(fs)

	err := dl.Download(context.Background(), DownloadRequest{
		URLs:     []string{server.URL},
		SHA256:   sri,
		DestPath: dest,
	})
	if err == nil {
		t.Fatal("expected sha256 mismatch error")
	}
}

func TestHTTPDownloaderFallbackURLs(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	dest := "/downloads/output"
	dl := NewDownloader(fs)

	err := dl.Download(context.Background(), DownloadRequest{
		URLs:     []string{server.URL + "/bad", server.URL + "/good"},
		SHA256:   "2689367b205c16ce32ed4200942b8b8b1e262dfc70d9bc9fbc77c49699a4f1df",
		DestPath: dest,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestHTTPDownloaderProgress(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	content := []byte("hello world test content for progress")
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(content)
	}))
	defer server.Close()

	dest := "/downloads/output"
	dl := NewDownloader(fs)

	var called atomic.Int32
	err := dl.Download(context.Background(), DownloadRequest{
		URLs:     []string{server.URL},
		SHA256:   "ee0c51f7bb7946c79540d76ad4f33ea1264c20b01a0f1908fb8b82d0ce6ac3b7",
		DestPath: dest,
		OnProgress: func(p DownloadProgress) {
			called.Add(1)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called.Load() == 0 {
		t.Error("progress callback was never called")
	}
}

func TestHTTPDownloaderCancellation(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	defer server.Close()

	dest := "/downloads/output"
	dl := NewDownloader(fs)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := dl.Download(ctx, DownloadRequest{
		URLs:     []string{server.URL},
		DestPath: dest,
	})
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}

func TestHTTPDownloaderNoURLs(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	dl := NewDownloader(fs)
	err := dl.Download(context.Background(), DownloadRequest{
		DestPath: "/downloads/output",
	})
	if err == nil {
		t.Fatal("expected error for empty URLs")
	}
}

func TestParseSHA256SRI(t *testing.T) {
	t.Parallel()

	raw := sha256.Sum256([]byte("test"))
	sri := "sha256-" + base64.StdEncoding.EncodeToString(raw[:])

	decoded, err := parseSHA256(sri)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(decoded) != string(raw[:]) {
		t.Errorf("decoded hash mismatch")
	}
}

func TestParseSHA256Hex(t *testing.T) {
	t.Parallel()

	raw := sha256.Sum256([]byte("test"))
	hex := ""
	for _, b := range raw {
		hex += string("0123456789abcdef"[b>>4]) + string("0123456789abcdef"[b&0xf])
	}

	decoded, err := parseSHA256(hex)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(decoded) != string(raw[:]) {
		t.Errorf("decoded hash mismatch")
	}
}

func TestDownloadErrorReturnsStatusCode(t *testing.T) {
	t.Parallel()

	fs := testutil.NewTestFS(t, map[string]any{
		"/downloads": &vfst.Dir{Perm: 0755},
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	dl := NewDownloader(fs)
	err := dl.Download(context.Background(), DownloadRequest{
		URLs:     []string{server.URL},
		SHA256:   "ignored",
		DestPath: "/downloads/output",
	})

	var dlErr *DownloadError
	if !errors.As(err, &dlErr) {
		t.Fatalf("expected DownloadError, got %T", err)
	}
	if dlErr.StatusCode != 503 {
		t.Errorf("StatusCode = %d, want 503", dlErr.StatusCode)
	}
}

func TestIsTransientError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"rate limited", 429, true},
		{"server error", 500, true},
		{"bad gateway", 502, true},
		{"service unavailable", 503, true},
		{"gateway timeout", 504, true},
		{"cloudflare error", 520, true},
		{"not found", 404, false},
		{"forbidden", 403, false},
		{"ok", 200, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &DownloadError{URL: "http://example.com", StatusCode: tt.statusCode}
			if got := IsTransientError(err); got != tt.want {
				t.Errorf("IsTransientError(%d) = %v, want %v", tt.statusCode, got, tt.want)
			}
		})
	}
}

func TestIsTransientErrorWithWrappedError(t *testing.T) {
	t.Parallel()

	dlErr := &DownloadError{URL: "http://example.com", StatusCode: 503}
	wrapped := fmt.Errorf("installing package: %w", dlErr)

	if !IsTransientError(wrapped) {
		t.Error("IsTransientError should find wrapped DownloadError")
	}
}
