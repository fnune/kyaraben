package nix

import (
	"context"
	"encoding/json"
)

// NixClient defines the interface for Nix operations.
// Use this interface when injecting the Nix client as a dependency to enable testing.
type NixClient interface {
	IsAvailable() bool
	Build(ctx context.Context, flakeRef string) (string, error)

	// Preferred over Build when the result needs to be accessible
	// from outside nix-portable's namespace.
	BuildWithLink(ctx context.Context, flakeRef string, outLink string) error

	BuildMultiple(ctx context.Context, flakeRefs []string) (map[string]string, error)
	Eval(ctx context.Context, expr string) (json.RawMessage, error)
	FlakeUpdate(ctx context.Context, flakePath string) error
	GetVersion(ctx context.Context) (string, error)
	EnsureFlakeDir() error
	GetFlakePath() string
	FlakeCheck(ctx context.Context, flakePath string) error

	// Translates virtualized /nix/store paths because nix-portable
	// stores actual files in $NP_LOCATION/.nix-portable/nix/store/.
	RealStorePath(virtualPath string) string

	GetNixPortableBinary() string
	GetNixPortableLocation() string
	SetOutputCallback(fn func(line string))

	GetPersistentNixPortablePath() string
	EnsurePersistentNixPortable() (string, error)
}

func (c *Client) GetFlakePath() string {
	return c.FlakePath
}

// Ensure Client implements NixClient.
var _ NixClient = (*Client)(nil)
