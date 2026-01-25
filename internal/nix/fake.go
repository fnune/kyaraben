package nix

import (
	"context"
	"encoding/json"
	"fmt"
)

// FakeClient is a fake NixClient for testing.
// Configure its behavior by setting fields before use.
type FakeClient struct {
	// Available controls whether IsAvailable returns true.
	Available bool

	// FlakePath is returned by GetFlakePath and EnsureFlakeDir creates it.
	FlakePathValue string

	// BuildResults maps flake references to store paths.
	// If a ref is not in the map and BuildError is nil, Build returns a generated path.
	BuildResults map[string]string

	// BuildError is returned by Build if set.
	BuildError error

	// EvalResults maps expressions to JSON results.
	EvalResults map[string]json.RawMessage

	// EvalError is returned by Eval if set.
	EvalError error

	// Version is returned by GetVersion.
	Version string

	// BuildCalls records all calls to Build.
	BuildCalls []string

	// EvalCalls records all calls to Eval.
	EvalCalls []string
}

// NewFakeClient creates a FakeClient with sensible defaults for testing.
func NewFakeClient() *FakeClient {
	return &FakeClient{
		Available:      true,
		FlakePathValue: "/tmp/fake-flake",
		BuildResults:   make(map[string]string),
		EvalResults:    make(map[string]json.RawMessage),
		Version:        "nix (Nix) 2.18.0",
	}
}

func (f *FakeClient) IsAvailable() bool {
	return f.Available
}

func (f *FakeClient) Build(ctx context.Context, flakeRef string) (string, error) {
	f.BuildCalls = append(f.BuildCalls, flakeRef)

	if f.BuildError != nil {
		return "", f.BuildError
	}

	if path, ok := f.BuildResults[flakeRef]; ok {
		return path, nil
	}

	// Return a generated path if not explicitly configured
	return fmt.Sprintf("/nix/store/fake-hash-%s", flakeRef), nil
}

func (f *FakeClient) BuildMultiple(ctx context.Context, flakeRefs []string) (map[string]string, error) {
	results := make(map[string]string)
	for _, ref := range flakeRefs {
		path, err := f.Build(ctx, ref)
		if err != nil {
			return nil, err
		}
		results[ref] = path
	}
	return results, nil
}

func (f *FakeClient) Eval(ctx context.Context, expr string) (json.RawMessage, error) {
	f.EvalCalls = append(f.EvalCalls, expr)

	if f.EvalError != nil {
		return nil, f.EvalError
	}

	if result, ok := f.EvalResults[expr]; ok {
		return result, nil
	}

	return json.RawMessage(`null`), nil
}

func (f *FakeClient) FlakeUpdate(ctx context.Context, flakePath string) error {
	return nil
}

func (f *FakeClient) GetVersion(ctx context.Context) (string, error) {
	return f.Version, nil
}

func (f *FakeClient) EnsureFlakeDir() error {
	return nil
}

func (f *FakeClient) GetFlakePath() string {
	return f.FlakePathValue
}

func (f *FakeClient) FlakeCheck(ctx context.Context, flakePath string) error {
	return nil
}

// Ensure FakeClient implements NixClient.
var _ NixClient = (*FakeClient)(nil)
