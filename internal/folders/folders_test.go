package folders

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestID(t *testing.T) {
	tests := []struct {
		category Category
		subdir   string
		want     string
	}{
		{CategoryROMs, "snes", "kyaraben-roms-snes"},
		{CategorySaves, "gb", "kyaraben-saves-gb"},
		{CategoryStates, "duckstation", "kyaraben-states-duckstation"},
		{CategoryScreenshots, "retroarch", "kyaraben-screenshots-retroarch"},
	}

	for _, tt := range tests {
		got := ID(tt.category, tt.subdir)
		if got != tt.want {
			t.Errorf("ID(%q, %q) = %q, want %q", tt.category, tt.subdir, got, tt.want)
		}
	}
}

func TestFrontendID(t *testing.T) {
	tests := []struct {
		frontend model.FrontendID
		suffix   string
		want     string
	}{
		{"esde", "gamelists-snes", "kyaraben-frontends-esde-gamelists-snes"},
		{"esde", "media-psx", "kyaraben-frontends-esde-media-psx"},
	}

	for _, tt := range tests {
		got := FrontendID(tt.frontend, tt.suffix)
		if got != tt.want {
			t.Errorf("FrontendID(%q, %q) = %q, want %q", tt.frontend, tt.suffix, got, tt.want)
		}
	}
}

func TestCategorySubdirType(t *testing.T) {
	tests := []struct {
		category Category
		want     SubdirType
	}{
		{CategoryROMs, SubdirSystem},
		{CategoryBIOS, SubdirSystem},
		{CategorySaves, SubdirSystem},
		{CategoryStates, SubdirEmulator},
		{CategoryScreenshots, SubdirEmulator},
	}

	for _, tt := range tests {
		got := tt.category.SubdirType()
		if got != tt.want {
			t.Errorf("%q.SubdirType() = %v, want %v", tt.category, got, tt.want)
		}
	}
}

func TestCategoryVersioning(t *testing.T) {
	tests := []struct {
		category Category
		want     bool
	}{
		{CategoryROMs, false},
		{CategoryBIOS, false},
		{CategorySaves, true},
		{CategoryStates, true},
		{CategoryScreenshots, false},
	}

	for _, tt := range tests {
		got := tt.category.Versioning()
		if got != tt.want {
			t.Errorf("%q.Versioning() = %v, want %v", tt.category, got, tt.want)
		}
	}
}

func TestGenerateSpecs(t *testing.T) {
	input := HostInput{
		Systems: []model.SystemID{"snes", "psx"},
		Emulators: []EmulatorInfo{
			{ID: "retroarch:bsnes", UsesStatesDir: true, UsesScreenshotsDir: true},
			{ID: "duckstation", UsesStatesDir: true, UsesScreenshotsDir: true},
			{ID: "eden", UsesStatesDir: false, UsesScreenshotsDir: false},
		},
		Frontends: []model.FrontendID{"esde"},
		FrontendSuffixes: func(fe model.FrontendID, systems []model.SystemID) []string {
			var suffixes []string
			for _, sys := range systems {
				suffixes = append(suffixes, "gamelists-"+string(sys))
				suffixes = append(suffixes, "media-"+string(sys))
			}
			return suffixes
		},
	}

	specs := GenerateSpecs(input)

	wantIDs := map[string]bool{
		"kyaraben-roms-snes":                     true,
		"kyaraben-roms-psx":                      true,
		"kyaraben-bios-snes":                     true,
		"kyaraben-bios-psx":                      true,
		"kyaraben-saves-snes":                    true,
		"kyaraben-saves-psx":                     true,
		"kyaraben-states-retroarch:bsnes":        true,
		"kyaraben-states-duckstation":            true,
		"kyaraben-screenshots-retroarch:bsnes":   true,
		"kyaraben-screenshots-duckstation":       true,
		"kyaraben-frontends-esde-gamelists-snes": true,
		"kyaraben-frontends-esde-gamelists-psx":  true,
		"kyaraben-frontends-esde-media-snes":     true,
		"kyaraben-frontends-esde-media-psx":      true,
	}

	gotIDs := make(map[string]bool)
	for _, spec := range specs {
		gotIDs[spec.ID] = true
	}

	for id := range wantIDs {
		if !gotIDs[id] {
			t.Errorf("missing expected folder ID: %s", id)
		}
	}

	if gotIDs["kyaraben-states-eden"] {
		t.Error("should not have states folder for eden (UsesStatesDir=false)")
	}

	if gotIDs["kyaraben-screenshots-eden"] {
		t.Error("should not have screenshots folder for eden (UsesScreenshotsDir=false)")
	}
}

func TestGenerateSpecsVersioning(t *testing.T) {
	input := HostInput{
		Systems: []model.SystemID{"snes"},
		Emulators: []EmulatorInfo{
			{ID: "retroarch:bsnes", UsesStatesDir: true, UsesScreenshotsDir: true},
		},
	}

	specs := GenerateSpecs(input)

	for _, spec := range specs {
		switch spec.Category {
		case CategorySaves, CategoryStates:
			if !spec.Versioning {
				t.Errorf("%s should have versioning enabled", spec.ID)
			}
		case CategoryROMs, CategoryBIOS, CategoryScreenshots:
			if spec.Versioning {
				t.Errorf("%s should not have versioning enabled", spec.ID)
			}
		}
	}
}
