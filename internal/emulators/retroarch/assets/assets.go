// Package assets embeds overlay and palette files for RetroArch cores.
package assets

import (
	"embed"
	"io/fs"
)

//go:embed overlays/*.png overlays/*.cfg
var embedded embed.FS

// OverlayFS returns the embedded overlay filesystem.
func OverlayFS() fs.FS {
	sub, _ := fs.Sub(embedded, "overlays")
	return sub
}

// OverlayFiles returns the list of overlay file names for a system.
// These are 5x integer-scaled placeholder overlays for 1280x800 (Steam Deck).
func OverlayFiles(systemType string) (pngFile, cfgFile string) {
	switch systemType {
	case "gb":
		return "placeholder_gb_5x.png", "placeholder_gb_5x.cfg"
	case "gbc":
		return "placeholder_gbc_5x.png", "placeholder_gbc_5x.cfg"
	case "gba":
		return "placeholder_gba_5x.png", "placeholder_gba_5x.cfg"
	case "ngp":
		return "placeholder_ngp_5x.png", "placeholder_ngp_5x.cfg"
	default:
		return "", ""
	}
}
