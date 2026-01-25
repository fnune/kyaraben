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

	if got := v.Eden.DefaultTargetForArch("x86_64"); got != "amd64" {
		t.Errorf("DefaultTargetForArch(x86_64) = %s, want amd64", got)
	}
	if got := v.Eden.DefaultTargetForArch("aarch64"); got != "aarch64" {
		t.Errorf("DefaultTargetForArch(aarch64) = %s, want aarch64", got)
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
