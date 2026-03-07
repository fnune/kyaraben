package folders

import (
	"fmt"

	"github.com/fnune/kyaraben/internal/model"
)

type Category string

const (
	CategoryROMs        Category = "roms"
	CategoryBIOS        Category = "bios"
	CategorySaves       Category = "saves"
	CategoryStates      Category = "states"
	CategoryScreenshots Category = "screenshots"
)

type SubdirType int

const (
	SubdirNone SubdirType = iota
	SubdirSystem
	SubdirEmulator
)

func (c Category) SubdirType() SubdirType {
	switch c {
	case CategoryROMs, CategoryBIOS, CategorySaves:
		return SubdirSystem
	case CategoryStates, CategoryScreenshots:
		return SubdirEmulator
	default:
		return SubdirNone
	}
}

func (c Category) Versioning() bool {
	switch c {
	case CategorySaves, CategoryStates:
		return true
	default:
		return false
	}
}

func ID(category Category, subdir string) string {
	if subdir == "" {
		return fmt.Sprintf("kyaraben-%s", category)
	}
	return fmt.Sprintf("kyaraben-%s-%s", category, subdir)
}

func FrontendID(frontend model.FrontendID, suffix string) string {
	return fmt.Sprintf("kyaraben-frontends-%s-%s", frontend, suffix)
}

type Spec struct {
	ID         string
	Category   Category
	Versioning bool
	System     model.SystemID
	Emulator   model.EmulatorID
	Frontend   model.FrontendID
}

type HostInput struct {
	Systems          []model.SystemID
	Emulators        []EmulatorInfo
	Frontends        []model.FrontendID
	FrontendSuffixes func(model.FrontendID, []model.SystemID) []string
}

type EmulatorInfo struct {
	ID                 model.EmulatorID
	UsesStatesDir      bool
	UsesScreenshotsDir bool
}

func GenerateSpecs(input HostInput) []Spec {
	var specs []Spec

	for _, category := range []Category{CategoryROMs, CategoryBIOS, CategorySaves} {
		for _, sys := range input.Systems {
			specs = append(specs, Spec{
				ID:         ID(category, string(sys)),
				Category:   category,
				Versioning: category.Versioning(),
				System:     sys,
			})
		}
	}

	for _, emu := range input.Emulators {
		if emu.UsesStatesDir {
			specs = append(specs, Spec{
				ID:         ID(CategoryStates, string(emu.ID)),
				Category:   CategoryStates,
				Versioning: CategoryStates.Versioning(),
				Emulator:   emu.ID,
			})
		}
		if emu.UsesScreenshotsDir {
			specs = append(specs, Spec{
				ID:         ID(CategoryScreenshots, string(emu.ID)),
				Category:   CategoryScreenshots,
				Versioning: CategoryScreenshots.Versioning(),
				Emulator:   emu.ID,
			})
		}
	}

	for _, fe := range input.Frontends {
		if input.FrontendSuffixes != nil {
			for _, suffix := range input.FrontendSuffixes(fe, input.Systems) {
				specs = append(specs, Spec{
					ID:         FrontendID(fe, suffix),
					Category:   "", // Frontend folders don't use standard categories
					Versioning: false,
					Frontend:   fe,
				})
			}
		}
	}

	return specs
}
