package nix

import (
	"context"
	"encoding/json"
	"fmt"
)

type FakeClient struct {
	Available      bool
	FlakePathValue string
	BuildResults   map[string]string
	BuildError     error
	EvalResults    map[string]json.RawMessage
	EvalError      error
	Version        string
	BuildCalls     []string
	EvalCalls      []string
}

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

	return fmt.Sprintf("/nix/store/fake-hash-%s", flakeRef), nil
}

func (f *FakeClient) BuildWithLink(ctx context.Context, flakeRef string, outLink string) error {
	f.BuildCalls = append(f.BuildCalls, flakeRef)
	return f.BuildError
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

func (f *FakeClient) RealStorePath(virtualPath string) string {
	return virtualPath
}

func (f *FakeClient) GetNixPortableBinary() string {
	return "/fake/nix-portable"
}

func (f *FakeClient) GetNixPortableLocation() string {
	return "/fake/nix-portable-location"
}

func (f *FakeClient) SetOutputCallback(fn func(line string)) {
}

func (f *FakeClient) SetProgressCallback(fn func(BuildProgress)) {
}

func (f *FakeClient) SetExpectedPackages(packages []ExpectedPackage) {
}

func (f *FakeClient) GetPersistentNixPortablePath() string {
	return "/fake/persistent-nix-portable"
}

func (f *FakeClient) EnsurePersistentNixPortable() (string, error) {
	return f.GetPersistentNixPortablePath(), nil
}

func (f *FakeClient) GarbageCollect(ctx context.Context) error {
	return nil
}

// Ensure FakeClient implements NixClient.
var _ NixClient = (*FakeClient)(nil)
