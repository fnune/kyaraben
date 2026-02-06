package versions

import (
	"os"
	"strings"
	"testing"
)

func TestMain(m *testing.M) {
	if err := Init(); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

func TestURLTemplateExpansion(t *testing.T) {
	v := MustGet()
	spec, ok := v.GetEmulator("eden")
	if !ok {
		t.Fatal("eden emulator not found")
	}
	entry := spec.GetVersion("v0.1.0")
	if entry == nil {
		t.Fatal("eden v0.1.0 not found")
	}

	tests := []struct {
		target   string
		expected string
	}{
		{"amd64", "https://github.com/eden-emulator/Releases/releases/download/v0.1.0/Eden-Linux-v0.1.0-amd64-clang-pgo.AppImage"},
		{"legacy", "https://github.com/eden-emulator/Releases/releases/download/v0.1.0/Eden-Linux-v0.1.0-legacy-clang-pgo.AppImage"},
		{"steamdeck", "https://github.com/eden-emulator/Releases/releases/download/v0.1.0/Eden-Linux-v0.1.0-steamdeck-clang-pgo.AppImage"},
		{"rog-ally", "https://github.com/eden-emulator/Releases/releases/download/v0.1.0/Eden-Linux-v0.1.0-rog-ally-clang-pgo.AppImage"},
		{"aarch64", "https://github.com/eden-emulator/Releases/releases/download/v0.1.0/Eden-Linux-v0.1.0-aarch64-clang-pgo.AppImage"},
	}

	for _, tt := range tests {
		url := entry.URL(tt.target, spec)
		if url != tt.expected {
			t.Errorf("URL(%s) mismatch:\ngot:  %s\nwant: %s", tt.target, url, tt.expected)
		}
	}
}

func TestDefaultTargetForArch(t *testing.T) {
	v := MustGet()
	spec, ok := v.GetEmulator("eden")
	if !ok {
		t.Fatal("eden emulator not found")
	}
	entry := spec.GetDefault()
	if entry == nil {
		t.Fatal("eden default version not found")
	}

	x86Target := entry.DefaultTargetForArch("x86_64")
	if x86Target == "" {
		t.Error("DefaultTargetForArch(x86_64) returned empty string")
	}
	if target := entry.Target(x86Target); target == nil || target.Arch != "x86_64" {
		t.Errorf("DefaultTargetForArch(x86_64) = %s, which is not a valid x86_64 target", x86Target)
	}

	// Test aarch64 with v0.1.0 which has that target
	entry010 := spec.GetVersion("v0.1.0")
	if entry010 == nil {
		t.Fatal("eden v0.1.0 not found")
	}
	armTarget := entry010.DefaultTargetForArch("aarch64")
	if armTarget == "" {
		t.Error("DefaultTargetForArch(aarch64) returned empty string for v0.1.0")
	}
	if target := entry010.Target(armTarget); target == nil || target.Arch != "aarch64" {
		t.Errorf("DefaultTargetForArch(aarch64) = %s, which is not a valid aarch64 target", armTarget)
	}
}

func TestTargetsForArch(t *testing.T) {
	v := MustGet()
	spec, ok := v.GetEmulator("eden")
	if !ok {
		t.Fatal("eden emulator not found")
	}

	// Test with v0.1.0 which has multiple targets
	entry := spec.GetVersion("v0.1.0")
	if entry == nil {
		t.Fatal("eden v0.1.0 not found")
	}

	x86Targets := entry.TargetsForArch("x86_64")
	if len(x86Targets) != 4 {
		t.Errorf("expected 4 x86_64 targets for v0.1.0, got %d: %v", len(x86Targets), x86Targets)
	}

	armTargets := entry.TargetsForArch("aarch64")
	if len(armTargets) != 1 {
		t.Errorf("expected 1 aarch64 target for v0.1.0, got %d: %v", len(armTargets), armTargets)
	}
}

func TestArchiveType(t *testing.T) {
	v := MustGet()

	tests := []struct {
		name     string
		emulator string
		target   string
		expected string
	}{
		{"eden is direct AppImage", "eden", "amd64", ""},
		{"duckstation is direct AppImage", "duckstation", "x64", ""},
		{"retroarch is 7z", "retroarch", "x86_64", "7z"},
		{"melonds is zip", "melonds", "x86_64", "zip"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := v.GetEmulator(tt.emulator)
			if !ok {
				t.Fatalf("emulator %s not found", tt.emulator)
			}
			entry := spec.GetDefault()
			if entry == nil {
				t.Fatalf("default version for %s not found", tt.emulator)
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
		emulator string
		target   string
		expected string
	}{
		{"eden has no binary path (direct)", "eden", "amd64", ""},
		{"retroarch x86_64 has per-target binary path", "retroarch", "x86_64", "RetroArch-Linux-x86_64/RetroArch-Linux-x86_64.AppImage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec, ok := v.GetEmulator(tt.emulator)
			if !ok {
				t.Fatalf("emulator %s not found", tt.emulator)
			}
			entry := spec.GetDefault()
			if entry == nil {
				t.Fatalf("default version for %s not found", tt.emulator)
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
	edenSpec, _ := v.GetEmulator("eden")
	edenEntry := edenSpec.GetDefault()
	if got := edenEntry.EffectiveReleaseTag(); got != edenEntry.Version {
		t.Errorf("Eden.EffectiveReleaseTag() = %q, want version %q", got, edenEntry.Version)
	}

	// Dolphin has a release_tag that differs from version
	dolphinSpec, _ := v.GetEmulator("dolphin")
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
	spec, _ := v.GetEmulator("dolphin")
	entry := spec.GetDefault()
	url := entry.URL("x86_64", spec)
	if entry.ReleaseTag != "" {
		// URL should contain the release tag, not just the version
		if !strings.Contains(url, entry.ReleaseTag) {
			t.Errorf("Dolphin URL should contain release_tag %q, got: %s", entry.ReleaseTag, url)
		}
	}
}

func TestMultipleVersions(t *testing.T) {
	v := MustGet()

	spec, ok := v.GetEmulator("eden")
	if !ok {
		t.Fatal("eden emulator not found")
	}

	// Check that we have multiple versions
	versions := spec.AvailableVersions()
	if len(versions) < 2 {
		t.Errorf("expected at least 2 eden versions, got %d: %v", len(versions), versions)
	}

	// Check default is v0.1.1
	if spec.Default != "v0.1.1" {
		t.Errorf("expected default version v0.1.1, got %s", spec.Default)
	}

	// Check we can get specific versions
	v010 := spec.GetVersion("v0.1.0")
	if v010 == nil {
		t.Error("v0.1.0 not found")
	}
	v011 := spec.GetVersion("v0.1.1")
	if v011 == nil {
		t.Error("v0.1.1 not found")
	}

	// Check GetDefault returns v0.1.1
	defaultEntry := spec.GetDefault()
	if defaultEntry == nil || defaultEntry.Version != "v0.1.1" {
		t.Errorf("GetDefault() should return v0.1.1, got %v", defaultEntry)
	}
}

func TestGetEmulator(t *testing.T) {
	v := MustGet()

	// Test known emulators
	knownEmulators := []string{
		"eden", "duckstation", "pcsx2", "ppsspp", "mgba",
		"cemu", "azahar", "dolphin", "melonds", "vita3k",
		"rpcs3", "flycast", "retroarch",
	}

	for _, name := range knownEmulators {
		spec, ok := v.GetEmulator(name)
		if !ok {
			t.Errorf("GetEmulator(%s) returned false", name)
			continue
		}
		if spec.Default == "" {
			t.Errorf("GetEmulator(%s) has empty default", name)
		}
		if spec.URLTemplate == "" {
			t.Errorf("GetEmulator(%s) has empty url_template", name)
		}
		if len(spec.Versions) == 0 {
			t.Errorf("GetEmulator(%s) has no versions", name)
		}
	}

	// Test unknown emulator
	_, ok := v.GetEmulator("nonexistent")
	if ok {
		t.Error("GetEmulator(nonexistent) should return false")
	}
}

func TestGetCoreSize(t *testing.T) {
	v := MustGet()

	knownCores := []string{"bsnes", "mesen", "genesis_plus_gx", "mupen64plus_next", "mednafen_saturn"}
	for _, core := range knownCores {
		size := v.GetCoreSize(core)
		if size <= 0 {
			t.Errorf("GetCoreSize(%s) = %d, want > 0", core, size)
		}
	}

	unknownSize := v.GetCoreSize("nonexistent_core")
	if unknownSize != 0 {
		t.Errorf("GetCoreSize(nonexistent_core) = %d, want 0", unknownSize)
	}
}
