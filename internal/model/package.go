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

// GitHubPackage references a GitHub release asset.
type GitHubPackage struct {
	Owner string // e.g., "stenzek"
	Repo  string // e.g., "duckstation"
	Asset string // e.g., "DuckStation-x64.AppImage"
}

func (p GitHubPackage) Source() PackageSource {
	return PackageSourceGitHub
}

// NixpkgsRef creates a PackageRef for a simple nixpkgs package.
func NixpkgsRef(attr string) PackageRef {
	return NixpkgsPackage{Attr: attr}
}

// NixpkgsOverlayRef creates a PackageRef for a nixpkgs package with an overlay.
func NixpkgsOverlayRef(attr, overlay string) PackageRef {
	return NixpkgsPackage{Attr: attr, Overlay: overlay}
}

// GitHubRef creates a PackageRef for a GitHub release.
func GitHubRef(owner, repo, asset string) PackageRef {
	return GitHubPackage{Owner: owner, Repo: repo, Asset: asset}
}
