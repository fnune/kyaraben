package nix

import (
	"context"
	"encoding/json"
)

// NixClient defines the interface for Nix operations.
// Use this interface when injecting the Nix client as a dependency to enable testing.
type NixClient interface {
	// IsAvailable returns true if nix-portable is available.
	IsAvailable() bool

	// Build builds a flake reference and returns the store path.
	Build(ctx context.Context, flakeRef string) (string, error)

	// BuildMultiple builds multiple flake references.
	BuildMultiple(ctx context.Context, flakeRefs []string) (map[string]string, error)

	// Eval evaluates a Nix expression and returns the JSON result.
	Eval(ctx context.Context, expr string) (json.RawMessage, error)

	// FlakeUpdate updates the flake lock file.
	FlakeUpdate(ctx context.Context, flakePath string) error

	// GetVersion returns the Nix version string.
	GetVersion(ctx context.Context) (string, error)

	// EnsureFlakeDir creates the flake directory if it doesn't exist.
	EnsureFlakeDir() error

	// GetFlakePath returns the path to the flake directory.
	GetFlakePath() string

	// FlakeCheck validates a flake without building it.
	FlakeCheck(ctx context.Context, flakePath string) error
}

// GetFlakePath returns the flake path for the Client.
func (c *Client) GetFlakePath() string {
	return c.FlakePath
}

// Ensure Client implements NixClient.
var _ NixClient = (*Client)(nil)
