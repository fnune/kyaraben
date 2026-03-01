package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/internal/importscanner"
)

type ImportCmd struct {
	Scan ImportScanCmd `cmd:"" help:"Scan existing collection and show what can be imported."`
}

type ImportScanCmd struct {
	Path   string `arg:"" required:"" help:"Path to existing collection (e.g. ~/Emulation)."`
	ESDE   string `name:"esde" help:"Path to ES-DE data directory (default: ~/ES-DE if it exists)."`
	Format string `enum:"text,json" default:"text" help:"Output format."`
}

func (cmd *ImportScanCmd) Run(ctx *Context) error {
	cfg, err := ctx.LoadConfig()
	if err != nil {
		return err
	}

	collection, err := ctx.NewCollection(cfg)
	if err != nil {
		return err
	}

	esdePath := cmd.ESDE
	if esdePath == "" {
		esdePath = detectESDE()
	}

	scanner := importscanner.NewScanner(ctx.FS, ctx.NewRegistry(), collection)
	report, err := scanner.Scan(importscanner.ScanOptions{
		SourcePath: cmd.Path,
		ESDEPath:   esdePath,
	})
	if err != nil {
		return err
	}

	switch cmd.Format {
	case "json":
		return outputJSON(report)
	default:
		return outputText(report)
	}
}

func detectESDE() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	esdePath := filepath.Join(home, "ES-DE")
	if info, err := os.Stat(esdePath); err == nil && info.IsDir() {
		return esdePath
	}
	return ""
}

func outputJSON(report *importscanner.ImportReport) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(report)
}

func outputText(report *importscanner.ImportReport) error {
	f := &textFormatter{report: report}
	f.print()
	return nil
}

type textFormatter struct {
	report *importscanner.ImportReport
}

func (f *textFormatter) print() {
	fmt.Println("Back up your collection before copying anything.")
	fmt.Println()

	if f.report.Mode == importscanner.ImportModeReorganize {
		fmt.Printf("Reorganizing: %s (same as Kyaraben collection)\n", f.report.SourcePath)
	} else {
		fmt.Printf("Existing collection: %s\n", f.report.SourcePath)
		fmt.Printf("Kyaraben collection: %s\n", f.report.KyarabenPath)
	}

	if f.report.ESDEPath != "" {
		fmt.Printf("ES-DE: %s\n", f.report.ESDEPath)
	}

	fmt.Println()
	fmt.Printf("Overall: %s in Kyaraben, %s to import\n",
		formatSizeDelta(f.report.Summary.TotalOnlyInKyaraben, true),
		formatSizeDelta(f.report.Summary.TotalOnlyInSource, false))
	fmt.Println()

	for _, sys := range f.report.Systems {
		f.printSystemSection(sys)
	}

	for _, fe := range f.report.Frontends {
		f.printFrontendSection(fe)
	}
}

func (f *textFormatter) printSystemSection(sys importscanner.SystemReport) {
	fmt.Println(strings.Repeat("=", 75))
	fmt.Println(sys.SystemName)
	fmt.Println(strings.Repeat("=", 75))
	fmt.Println()

	for _, data := range sys.SystemData {
		f.printDataComparison(data)
	}

	for _, emu := range sys.Emulators {
		fmt.Println(strings.Repeat("-", 75))
		fmt.Printf("%s (%s emulator)\n", emu.EmulatorName, sys.SystemName)
		fmt.Println(strings.Repeat("-", 75))
		fmt.Println()

		for _, data := range emu.EmulatorData {
			f.printDataComparison(data)
		}
	}
}

func (f *textFormatter) printFrontendSection(fe importscanner.FrontendReport) {
	fmt.Println(strings.Repeat("=", 75))
	fmt.Printf("%s (frontend)\n", fe.FrontendName)
	fmt.Println(strings.Repeat("=", 75))
	fmt.Println()

	for _, data := range fe.FrontendData {
		f.printDataComparison(data)
	}
}

