package model

// PackageRef is the interface for referencing packages from different sources.
// Implementations provide source-specific information needed to fetch/build the package.
type PackageRef interface {
	Source() PackageSource
}

// NixpkgsPackage references a package from nixpkgs.
type NixpkgsPackage struct {
	Attr    string // e.g., "duckstation", "tic-80"
	Overlay string // Optional overlay expression for complex packages
}

func (p NixpkgsPackage) Source() PackageSource {
	return PackageSourceNixpkgs
}

// GitHubAppImage references a GitHub release AppImage asset.
// Deprecated: Use VersionedAppImage instead, which reads from versions.toml
type GitHubAppImage struct {
	Name    string            // Package name for nix (e.g., "eden")
	Owner   string            // e.g., "eden-emulator"
	Repo    string            // e.g., "Releases"
	Version string            // e.g., "v0.0.4"
	Assets  map[string]string // arch -> asset filename (e.g., "x86_64" -> "Eden-Linux-v0.0.4-amd64-clang-pgo.AppImage")
	Hashes  map[string]string // arch -> sha256 hash
}

func (p GitHubAppImage) Source() PackageSource {
	return PackageSourceGitHub
}

// VersionedAppImage references an AppImage whose version info is in versions.toml
type VersionedAppImage struct {
	Name string // Key in versions.toml (e.g., "eden")
}

func (p VersionedAppImage) Source() PackageSource {
	return PackageSourceVersioned
}

// NixpkgsRef creates a PackageRef for a simple nixpkgs package.
func NixpkgsRef(attr string) PackageRef {
	return NixpkgsPackage{Attr: attr}
}

// NixpkgsOverlayRef creates a PackageRef for a nixpkgs package with an overlay.
func NixpkgsOverlayRef(attr, overlay string) PackageRef {
	return NixpkgsPackage{Attr: attr, Overlay: overlay}
}

// GitHubAppImageRef creates a PackageRef for a GitHub release AppImage.
// Deprecated: Use VersionedAppImageRef instead
func GitHubAppImageRef(name, owner, repo, version string, assets, hashes map[string]string) PackageRef {
	return GitHubAppImage{
		Name:    name,
		Owner:   owner,
		Repo:    repo,
		Version: version,
		Assets:  assets,
		Hashes:  hashes,
	}
}

// VersionedAppImageRef creates a PackageRef that reads version info from versions.toml
func VersionedAppImageRef(name string) PackageRef {
	return VersionedAppImage{Name: name}
}
