// Package sdlmappings supplies SDL game controller mappings so RetroArch's
// bundled SDL2 numbers pad buttons in the order kyaraben's bindings assume.
// See scripts/update-sdl-mappings.sh for which mappings are included and why.
package sdlmappings

import (
	"bufio"
	_ "embed"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
)

//go:embed gamecontrollerdb.txt
var database string

// Linux caps a single environment variable at 128 KiB (MAX_ARG_STRLEN).
const Budget = 120 * 1024

// Vendors whose pads people most often use, as their IDs appear in SDL GUIDs
// (byte-swapped).
var popularVendors = []string{
	"4c05", // Sony
	"5e04", // Microsoft
	"7e05", // Nintendo
	"c82d", // 8BitDo
	"6d04", // Logitech
	"3735", // GameSir
	"d620", // PowerA
	"6f0e", // PDP
	"0d0f", // Hori
	"3215", // Razer
	"3807", // Mad Catz
}

// ForDevices returns mappings for SDL_GAMECONTROLLERCONFIG, one per line:
// first the joysticks listed in a /proc/bus/input/devices file, then pads
// from popular vendors so ones plugged in later are covered too, up to
// budget bytes.
func ForDevices(devices io.Reader, budget int) (string, error) {
	guids, err := joystickGUIDs(devices)
	if err != nil {
		return "", err
	}

	var out strings.Builder
	included := map[string]bool{}
	add := func(line string) {
		if included[line[:32]] || out.Len()+len(line)+1 > budget {
			return
		}
		included[line[:32]] = true
		if out.Len() > 0 {
			out.WriteByte('\n')
		}
		out.WriteString(line)
	}
	for _, line := range lookup(guids) {
		add(line)
	}
	for _, line := range byPopularity() {
		add(line)
	}
	return out.String(), nil
}

func entries() []string {
	var lines []string
	for _, line := range strings.Split(database, "\n") {
		if len(line) >= 33 && !strings.HasPrefix(line, "#") {
			lines = append(lines, line)
		}
	}
	return lines
}

func byPopularity() []string {
	rank := func(line string) int {
		for i, v := range popularVendors {
			if line[8:12] == v {
				return i
			}
		}
		return len(popularVendors)
	}
	lines := entries()
	sort.SliceStable(lines, func(i, j int) bool { return rank(lines[i]) < rank(lines[j]) })
	return lines
}

func lookup(guids []string) []string {
	connected := make(map[string]bool, len(guids))
	versionless := make(map[string][]string, len(guids))
	for _, g := range guids {
		connected[g] = true
		key := withoutVersion(g)
		versionless[key] = append(versionless[key], g)
	}

	exact := map[string]string{}
	fallback := map[string]string{}
	for _, line := range entries() {
		guid, rest := line[:32], line[32:]
		if connected[guid] {
			exact[guid] = line
		}
		// SDL2 matches a mapping only on the exact GUID, so an entry written
		// for any version is rewritten onto each connected version.
		if guid[24:28] == "0000" {
			for _, g := range versionless[guid] {
				fallback[g] = g + rest
			}
		}
	}

	var out []string
	for _, g := range guids {
		if line, ok := exact[g]; ok {
			out = append(out, line)
		} else if line, ok := fallback[g]; ok {
			out = append(out, line)
		}
	}
	return out
}

// joystickGUIDs builds SDL2's Linux GUIDs (bus, vendor, product and version
// as little-endian 16-bit words, each padded with zeros) for every device
// with a joystick handler.
func joystickGUIDs(devices io.Reader) ([]string, error) {
	var guids []string
	var id string
	seen := map[string]bool{}
	scanner := bufio.NewScanner(devices)
	for scanner.Scan() {
		line := scanner.Text()
		switch {
		case strings.HasPrefix(line, "I: "):
			id = guidFromIDLine(line)
		case strings.HasPrefix(line, "H: Handlers="):
			if id != "" && hasJoystickHandler(line) && !seen[id] {
				seen[id] = true
				guids = append(guids, id)
			}
		case line == "":
			id = ""
		}
	}
	return guids, scanner.Err()
}

func guidFromIDLine(line string) string {
	fields := map[string]string{}
	for _, f := range strings.Fields(strings.TrimPrefix(line, "I: ")) {
		if k, v, ok := strings.Cut(f, "="); ok {
			fields[k] = v
		}
	}
	var b strings.Builder
	for _, key := range []string{"Bus", "Vendor", "Product", "Version"} {
		word, err := strconv.ParseUint(fields[key], 16, 16)
		if err != nil {
			return ""
		}
		fmt.Fprintf(&b, "%02x%02x0000", word&0xff, word>>8)
	}
	return b.String()
}

func hasJoystickHandler(line string) bool {
	for _, h := range strings.Fields(strings.TrimPrefix(line, "H: Handlers=")) {
		if strings.HasPrefix(h, "js") {
			return true
		}
	}
	return false
}

func withoutVersion(guid string) string {
	return guid[:24] + "0000" + guid[28:]
}
