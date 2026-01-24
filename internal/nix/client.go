package nix

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/paths"
)

// Client provides an interface to the Nix CLI.
type Client struct {
	// NixBinary is the path to the nix binary. Defaults to "nix".
	NixBinary string

	// StorePath is where to store Nix data. Defaults to XDG data directory.
	StorePath string

	// FlakePath is where to write/read the generated flake.
	FlakePath string
}

// NewClient creates a new Nix client with default settings.
func NewClient() (*Client, error) {
	kyarabenDir, err := paths.KyarabenDataDir()
	if err != nil {
		return nil, err
	}

	return &Client{
		NixBinary: "nix",
		StorePath: filepath.Join(kyarabenDir, "store"),
		FlakePath: filepath.Join(kyarabenDir, "flake"),
	}, nil
}

// IsAvailable checks if the nix binary is available.
func (c *Client) IsAvailable() bool {
	_, err := exec.LookPath(c.NixBinary)
	return err == nil
}

// Build builds a flake output and returns the store path.
func (c *Client) Build(ctx context.Context, flakeRef string) (string, error) {
	args := []string{
		"build",
		flakeRef,
		"--no-link",
		"--print-out-paths",
		"--extra-experimental-features", "nix-command flakes",
	}

	cmd := exec.CommandContext(ctx, c.NixBinary, args...)
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
		"--extra-experimental-features", "nix-command flakes",
	}

	cmd := exec.CommandContext(ctx, c.NixBinary, args...)
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
		"--extra-experimental-features", "nix-command flakes",
	}

	cmd := exec.CommandContext(ctx, c.NixBinary, args...)
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
	cmd := exec.CommandContext(ctx, c.NixBinary, "--version")
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
