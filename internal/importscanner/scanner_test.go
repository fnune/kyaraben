package importscanner

import (
	"testing"

	"github.com/twpayne/go-vfs/v5/vfst"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/paths"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

func TestScannerBasic(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/source/roms/gamecube/game1.iso":        "game data",
		"/source/roms/gamecube/game2.iso":        "more game data",
		"/source/saves/gamecube/save1.gci":       "save data",
		"/source/states/dolphin/state1.sav":      "state data",
		"/source/screenshots/dolphin/screen.png": "screenshot",
		"/kyaraben/roms":                         &vfst.Dir{Perm: 0755},
		"/kyaraben/bios":                         &vfst.Dir{Perm: 0755},
		"/kyaraben/saves":                        &vfst.Dir{Perm: 0755},
		"/kyaraben/states":                       &vfst.Dir{Perm: 0755},
		"/kyaraben/screenshots":                  &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	reg := registry.NewDefault()
	collection, err := store.NewCollection(fs, paths.DefaultPaths(), "/kyaraben")
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(fs, reg, collection)
	report, err := scanner.Scan(ScanOptions{SourcePath: "/source"})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if report.Mode != ImportModeCopy {
		t.Errorf("expected ImportModeCopy, got %s", report.Mode)
	}

	if report.SourcePath != "/source" {
		t.Errorf("expected source path /source, got %s", report.SourcePath)
	}

	var gcReport *SystemReport
	for i := range report.Systems {
		if report.Systems[i].System == model.SystemIDGameCube {
			gcReport = &report.Systems[i]
			break
		}
	}

	if gcReport == nil {
		t.Fatal("expected GameCube system report")
	}

	var romsComparison, savesComparison *DataComparison
	for i := range gcReport.SystemData {
		switch gcReport.SystemData[i].DataType {
		case DataTypeROMs:
			romsComparison = &gcReport.SystemData[i]
		case DataTypeSaves:
			savesComparison = &gcReport.SystemData[i]
		}
	}

	if romsComparison == nil {
		t.Fatal("expected ROMs comparison")
	}
	if romsComparison.Source.FileCount != 2 {
		t.Errorf("expected 2 ROM files, got %d", romsComparison.Source.FileCount)
	}
	if len(romsComparison.Diff.OnlyInSource) != 2 {
		t.Errorf("expected 2 files only in source, got %d", len(romsComparison.Diff.OnlyInSource))
	}

	if savesComparison == nil {
		t.Fatal("expected Saves comparison")
	}
	if savesComparison.Source.FileCount != 1 {
		t.Errorf("expected 1 save file, got %d", savesComparison.Source.FileCount)
	}
}

func TestScannerReorganizeMode(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/kyaraben/roms/gamecube/game1.iso": "game data",
		"/kyaraben/bios":                    &vfst.Dir{Perm: 0755},
		"/kyaraben/saves":                   &vfst.Dir{Perm: 0755},
		"/kyaraben/states":                  &vfst.Dir{Perm: 0755},
		"/kyaraben/screenshots":             &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	reg := registry.NewDefault()
	collection, err := store.NewCollection(fs, paths.DefaultPaths(), "/kyaraben")
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(fs, reg, collection)
	report, err := scanner.Scan(ScanOptions{SourcePath: "/kyaraben"})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if report.Mode != ImportModeReorganize {
		t.Errorf("expected ImportModeReorganize, got %s", report.Mode)
	}
}

func TestScannerDiff(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/source/roms/gamecube/game1.iso":   "game data",
		"/source/roms/gamecube/game2.iso":   "more data",
		"/kyaraben/roms/gamecube/game1.iso": "game data",
		"/kyaraben/roms/gamecube/game3.iso": "different",
		"/kyaraben/bios":                    &vfst.Dir{Perm: 0755},
		"/kyaraben/saves":                   &vfst.Dir{Perm: 0755},
		"/kyaraben/states":                  &vfst.Dir{Perm: 0755},
		"/kyaraben/screenshots":             &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	reg := registry.NewDefault()
	collection, err := store.NewCollection(fs, paths.DefaultPaths(), "/kyaraben")
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(fs, reg, collection)
	report, err := scanner.Scan(ScanOptions{SourcePath: "/source"})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	var gcReport *SystemReport
	for i := range report.Systems {
		if report.Systems[i].System == model.SystemIDGameCube {
			gcReport = &report.Systems[i]
			break
		}
	}

	if gcReport == nil {
		t.Fatal("expected GameCube system report")
	}

	var romsComparison *DataComparison
	for i := range gcReport.SystemData {
		if gcReport.SystemData[i].DataType == DataTypeROMs {
			romsComparison = &gcReport.SystemData[i]
			break
		}
	}

	if romsComparison == nil {
		t.Fatal("expected ROMs comparison")
	}

	if len(romsComparison.Diff.OnlyInSource) != 1 {
		t.Errorf("expected 1 file only in source (game2.iso), got %d", len(romsComparison.Diff.OnlyInSource))
	}
	if len(romsComparison.Diff.OnlyInKyaraben) != 1 {
		t.Errorf("expected 1 file only in kyaraben (game3.iso), got %d", len(romsComparison.Diff.OnlyInKyaraben))
	}
}

func TestScannerEmptyDirectories(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/source/roms":          &vfst.Dir{Perm: 0755},
		"/kyaraben/roms":        &vfst.Dir{Perm: 0755},
		"/kyaraben/bios":        &vfst.Dir{Perm: 0755},
		"/kyaraben/saves":       &vfst.Dir{Perm: 0755},
		"/kyaraben/states":      &vfst.Dir{Perm: 0755},
		"/kyaraben/screenshots": &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	reg := registry.NewDefault()
	collection, err := store.NewCollection(fs, paths.DefaultPaths(), "/kyaraben")
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(fs, reg, collection)
	report, err := scanner.Scan(ScanOptions{SourcePath: "/source"})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if report.Summary.TotalOnlyInSource != 0 {
		t.Errorf("expected 0 bytes to import, got %d", report.Summary.TotalOnlyInSource)
	}
	if report.Summary.TotalOnlyInKyaraben != 0 {
		t.Errorf("expected 0 bytes in kyaraben, got %d", report.Summary.TotalOnlyInKyaraben)
	}
}

func TestScannerESDE(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/source/roms":                          &vfst.Dir{Perm: 0755},
		"/esde/gamelists/snes/gamelist.xml":     "<xml>gamelist</xml>",
		"/esde/downloaded_media/snes/cover.png": "cover",
		"/kyaraben/roms":                        &vfst.Dir{Perm: 0755},
		"/kyaraben/bios":                        &vfst.Dir{Perm: 0755},
		"/kyaraben/saves":                       &vfst.Dir{Perm: 0755},
		"/kyaraben/states":                      &vfst.Dir{Perm: 0755},
		"/kyaraben/screenshots":                 &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	reg := registry.NewDefault()
	collection, err := store.NewCollection(fs, paths.DefaultPaths(), "/kyaraben")
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(fs, reg, collection)
	report, err := scanner.Scan(ScanOptions{
		SourcePath: "/source",
		ESDEPath:   "/esde",
	})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	if len(report.Frontends) != 1 {
		t.Fatalf("expected 1 frontend report, got %d", len(report.Frontends))
	}

	esdeReport := report.Frontends[0]
	if esdeReport.Frontend != model.FrontendIDESDE {
		t.Errorf("expected ES-DE frontend, got %s", esdeReport.Frontend)
	}

	if len(esdeReport.FrontendData) != 2 {
		t.Errorf("expected 2 frontend data comparisons (gamelists, media), got %d", len(esdeReport.FrontendData))
	}
}

func TestScannerSymlink(t *testing.T) {
	fs, cleanup, err := vfst.NewTestFS(map[string]any{
		"/source/roms/gamecube/game.iso": "game",
		"/kyaraben/roms":                 &vfst.Dir{Perm: 0755},
		"/kyaraben/bios":                 &vfst.Dir{Perm: 0755},
		"/kyaraben/saves":                &vfst.Dir{Perm: 0755},
		"/kyaraben/states":               &vfst.Dir{Perm: 0755},
		"/kyaraben/screenshots":          &vfst.Dir{Perm: 0755},
		"/actual/roms/gamecube":          &vfst.Dir{Perm: 0755},
	})
	if err != nil {
		t.Fatal(err)
	}
	defer cleanup()

	if err := fs.Symlink("/actual/roms/gamecube", "/kyaraben/roms/gamecube"); err != nil {
		t.Fatal(err)
	}

	reg := registry.NewDefault()
	collection, err := store.NewCollection(fs, paths.DefaultPaths(), "/kyaraben")
	if err != nil {
		t.Fatal(err)
	}

	scanner := NewScanner(fs, reg, collection)
	report, err := scanner.Scan(ScanOptions{SourcePath: "/source"})
	if err != nil {
		t.Fatalf("Scan() error = %v", err)
	}

	var gcReport *SystemReport
	for i := range report.Systems {
		if report.Systems[i].System == model.SystemIDGameCube {
			gcReport = &report.Systems[i]
			break
		}
	}

	if gcReport == nil {
		t.Fatal("expected GameCube system report")
	}

	var romsComparison *DataComparison
	for i := range gcReport.SystemData {
		if gcReport.SystemData[i].DataType == DataTypeROMs {
			romsComparison = &gcReport.SystemData[i]
			break
		}
	}

	if romsComparison == nil {
		t.Fatal("expected ROMs comparison")
	}

	if romsComparison.Kyaraben.Symlink == nil {
		t.Error("expected kyaraben ROMs path to be detected as symlink")
	} else if !romsComparison.Kyaraben.Symlink.Intact {
		t.Error("expected symlink to be intact")
	}
}
