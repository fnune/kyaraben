package nix

import (
	"testing"
)

func TestProgressParser_EvaluatingPhase(t *testing.T) {
	p := NewProgressParser()

	tests := []struct {
		input string
		want  BuildPhase
	}{
		{"evaluating derivation 'github:nixos/nixpkgs/...'", PhaseEvaluating},
		{"evaluating file '/nix/store/...'", PhaseEvaluating},
	}

	for _, tt := range tests {
		result := p.Parse(tt.input)
		if result == nil {
			t.Errorf("Parse(%q) = nil, want phase %s", tt.input, tt.want)
			continue
		}
		if result.Phase != tt.want {
			t.Errorf("Parse(%q).Phase = %s, want %s", tt.input, result.Phase, tt.want)
		}
	}
}

func TestProgressParser_WithExpectedPackages(t *testing.T) {
	p := NewProgressParser()
	p.SetExpectedPackages([]ExpectedPackage{
		{Name: "duckstation", DisplayName: "DuckStation", SizeBytes: 100 * 1024 * 1024},
		{Name: "dolphin", DisplayName: "Dolphin", SizeBytes: 50 * 1024 * 1024},
	})

	result := p.Parse("evaluating derivation...")
	if result == nil || result.Phase != PhaseEvaluating {
		t.Error("Expected evaluating phase")
	}

	result = p.Parse("building '/nix/store/abc-duckstation-0.1.drv'")
	if result == nil {
		t.Fatal("Expected result for known package")
	}
	if result.Phase != PhaseInstalling {
		t.Errorf("Phase = %s, want installing", result.Phase)
	}
	if result.PackageName != "DuckStation" {
		t.Errorf("PackageName = %q, want %q", result.PackageName, "DuckStation")
	}
	if result.ProgressPercent < 5 {
		t.Errorf("ProgressPercent = %d, expected > 5", result.ProgressPercent)
	}

	result = p.Parse("copying path '/nix/store/def-dolphin-5.0' from 'https://cache.nixos.org'")
	if result == nil {
		t.Fatal("Expected result for second package")
	}
	if result.PackageName != "Dolphin" {
		t.Errorf("PackageName = %q, want %q", result.PackageName, "Dolphin")
	}
	if result.ProgressPercent <= 50 {
		t.Errorf("ProgressPercent = %d, expected > 50", result.ProgressPercent)
	}
}

func TestProgressParser_NixPortableStorePaths(t *testing.T) {
	p := NewProgressParser()
	p.SetExpectedPackages([]ExpectedPackage{
		{Name: "es-de", DisplayName: "ES-DE", SizeBytes: 100 * 1024 * 1024},
	})

	result := p.Parse("building '~/.local/state/kyaraben/rhxla0prb23k4aqb6dzhwjvqfv9cc8pm-es-de-3.4.0.drv'...")
	if result == nil {
		t.Fatal("Expected result for nix-portable store path")
	}
	if result.PackageName != "ES-DE" {
		t.Errorf("PackageName = %q, want %q", result.PackageName, "ES-DE")
	}
}

func TestProgressParser_IgnoresUnknownPackages(t *testing.T) {
	p := NewProgressParser()
	p.SetExpectedPackages([]ExpectedPackage{
		{Name: "duckstation", DisplayName: "DuckStation", SizeBytes: 100 * 1024 * 1024},
	})

	result := p.Parse("building '/nix/store/abc-git-minimal-2.0.drv'")
	if result != nil {
		t.Error("Expected nil for unknown package")
	}

	result = p.Parse("building '/nix/store/abc-gcc-12.0.drv'")
	if result != nil {
		t.Error("Expected nil for unknown package")
	}
}

func TestProgressParser_SkipsDuplicatePackage(t *testing.T) {
	p := NewProgressParser()
	p.SetExpectedPackages([]ExpectedPackage{
		{Name: "duckstation", DisplayName: "DuckStation", SizeBytes: 100 * 1024 * 1024},
	})

	result1 := p.Parse("building '/nix/store/abc-duckstation-0.1.drv'")
	if result1 == nil {
		t.Fatal("First parse should return result")
	}

	result2 := p.Parse("building '/nix/store/def-duckstation-0.1.drv'")
	if result2 != nil {
		t.Error("Second parse of same package should return nil")
	}
}

func TestProgressParser_EmptyLine(t *testing.T) {
	p := NewProgressParser()
	result := p.Parse("")
	if result != nil {
		t.Error("Parse(\"\") should return nil")
	}

	result = p.Parse("   ")
	if result != nil {
		t.Error("Parse(whitespace) should return nil")
	}
}

func TestProgressParser_ProgressPercentage(t *testing.T) {
	p := NewProgressParser()
	p.SetExpectedPackages([]ExpectedPackage{
		{Name: "pkg1", DisplayName: "Pkg1", SizeBytes: 100},
		{Name: "pkg2", DisplayName: "Pkg2", SizeBytes: 100},
		{Name: "pkg3", DisplayName: "Pkg3", SizeBytes: 100},
		{Name: "pkg4", DisplayName: "Pkg4", SizeBytes: 100},
	})

	p.Parse("building '/nix/store/abc-pkg1.drv'")
	result := p.Parse("building '/nix/store/abc-pkg2.drv'")
	if result.ProgressPercent < 40 || result.ProgressPercent > 60 {
		t.Errorf("After 2/4 packages, expected ~50%%, got %d%%", result.ProgressPercent)
	}

	p.Parse("building '/nix/store/abc-pkg3.drv'")
	result = p.Parse("building '/nix/store/abc-pkg4.drv'")
	if result.ProgressPercent < 95 {
		t.Errorf("After all packages, expected ~100%%, got %d%%", result.ProgressPercent)
	}
}

func TestMatchesExpected(t *testing.T) {
	tests := []struct {
		nixPkg   string
		expected string
		want     bool
	}{
		{"duckstation-0.1", "duckstation", true},
		{"duckstation", "duckstation", true},
		{"DuckStation-0.1", "duckstation", true},
		{"libretro-bsnes-1.0", "libretro-bsnes", true},
		{"bsnes-libretro-1.0", "libretro-bsnes", true},
		{"git-minimal", "duckstation", false},
		{"gcc-12", "duckstation", false},
	}

	for _, tt := range tests {
		got := matchesExpected(tt.nixPkg, tt.expected)
		if got != tt.want {
			t.Errorf("matchesExpected(%q, %q) = %v, want %v", tt.nixPkg, tt.expected, got, tt.want)
		}
	}
}

func TestExtractPackageName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"duckstation", "duckstation"},
		{"es-de", "es-de"},
		{"retroarch-1.17.0", "retroarch-1.17.0"},
		{"2.0.0-package-name", "package-name"},
		{"", ""},
	}

	for _, tt := range tests {
		got := extractPackageName(tt.input)
		if got != tt.want {
			t.Errorf("extractPackageName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
