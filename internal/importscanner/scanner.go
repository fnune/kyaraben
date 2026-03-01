package importscanner

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/model"
	"github.com/fnune/kyaraben/internal/registry"
	"github.com/fnune/kyaraben/internal/store"
)

var ignoredDirs = map[string]bool{
	".stfolder":    true,
	".stversions":  true,
	"shader_cache": true,
	"cache":        true,
}

type Scanner struct {
	fs         vfs.FS
	registry   *registry.Registry
	collection *store.Collection
	layout     SourceLayout
}

func NewScanner(fs vfs.FS, reg *registry.Registry, collection *store.Collection) *Scanner {
	return &Scanner{
		fs:         fs,
		registry:   reg,
		collection: collection,
		layout:     &GenericLayout{},
	}
}

type ScanOptions struct {
	SourcePath string
	ESDEPath   string
	Layout     string
}

func (s *Scanner) Scan(opts ScanOptions) (*ImportReport, error) {
	sourcePath, err := expandPath(opts.SourcePath)
	if err != nil {
		return nil, err
	}

	if _, err := s.fs.Stat(sourcePath); err != nil {
		return nil, fmt.Errorf("source path: %w", err)
	}

	if opts.Layout != "" {
		s.layout = GetLayout(opts.Layout)
	}

	kyarabenPath := s.collection.Root()
	mode := detectMode(sourcePath, kyarabenPath)

	report := &ImportReport{
		SourcePath:   sourcePath,
		ESDEPath:     opts.ESDEPath,
		KyarabenPath: kyarabenPath,
		Mode:         mode,
	}

	systemReports := make(map[model.SystemID]*SystemReport)
	emulatorReports := make(map[model.EmulatorID]*EmulatorReport)

	for _, sys := range s.registry.AllSystems() {
		systemReports[sys.ID] = &SystemReport{
			System:     sys.ID,
			SystemName: sys.Name,
		}
	}

	for _, emu := range s.registry.AllEmulators() {
		emulatorReports[emu.ID] = &EmulatorReport{
			Emulator:     emu.ID,
			EmulatorName: emu.Name,
		}
	}

	if err := s.scanSystemData(sourcePath, systemReports); err != nil {
		return nil, err
	}
	if err := s.scanEmulatorData(sourcePath, emulatorReports); err != nil {
		return nil, err
	}

	if opts.ESDEPath != "" {
		esdePath, err := expandPath(opts.ESDEPath)
		if err != nil {
			return nil, err
		}
		if err := s.scanESDE(esdePath, report); err != nil {
			return nil, err
		}
	}

	for _, sys := range s.registry.AllSystems() {
		sr := systemReports[sys.ID]
		for _, emu := range s.registry.GetEmulatorsForSystem(sys.ID) {
			if er, ok := emulatorReports[emu.ID]; ok && hasEmulatorData(er) {
				sr.Emulators = append(sr.Emulators, *er)
			}
		}

		if hasSystemData(sr) || len(sr.Emulators) > 0 {
			report.Systems = append(report.Systems, *sr)
		}
	}

	sortSystems(report.Systems)
	s.computeSummary(report)

	return report, nil
}

