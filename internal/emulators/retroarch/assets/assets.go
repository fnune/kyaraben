// Package assets embeds overlay and palette files for RetroArch cores.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed overlays/*.png overlays/*.cfg palettes/*.pal
var embedded embed.FS

// OverlayFS returns the embedded overlay filesystem.
func OverlayFS() fs.FS {
	sub, _ := fs.Sub(embedded, "overlays")
	return sub
}

// PaletteFS returns the embedded palette filesystem.
func PaletteFS() fs.FS {
	sub, _ := fs.Sub(embedded, "palettes")
	return sub
}

// OverlayFiles returns the list of overlay file names for a system.
func OverlayFiles(systemType string) (pngFile, cfgFile string) {
	switch systemType {
	case "gb":
		return "Perfect_GB-DMG.png", "Perfect_GB-DMG.cfg"
	case "gbc":
		return "Perfect_GBC.png", "Perfect_GBC.cfg"
	case "gba":
		return "Perfect_GBA.png", "Perfect_GBA.cfg"
	case "crt":
		return "Jeltron_CRT.png", "Jeltron_CRT.cfg"
	default:
		return "", ""
	}
}
