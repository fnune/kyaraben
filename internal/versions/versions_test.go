package versions

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if err := Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestURLTemplateExpansion(t *testing.T) {
	v := MustGet()
	spec, ok := v.GetPackage("eden")
	if !ok {
		t.Fatal("eden package not found")
	}
	entry := spec.GetVersion("v0.1.0")
	if entry == nil {
		t.Fatal("eden v0.1.0 not found")
	}

	tests := []struct {
		target   string
		expected string
	}{
		{"x64", "https://stable.eden-emu.dev/v0.1.0/Eden-Linux-v0.1.0-amd64-clang-pgo.AppImage"},
		{"steamdeck", "https://stable.eden-emu.dev/v0.1.0/Eden-Linux-v0.1.0-steamdeck-clang-pgo.AppImage"},
		{"rog-ally", "https://stable.eden-emu.dev/v0.1.0/Eden-Linux-v0.1.0-rog-ally-clang-pgo.AppImage"},
		{"aarch64", "https://stable.eden-emu.dev/v0.1.0/Eden-Linux-v0.1.0-aarch64-clang-pgo.AppImage"},
	}

	for _, tt := range tests {
		url := entry.URL(tt.target, spec)
		if url != tt.expected {
			t.Errorf("URL(%s) mismatch:\ngot:  %s\nwant: %s", tt.target, url, tt.expected)
		}
	}
}

func TestSelectTarget(t *testing.T) {
	v := MustGet()
	spec, ok := v.GetPackage("eden")
	if !ok {
		t.Fatal("eden package not found")
	}
	entry := spec.GetDefault()
	if entry == nil {
		t.Fatal("eden default version not found")
	}

	tests := []struct {
		name     string
		arch     string
		expected string
	}{
		{"steamdeck", "x86_64", "steamdeck"},
		{"rog-ally", "x86_64", "rog-ally"},
		{"aarch64", "aarch64", "aarch64"},
		{"x64", "x86_64", "x64"},
		{"unknown", "x86_64", "x64"},
		{"unknown", "aarch64", "aarch64"},
		{"unknown", "unknown", ""},
	}

	for _, tt := range tests {
		got := entry.SelectTarget(tt.name, tt.arch)
		if got != tt.expected {
			t.Errorf("SelectTarget(%s, %s) = %q, want %q", tt.name, tt.arch, got, tt.expected)
		}
	}
}

func TestArchiveType(t *testing.T) {
	v := MustGet()

	tests := []struct {
		name     string
		pkg      string
		target   string
		expected string
	}{
		{"eden is direct AppImage", "eden", "x64", ""},
		{"duckstation is direct AppImage", "duckstation", "x64", ""},
		{"retroarch is 7z", "retroarch", "x64", "7z"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := v.GetPackage(tt.pkg)
			if !ok {
				t.Fatalf("package %s not found", tt.pkg)
			}
			entry := spec.GetDefault()
			if entry == nil {
				t.Fatalf("default version for %s not found", tt.pkg)
			}
			got := entry.ArchiveType(tt.target, spec)
			if got != tt.expected {
				t.Errorf("ArchiveType(%s) = %q, want %q", tt.target, got, tt.expected)
			}
		})
	}
}

func TestBinaryPathForTarget(t *testing.T) {
	v := MustGet()

	tests := []struct {
		name     string
		pkg      string
		target   string
		expected string
	}{
		{"eden has no binary path (direct)", "eden", "x64", ""},
		{"retroarch x64 has per-target binary path", "retroarch", "x64", "RetroArch-Linux-x86_64/RetroArch-Linux-x86_64.AppImage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := v.GetPackage(tt.pkg)
			if !ok {
				t.Fatalf("package %s not found", tt.pkg)
			}
			entry := spec.GetDefault()
			if entry == nil {
				t.Fatalf("default version for %s not found", tt.pkg)
			}
			got := entry.BinaryPathForTarget(tt.target, spec)
			if got != tt.expected {
				t.Errorf("BinaryPathForTarget(%s) = %q, want %q", tt.target, got, tt.expected)
			}
		})
	}
}