func (s *Scanner) scanSystemData(sourcePath string, reports map[model.SystemID]*SystemReport) error {
	sourceFound := make(map[model.SystemID]map[DataType]string)

	err := walkDir(s.fs, sourcePath, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || p == sourcePath {
			return nil
		}

		rel, _ := filepath.Rel(sourcePath, p)
		classification := s.layout.Classify(rel)
		if classification == nil || classification.System == nil {
			return nil
		}

		dt := classification.DataType
		if dt != DataTypeROMs && dt != DataTypeBIOS && dt != DataTypeSaves {
			return nil
		}

		sys := *classification.System
		if sourceFound[sys] == nil {
			sourceFound[sys] = make(map[DataType]string)
		}
		if _, exists := sourceFound[sys][dt]; !exists {
			sourceFound[sys][dt] = p
		}

		return filepath.SkipDir
	})
	if err != nil {
		return err
	}

	dataTypes := []struct {
		dt DataType
		fn func(model.SystemID) string
	}{
		{DataTypeROMs, s.collection.SystemRomsDir},
		{DataTypeBIOS, s.collection.SystemBiosDir},
		{DataTypeSaves, s.collection.SystemSavesDir},
	}

	allSystemsMap := make(map[model.SystemID]bool)
	for sys := range sourceFound {
		allSystemsMap[sys] = true
	}
	for _, sys := range s.registry.AllSystems() {
		allSystemsMap[sys.ID] = true
	}
	allSystems := make([]model.SystemID, 0, len(allSystemsMap))
	for sys := range allSystemsMap {
		allSystems = append(allSystems, sys)
	}
	sort.Slice(allSystems, func(i, j int) bool { return allSystems[i] < allSystems[j] })

	for _, sys := range allSystems {
		sr := reports[sys]
		if sr == nil {
			name := string(sys)
			if regSys, err := s.registry.GetSystem(sys); err == nil {
				name = regSys.Name
			}
			sr = &SystemReport{System: sys, SystemName: name}
			reports[sys] = sr
		}

		for _, dtInfo := range dataTypes {
			kyarabenDir := dtInfo.fn(sys)

			var srcDir string
			if sourceFound[sys] != nil && sourceFound[sys][dtInfo.dt] != "" {
				srcDir = sourceFound[sys][dtInfo.dt]
			} else {
				expectedPaths := s.layout.ExpectedPaths(dtInfo.dt)
				if len(expectedPaths) > 0 {
					srcDir = filepath.Join(sourcePath, expectedPaths[0], string(sys))
				} else {
					srcDir = kyarabenDir
				}
			}

			comparison, err := s.compareDirectories(srcDir, kyarabenDir, dtInfo.dt)
			if err != nil {
				return err
			}

			if comparison.Source.Exists || comparison.Kyaraben.Exists {
				sr.SystemData = append(sr.SystemData, comparison)
			}
		}
	}

	return nil
}

func (s *Scanner) scanEmulatorData(sourcePath string, reports map[model.EmulatorID]*EmulatorReport) error {
	sourceFound := make(map[model.EmulatorID]map[DataType]string)

	err := walkDir(s.fs, sourcePath, func(p string, d fs.DirEntry, err error) error {
		if err != nil || !d.IsDir() || p == sourcePath {
			return nil
		}

		rel, _ := filepath.Rel(sourcePath, p)
		classification := s.layout.Classify(rel)
		if classification == nil || classification.Emulator == nil {
			return nil
		}

		dt := classification.DataType
		if dt != DataTypeStates && dt != DataTypeScreenshots {
			return nil
		}

		emu := *classification.Emulator
		if sourceFound[emu] == nil {
			sourceFound[emu] = make(map[DataType]string)
		}
		if _, exists := sourceFound[emu][dt]; !exists {
			sourceFound[emu][dt] = p
		}

		return filepath.SkipDir
	})
	if err != nil {
		return err
	}

	emuDataTypes := []struct {
		dt DataType
		fn func(model.EmulatorID) string
	}{
		{DataTypeStates, s.collection.EmulatorStatesDir},
		{DataTypeScreenshots, s.collection.EmulatorScreenshotsDir},
	}

	allEmulatorsMap := make(map[model.EmulatorID]bool)
	for emu := range sourceFound {
		allEmulatorsMap[emu] = true
	}
	for _, emu := range s.registry.AllEmulators() {
		allEmulatorsMap[emu.ID] = true
	}
	allEmulators := make([]model.EmulatorID, 0, len(allEmulatorsMap))
	for emu := range allEmulatorsMap {
		allEmulators = append(allEmulators, emu)
	}
	sort.Slice(allEmulators, func(i, j int) bool { return allEmulators[i] < allEmulators[j] })

	for _, emu := range allEmulators {
		er := reports[emu]
		if er == nil {
			name := string(emu)
			if regEmu, err := s.registry.GetEmulator(emu); err == nil {
				name = regEmu.Name
			}
			er = &EmulatorReport{Emulator: emu, EmulatorName: name}
			reports[emu] = er
		}

		for _, dtInfo := range emuDataTypes {
			kyarabenDir := dtInfo.fn(emu)

			var srcDir string
			if sourceFound[emu] != nil && sourceFound[emu][dtInfo.dt] != "" {
				srcDir = sourceFound[emu][dtInfo.dt]
			} else {
				expectedPaths := s.layout.ExpectedPaths(dtInfo.dt)
				if len(expectedPaths) > 0 {
					srcDir = filepath.Join(sourcePath, expectedPaths[0], string(emu))
				} else {
					srcDir = kyarabenDir
				}
			}

			comparison, err := s.compareDirectories(srcDir, kyarabenDir, dtInfo.dt)
			if err != nil {
				return err
			}

			if comparison.Source.Exists || comparison.Kyaraben.Exists {
				er.EmulatorData = append(er.EmulatorData, comparison)
			}
		}
	}

	return nil
}