func (f *textFormatter) printDataComparison(data importscanner.DataComparison) {
	label := dataTypeLabel(data.DataType)
	kyarabenDelta := formatSizeDelta(data.Diff.KyarabenDelta, true)
	sourceDelta := formatSizeDelta(data.Diff.SourceDelta, false)

	fmt.Printf("%-50s %s  %s\n", label, kyarabenDelta, sourceDelta)
	fmt.Println()

	foundLabel := "Found at:"
	expectLabel := "Kyaraben expects:"
	if f.report.Mode == importscanner.ImportModeReorganize {
		foundLabel = "Current:"
		expectLabel = "Expected:"
	}

	fmt.Printf("  %-16s %s\n", foundLabel, data.Source.Path)
	if data.Source.Exists {
		fmt.Printf("  %16s %d files, %s\n", "", data.Source.FileCount, formatSize(data.Source.TotalSize))
		if data.Source.Symlink != nil {
			status := "intact"
			if !data.Source.Symlink.Intact {
				status = "broken"
			}
			fmt.Printf("  %16s (symlink to %s, %s)\n", "", data.Source.Symlink.Target, status)
		}
	} else {
		fmt.Printf("  %16s (not found)\n", "")
	}

	fmt.Printf("  %-16s %s\n", expectLabel, data.Kyaraben.Path)
	if data.Kyaraben.Exists {
		fmt.Printf("  %16s %d files, %s\n", "", data.Kyaraben.FileCount, formatSize(data.Kyaraben.TotalSize))
	} else {
		fmt.Printf("  %16s (not found)\n", "")
	}

	for _, note := range data.Notes {
		fmt.Printf("\n  Note: %s\n", note)
	}

	if len(data.Diff.OnlyInSource) > 0 {
		fmt.Println()
		missingLabel := "Missing from Kyaraben"
		if f.report.Mode == importscanner.ImportModeReorganize {
			missingLabel = "Needs to move"
		}
		fmt.Printf("  %s (%d files, %s):\n", missingLabel, len(data.Diff.OnlyInSource), formatSize(data.Diff.SourceDelta))
		f.printFileList(data.Diff.OnlyInSource, 5)
	}

	if len(data.Diff.OnlyInKyaraben) > 0 {
		fmt.Println()
		extraLabel := "Only in Kyaraben"
		if f.report.Mode == importscanner.ImportModeReorganize {
			extraLabel = "Already in place"
		}
		fmt.Printf("  %s (%d files, %s):\n", extraLabel, len(data.Diff.OnlyInKyaraben), formatSize(data.Diff.KyarabenDelta))
		f.printFileList(data.Diff.OnlyInKyaraben, 5)
	}

	fmt.Println()
}

func (f *textFormatter) printFileList(files []importscanner.FileInfo, max int) {
	for i, file := range files {
		if i >= max {
			fmt.Printf("    ... and %d more\n", len(files)-max)
			break
		}
		fmt.Printf("    %s\n", file.RelPath)
	}
}

func dataTypeLabel(dt importscanner.DataType) string {
	switch dt {
	case importscanner.DataTypeROMs:
		return "ROMs"
	case importscanner.DataTypeBIOS:
		return "BIOS"
	case importscanner.DataTypeSaves:
		return "Saves"
	case importscanner.DataTypeStates:
		return "States"
	case importscanner.DataTypeScreenshots:
		return "Screenshots"
	case importscanner.DataTypeGamelists:
		return "Gamelists"
	case importscanner.DataTypeMedia:
		return "Scraped media"
	default:
		return string(dt)
	}
}

func formatSize(bytes int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)

	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.1f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.1f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func formatSizeDelta(bytes int64, positive bool) string {
	prefix := "-"
	if positive {
		prefix = "+"
	}
	if bytes == 0 {
		return prefix + "0"
	}
	return prefix + formatSize(bytes)
}