func TestEffectiveReleaseTag(t *testing.T) {
	v := MustGet()

	// Eden has no release_tag, should use version
	edenSpec, _ := v.GetPackage("eden")
	edenEntry := edenSpec.GetDefault()
	if got := edenEntry.EffectiveReleaseTag(); got != edenEntry.Version {
		t.Errorf("Eden.EffectiveReleaseTag() = %q, want version %q", got, edenEntry.Version)
	}

	// Dolphin has a release_tag that differs from version
	dolphinSpec, _ := v.GetPackage("dolphin")
	dolphinEntry := dolphinSpec.GetDefault()
	if dolphinEntry.ReleaseTag != "" && dolphinEntry.ReleaseTag != dolphinEntry.Version {
		if got := dolphinEntry.EffectiveReleaseTag(); got != dolphinEntry.ReleaseTag {
			t.Errorf("Dolphin.EffectiveReleaseTag() = %q, want release_tag %q", got, dolphinEntry.ReleaseTag)
		}
	}
}

func TestURLWithReleaseTag(t *testing.T) {
	v := MustGet()

	// Dolphin URL should use release_tag substitution
	spec, _ := v.GetPackage("dolphin")
	entry := spec.GetDefault()
	url := entry.URL("x64", spec)
	if entry.ReleaseTag != "" {
		// URL should contain the release tag, not just the version
		if !strings.Contains(url, entry.ReleaseTag) {
			t.Errorf("Dolphin URL should contain release_tag %q, got: %s", entry.ReleaseTag, url)
		}
	}
}

func TestMultipleVersions(t *testing.T) {
	v := MustGet()

	spec, ok := v.GetPackage("eden")
	if !ok {
		t.Fatal("eden package not found")
	}

	if len(spec.Versions) < 2 {
		t.Errorf("expected at least 2 eden versions, got %d", len(spec.Versions))
	}

	if spec.Default == "" {
		t.Fatal("default version is empty")
	}
	if _, ok := spec.Versions[spec.Default]; !ok {
		t.Errorf("default version %q not present in Versions map", spec.Default)
	}

	defaultEntry := spec.GetDefault()
	if defaultEntry == nil {
		t.Fatal("GetDefault() returned nil")
	}
	if defaultEntry.Version != spec.Default {
		t.Errorf("GetDefault().Version = %q, want %q", defaultEntry.Version, spec.Default)
	}

	lookedUp := spec.GetVersion(spec.Default)
	if lookedUp == nil || lookedUp.Version != defaultEntry.Version {
		t.Errorf("GetVersion(default) returned different entry than GetDefault()")
	}

	if spec.GetVersion("this-version-does-not-exist") != nil {
		t.Error("GetVersion(nonexistent) should return nil")
	}
}

func TestAvailableVersionsPreservesDefinitionOrder(t *testing.T) {
	const fixture = `
[xenia-edge]
releases_url = "github:has207/xenia-edge"
url_template = "https://example.invalid/{release_tag}/xenia.AppImage"
default = "zzz"

[xenia-edge."zzz"]
release_tag = "zzz"

[xenia-edge."zzz".targets.x64]
sha256 = "sha256-AAAA="
size = 1

[xenia-edge."aaa"]
release_tag = "aaa"

[xenia-edge."aaa".targets.x64]
sha256 = "sha256-BBBB="
size = 2

[xenia-edge."mmm"]
release_tag = "mmm"

[xenia-edge."mmm".targets.x64]
sha256 = "sha256-CCCC="
size = 3
`

	v, err := parse(fixture)
	if err != nil {
		t.Fatalf("parse fixture: %v", err)
	}
	spec, ok := v.GetPackage("xenia-edge")
	if !ok {
		t.Fatal("xenia-edge package not found in fixture")
	}

	got := spec.AvailableVersions()
	want := []string{"zzz", "aaa", "mmm"}
	if len(got) != len(want) {
		t.Fatalf("got %d versions, want %d: %v", len(got), len(want), got)
	}
	for i, w := range want {
		if got[i] != w {
			t.Errorf("AvailableVersions()[%d] = %q, want %q (full: %v)", i, got[i], w, got)
		}
	}
}

