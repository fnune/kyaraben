package model

// PackageRef is the interface for referencing packages from different sources.
// Implementations provide source-specific information needed to fetch/build the package.
type PackageRef interface {
	Source() PackageSource
	PackageName() string // Key for versions.toml lookup
}

// AppImage references an AppImage whose version info is in versions.toml
type AppImage struct {
	Name string // Key in versions.toml (e.g., "eden")
}

func (p AppImage) Source() PackageSource { return PackageSourceAppImage }
func (p AppImage) PackageName() string   { return p.Name }

// AppImageRef creates a PackageRef that reads version info from versions.toml
func AppImageRef(name string) PackageRef {
	return AppImage{Name: name}
}
