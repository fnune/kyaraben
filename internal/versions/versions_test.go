package versions

import "testing"

func TestURLTemplateExpansion(t *testing.T) {
	v := MustGet()

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
		url := v.Eden.URL(tt.target)
		if url != tt.expected {
			t.Errorf("URL(%s) mismatch:\ngot:  %s\nwant: %s", tt.target, url, tt.expected)
		}
	}
}

func TestDefaultTargetForArch(t *testing.T) {
	v := MustGet()

	x86Target := v.Eden.DefaultTargetForArch("x86_64")
	if x86Target == "" {
		t.Error("DefaultTargetForArch(x86_64) returned empty string")
	}
	if target := v.Eden.Target(x86Target); target == nil || target.Arch != "x86_64" {
		t.Errorf("DefaultTargetForArch(x86_64) = %s, which is not a valid x86_64 target", x86Target)
	}

	armTarget := v.Eden.DefaultTargetForArch("aarch64")
	if armTarget == "" {
		t.Error("DefaultTargetForArch(aarch64) returned empty string")
	}
	if target := v.Eden.Target(armTarget); target == nil || target.Arch != "aarch64" {
		t.Errorf("DefaultTargetForArch(aarch64) = %s, which is not a valid aarch64 target", armTarget)
	}
}

func TestTargetsForArch(t *testing.T) {
	v := MustGet()

	x86Targets := v.Eden.TargetsForArch("x86_64")
	if len(x86Targets) != 4 {
		t.Errorf("expected 4 x86_64 targets, got %d: %v", len(x86Targets), x86Targets)
	}

	armTargets := v.Eden.TargetsForArch("aarch64")
	if len(armTargets) != 1 {
		t.Errorf("expected 1 aarch64 target, got %d: %v", len(armTargets), armTargets)
	}
}

func TestArchiveType(t *testing.T) {
	v := MustGet()

	tests := []struct {
		name     string
		appimage *AppImageVersion
		target   string
		expected string
	}{
		{"eden is direct AppImage", &v.Eden, "amd64", ""},
		{"duckstation is direct AppImage", &v.DuckStation, "x64", ""},
		{"retroarch is 7z", &v.RetroArch, "x86_64", "7z"},
		{"tic80 is tar.gz", &v.TIC80, "x64", "tar.gz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appimage.ArchiveType(tt.target)
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
		appimage *AppImageVersion
		target   string
		expected string
	}{
		{"eden has no binary path (direct)", &v.Eden, "amd64", ""},
		{"retroarch x86_64 has per-target binary path", &v.RetroArch, "x86_64", "RetroArch-Linux-x86_64.AppImage"},
		{"retroarch aarch64 has per-target binary path", &v.RetroArch, "aarch64", "RetroArch-Linux-aarch64.AppImage"},
		{"tic80 has binary path", &v.TIC80, "x64", "bin/tic80"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.appimage.BinaryPathForTarget(tt.target)
			if got != tt.expected {
				t.Errorf("BinaryPathForTarget(%s) = %q, want %q", tt.target, got, tt.expected)
			}
		})
	}
}

func TestEffectiveReleaseTag(t *testing.T) {
	v := MustGet()

	// Eden has no release_tag, should use version
	if got := v.Eden.EffectiveReleaseTag(); got != v.Eden.Version {
		t.Errorf("Eden.EffectiveReleaseTag() = %q, want version %q", got, v.Eden.Version)
	}

	// Dolphin has a release_tag that differs from version
	if v.Dolphin.ReleaseTag != "" && v.Dolphin.ReleaseTag != v.Dolphin.Version {
		if got := v.Dolphin.EffectiveReleaseTag(); got != v.Dolphin.ReleaseTag {
			t.Errorf("Dolphin.EffectiveReleaseTag() = %q, want release_tag %q", got, v.Dolphin.ReleaseTag)
		}
	}
}

func TestURLWithReleaseTag(t *testing.T) {
	v := MustGet()

	// Dolphin URL should use release_tag substitution
	url := v.Dolphin.URL("x86_64")
	if v.Dolphin.ReleaseTag != "" {
		// URL should contain the release tag, not just the version
		expectedPart := v.Dolphin.ReleaseTag
		if !contains(url, expectedPart) {
			t.Errorf("Dolphin URL should contain release_tag %q, got: %s", expectedPart, url)
		}
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsAt(s, substr, 0))
}

func containsAt(s, substr string, start int) bool {
	for i := start; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
