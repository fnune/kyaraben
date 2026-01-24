package nix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/fnune/kyaraben/internal/paths"
)

// Client provides an interface to the Nix CLI.
type Client struct {
	// NixPortableBinary is the path to the nix-portable binary.
	NixPortableBinary string

	// NixPortableLocation is the directory where nix-portable stores its data.
	// This is passed via NP_LOCATION environment variable.
	NixPortableLocation string

	// FlakePath is where to write/read the generated flake.
	FlakePath string
}

// NewClient creates a new Nix client with default settings.
// It attempts to find the bundled nix-portable binary. If not found,
// the client is still created but IsAvailable() will return false.
func NewClient() (*Client, error) {
	kyarabenDir, err := paths.KyarabenDataDir()
	if err != nil {
		return nil, err
	}

	// Try to find nix-portable, but don't fail if not found.
	// This allows dry-run and other non-Nix operations to work.
	nixPortable, _ := findNixPortable()

	return &Client{
		NixPortableBinary:   nixPortable,
		NixPortableLocation: filepath.Join(kyarabenDir, "nix-portable"),
		FlakePath:           filepath.Join(kyarabenDir, "flake"),
	}, nil
}

// findNixPortable locates the bundled nix-portable binary.
// It searches relative to the current executable.
func findNixPortable() (string, error) {
	targetTriple := getTargetTriple()
	binaryName := "nix-portable-" + targetTriple

	// Get the directory of the current executable
	execPath, err := os.Executable()
	if err != nil {
		return "", fmt.Errorf("getting executable path: %w", err)
	}
	execDir := filepath.Dir(execPath)

	// Search locations in order of preference:
	// 1. Same directory as executable (AppImage/installed)
	// 2. ../binaries/ relative to executable (development with sidecar structure)
	// 3. ui/src-tauri/binaries/ from project root (development)
	searchPaths := []string{
		filepath.Join(execDir, binaryName),
		filepath.Join(execDir, "..", "binaries", binaryName),
	}

	// For development, also check relative to working directory
	if cwd, err := os.Getwd(); err == nil {
		searchPaths = append(searchPaths,
			filepath.Join(cwd, "ui", "src-tauri", "binaries", binaryName),
		)
	}

	for _, path := range searchPaths {
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}
	}

	return "", fmt.Errorf("nix-portable binary not found (searched for %s)", binaryName)
}

// getTargetTriple returns the Rust target triple for the current platform.
func getTargetTriple() string {
	arch := runtime.GOARCH
	os := runtime.GOOS

	switch os {
	case "linux":
		switch arch {
		case "amd64":
			return "x86_64-unknown-linux-gnu"
		case "arm64":
			return "aarch64-unknown-linux-gnu"
		default:
			return "unknown-unknown-linux-gnu"
		}
	case "darwin":
		switch arch {
		case "amd64":
			return "x86_64-apple-darwin"
		case "arm64":
			return "aarch64-apple-darwin"
		default:
			return "unknown-apple-darwin"
		}
	default:
		return "unknown-unknown-unknown"
	}
}

// IsAvailable checks if the nix-portable binary is available.
func (c *Client) IsAvailable() bool {
	_, err := os.Stat(c.NixPortableBinary)
	return err == nil
}

// runNix executes a nix command through nix-portable.
// Note: nix-portable has nix-command and flakes enabled by default,
// so we don't need --extra-experimental-features flags.
func (c *Client) runNix(ctx context.Context, args []string) (*exec.Cmd, error) {
	if !c.IsAvailable() {
		return nil, fmt.Errorf("nix-portable is not available (bundled binary not found)")
	}

	// nix-portable wraps nix, so we call: nix-portable nix <args>
	fullArgs := append([]string{"nix"}, args...)
	cmd := exec.CommandContext(ctx, c.NixPortableBinary, fullArgs...)

	// Set NP_LOCATION to control where nix-portable stores its data
	cmd.Env = append(os.Environ(), "NP_LOCATION="+c.NixPortableLocation)

	return cmd, nil
}

// Build builds a flake output and returns the store path.
func (c *Client) Build(ctx context.Context, flakeRef string) (string, error) {
	args := []string{
		"build",
		flakeRef,
		"--no-link",
		"--print-out-paths",
	}

	cmd, err := c.runNix(ctx, args)
	if err != nil {
		return "", err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("nix build failed: %w\nstderr: %s", err, stderr.String())
	}

	// Parse output - it's the store path
	storePath := strings.TrimSpace(stdout.String())
	if storePath == "" {
		return "", fmt.Errorf("nix build produced no output")
	}

	return storePath, nil
}

// BuildMultiple builds multiple flake outputs.
func (c *Client) BuildMultiple(ctx context.Context, flakeRefs []string) (map[string]string, error) {
	results := make(map[string]string)

	for _, ref := range flakeRefs {
		path, err := c.Build(ctx, ref)
		if err != nil {
			return nil, fmt.Errorf("building %s: %w", ref, err)
		}
		results[ref] = path
	}

	return results, nil
}

// Eval evaluates a Nix expression and returns the result as JSON.
func (c *Client) Eval(ctx context.Context, expr string) (json.RawMessage, error) {
	args := []string{
		"eval",
		"--json",
		"--expr", expr,
	}

	cmd, err := c.runNix(ctx, args)
	if err != nil {
		return nil, err
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("nix eval failed: %w\nstderr: %s", err, stderr.String())
	}

	return json.RawMessage(stdout.Bytes()), nil
}

// FlakeUpdate updates flake inputs.
func (c *Client) FlakeUpdate(ctx context.Context, flakePath string) error {
	args := []string{
		"flake",
		"update",
	}

	cmd, err := c.runNix(ctx, args)
	if err != nil {
		return err
	}
	cmd.Dir = flakePath
	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("nix flake update failed: %w\nstderr: %s", err, stderr.String())
	}

	return nil
}

// GetVersion returns the version of the nix binary.
func (c *Client) GetVersion(ctx context.Context) (string, error) {
	cmd, err := c.runNix(ctx, []string{"--version"})
	if err != nil {
		return "", err
	}
	var stdout bytes.Buffer
	cmd.Stdout = &stdout

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("getting nix version: %w", err)
	}

	return strings.TrimSpace(stdout.String()), nil
}

// EnsureFlakeDir ensures the flake directory exists.
func (c *Client) EnsureFlakeDir() error {
	return os.MkdirAll(c.FlakePath, 0755)
}