func (s *Scanner) scanESDE(esdePath string, report *ImportReport) error {
	frontendReport := FrontendReport{
		Frontend:     model.FrontendIDESDE,
		FrontendName: "ES-DE",
	}

	gamelistsSource := filepath.Join(esdePath, "gamelists")
	gamelistsKyaraben := filepath.Join(s.collection.Root(), "frontends", "esde", "gamelists")
	if comparison, err := s.compareDirectories(gamelistsSource, gamelistsKyaraben, DataTypeGamelists); err == nil {
		if comparison.Source.Exists || comparison.Kyaraben.Exists {
			frontendReport.FrontendData = append(frontendReport.FrontendData, comparison)
		}
	}

	mediaSource := filepath.Join(esdePath, "downloaded_media")
	mediaKyaraben := filepath.Join(s.collection.Root(), "frontends", "esde", "media")
	if comparison, err := s.compareDirectories(mediaSource, mediaKyaraben, DataTypeMedia); err == nil {
		if comparison.Source.Exists || comparison.Kyaraben.Exists {
			frontendReport.FrontendData = append(frontendReport.FrontendData, comparison)
		}
	}

	if len(frontendReport.FrontendData) > 0 {
		report.Frontends = append(report.Frontends, frontendReport)
	}

	return nil
}

func (s *Scanner) compareDirectories(srcPath, kyarabenPath string, dataType DataType) (DataComparison, error) {
	srcInfo, err := s.scanFolder(srcPath)
	if err != nil {
		return DataComparison{}, err
	}

	kyarabenInfo, err := s.scanFolder(kyarabenPath)
	if err != nil {
		return DataComparison{}, err
	}

	diff := s.computeDiff(srcPath, kyarabenPath, srcInfo, kyarabenInfo)

	comparison := DataComparison{
		DataType: dataType,
		Source:   srcInfo,
		Kyaraben: kyarabenInfo,
		Diff:     diff,
	}

	if srcInfo.IsFlat && !kyarabenInfo.IsFlat && kyarabenInfo.Exists {
		comparison.Notes = append(comparison.Notes, "Source is flat, Kyaraben expects subdirectories")
	}

	return comparison, nil
}

func (s *Scanner) scanFolder(path string) (FolderInfo, error) {
	info := FolderInfo{Path: path}

	stat, err := s.fs.Lstat(path)
	if os.IsNotExist(err) {
		return info, nil
	}
	if err != nil {
		return info, err
	}

	info.Exists = true

	if stat.Mode()&os.ModeSymlink != 0 {
		target, err := s.fs.Readlink(path)
		if err == nil {
			info.Symlink = &SymlinkInfo{Target: target}
			_, statErr := s.fs.Stat(path)
			info.Symlink.Intact = statErr == nil
		}
	}

	var fileCount int
	var totalSize int64
	hasSubdirs := false

	err = walkDir(s.fs, path, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if p == path {
			return nil
		}

		if d.IsDir() {
			rel, _ := filepath.Rel(path, p)
			if !strings.Contains(rel, string(filepath.Separator)) && !ignoredDirs[d.Name()] {
				hasSubdirs = true
			}
			return nil
		}

		fi, err := d.Info()
		if err != nil {
			return nil
		}
		fileCount++
		totalSize += fi.Size()
		return nil
	})
	if err != nil {
		return info, err
	}

	info.FileCount = fileCount
	info.TotalSize = totalSize
	info.IsFlat = !hasSubdirs && fileCount > 0

	return info, nil
}