func TestAllPackagesHaveSize(t *testing.T) {
	v := MustGet()
	for name, spec := range v.Packages {
		for version, entry := range spec.Versions {
			for target, build := range entry.Targets {
				if build.Size == 0 {
					t.Errorf("%s %s %s: missing size", name, version, target)
				}
			}
		}
	}
}

func TestAllPackagesHaveSHA256(t *testing.T) {
	v := MustGet()
	for name, spec := range v.Packages {
		for version, entry := range spec.Versions {
			for target, build := range entry.Targets {
				if build.SHA256 == "" {
					t.Errorf("%s %s %s: missing SHA256", name, version, target)
				}
			}
		}
	}
}

func TestGetPackage(t *testing.T) {
	v := MustGet()

	// Test known packages
	knownPackages := []string{
		"eden", "duckstation", "pcsx2", "ppsspp",
		"cemu", "dolphin", "vita3k",
		"rpcs3", "flycast", "retroarch", "syncthing",
	}

	for _, name := range knownPackages {
		spec, ok := v.GetPackage(name)
		if !ok {
			t.Errorf("GetPackage(%s) returned false", name)
			continue
		}
		if spec.Default == "" {
			t.Errorf("GetPackage(%s) has empty default", name)
		}
		if spec.URLTemplate == "" {
			t.Errorf("GetPackage(%s) has empty url_template", name)
		}
		if len(spec.Versions) == 0 {
			t.Errorf("GetPackage(%s) has no versions", name)
		}
	}

	// Test unknown package
	_, ok := v.GetPackage("nonexistent")
	if ok {
		t.Error("GetPackage(nonexistent) should return false")
	}
}

func TestCoresArePackages(t *testing.T) {
	v := MustGet()

	knownCores := []string{"bsnes", "mesen", "genesis_plus_gx", "mupen64plus_next", "mednafen_saturn"}
	for _, core := range knownCores {
		spec, ok := v.GetPackage(core)
		if !ok {
			t.Errorf("core %s not found as package", core)
			continue
		}
		if !spec.IsRetroArchCore() {
			t.Errorf("core %s should have install_type=retroarch-core", core)
		}
		if spec.BinaryPath == "" {
			t.Errorf("core %s missing binary_path", core)
		}
		entry := spec.GetDefault()
		if entry == nil {
			t.Errorf("core %s has no default version", core)
			continue
		}
		target := entry.SelectTarget("x64", "x86_64")
		if target == "" {
			t.Errorf("core %s has no x64 target", core)
			continue
		}
		build := entry.Target(target)
		if build == nil || build.Size <= 0 {
			t.Errorf("core %s has no size", core)
		}
	}
}

func TestVersionsTomlIntegrity(t *testing.T) {
	t.Parallel()
	v := MustGet()

	for name, spec := range v.Packages {
		spec := spec
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			entry := spec.GetDefault()
			if entry == nil {
				t.Fatal("no default version")
			}
			if len(entry.Targets) == 0 {
				t.Fatal("no targets defined")
			}

			target := entry.SelectTarget("x64", "x86_64")
			if target == "" {
				t.Errorf("no target for x64")
			}

			first := entry.SelectTarget("x64", "x86_64")
			for i := 0; i < 10; i++ {
				if entry.SelectTarget("x64", "x86_64") != first {
					t.Error("SelectTarget is non-deterministic")
					break
				}
			}

			for targetName := range entry.Targets {
				url := entry.URL(targetName, &spec)
				if url == "" {
					t.Errorf("target %s: empty URL", targetName)
				}
				if !strings.HasPrefix(url, "https://") {
					t.Errorf("target %s: URL not https: %s", targetName, url)
				}
				if strings.Contains(url, "{") {
					t.Errorf("target %s: unsubstituted placeholder in URL: %s", targetName, url)
				}

				ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
				size, err := FetchSize(ctx, url)
				cancel()
				if err != nil {
					t.Errorf("target %s: URL not reachable: %v", targetName, err)
				} else if size == 0 {
					t.Errorf("target %s: URL returned zero size", targetName)
				}
			}
		})
	}

}