func (s *Scanner) computeDiff(srcPath, kyarabenPath string, srcInfo, kyarabenInfo FolderInfo) DiffInfo {
	diff := DiffInfo{}

	if !srcInfo.Exists && !kyarabenInfo.Exists {
		return diff
	}

	srcFiles := make(map[string]int64)
	kyarabenFiles := make(map[string]int64)

	if srcInfo.Exists {
		_ = walkDir(s.fs, srcPath, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(srcPath, p)
			fi, _ := d.Info()
			if fi != nil {
				srcFiles[rel] = fi.Size()
			}
			return nil
		})
	}

	if kyarabenInfo.Exists {
		_ = walkDir(s.fs, kyarabenPath, func(p string, d fs.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			rel, _ := filepath.Rel(kyarabenPath, p)
			fi, _ := d.Info()
			if fi != nil {
				kyarabenFiles[rel] = fi.Size()
			}
			return nil
		})
	}

	for file, size := range srcFiles {
		if _, inKyaraben := kyarabenFiles[file]; !inKyaraben {
			diff.OnlyInSource = append(diff.OnlyInSource, FileInfo{RelPath: file, Size: size})
			diff.SourceDelta += size
		}
	}

	for file, size := range kyarabenFiles {
		if _, inSource := srcFiles[file]; !inSource {
			diff.OnlyInKyaraben = append(diff.OnlyInKyaraben, FileInfo{RelPath: file, Size: size})
			diff.KyarabenDelta += size
		}
	}

	sortFileInfos(diff.OnlyInSource)
	sortFileInfos(diff.OnlyInKyaraben)

	return diff
}

func (s *Scanner) computeSummary(report *ImportReport) {
	for _, sys := range report.Systems {
		for _, data := range sys.SystemData {
			report.Summary.TotalOnlyInSource += data.Diff.SourceDelta
			report.Summary.TotalOnlyInKyaraben += data.Diff.KyarabenDelta
		}
		for _, emu := range sys.Emulators {
			for _, data := range emu.EmulatorData {
				report.Summary.TotalOnlyInSource += data.Diff.SourceDelta
				report.Summary.TotalOnlyInKyaraben += data.Diff.KyarabenDelta
			}
		}
	}
	for _, fe := range report.Frontends {
		for _, data := range fe.FrontendData {
			report.Summary.TotalOnlyInSource += data.Diff.SourceDelta
			report.Summary.TotalOnlyInKyaraben += data.Diff.KyarabenDelta
		}
	}
}

func detectMode(sourcePath, kyarabenPath string) ImportMode {
	srcAbs, _ := filepath.Abs(sourcePath)
	kyarabenAbs, _ := filepath.Abs(kyarabenPath)

	if srcAbs == kyarabenAbs {
		return ImportModeReorganize
	}

	if strings.HasPrefix(srcAbs, kyarabenAbs+string(filepath.Separator)) ||
		strings.HasPrefix(kyarabenAbs, srcAbs+string(filepath.Separator)) {
		return ImportModeReorganize
	}

	return ImportModeCopy
}

func expandPath(path string) (string, error) {
	if len(path) > 0 && path[0] == '~' {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(home, path[1:])
	}
	return filepath.Abs(path)
}

func hasSystemData(sr *SystemReport) bool {
	for _, d := range sr.SystemData {
		if d.Source.Exists || d.Kyaraben.Exists {
			return true
		}
	}
	return false
}

func hasEmulatorData(er *EmulatorReport) bool {
	for _, d := range er.EmulatorData {
		if d.Source.Exists || d.Kyaraben.Exists {
			return true
		}
	}
	return false
}

func sortSystems(systems []SystemReport) {
	sort.Slice(systems, func(i, j int) bool {
		return systems[i].System < systems[j].System
	})
}

func sortFileInfos(files []FileInfo) {
	sort.Slice(files, func(i, j int) bool {
		return files[i].RelPath < files[j].RelPath
	})
}

func walkDir(fsys vfs.FS, root string, fn func(path string, d fs.DirEntry, err error) error) error {
	info, err := fsys.Lstat(root)
	if err != nil {
		return fn(root, nil, err)
	}

	return walkDirRecursive(fsys, root, fs.FileInfoToDirEntry(info), fn)
}

func walkDirRecursive(fsys vfs.FS, path string, d fs.DirEntry, fn func(string, fs.DirEntry, error) error) error {
	if err := fn(path, d, nil); err != nil {
		if err == filepath.SkipDir {
			return nil
		}
		return err
	}

	if !d.IsDir() {
		return nil
	}

	entries, err := fsys.ReadDir(path)
	if err != nil {
		return fn(path, d, err)
	}

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		if err := walkDirRecursive(fsys, entryPath, entry, fn); err != nil {
			if err == filepath.SkipDir {
				continue
			}
			return err
		}
	}

	return nil
}
