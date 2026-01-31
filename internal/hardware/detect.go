package hardware

import (
	"os"
	"runtime"
	"strings"
)

type Target struct {
	Name string
	Arch string
}

// DetectTarget returns the best Eden target for the current hardware.
func DetectTarget() Target {
	arch := runtime.GOARCH
	if arch == "arm64" {
		return Target{Name: "aarch64", Arch: "aarch64"}
	}

	// x86_64: try to detect specific hardware
	productName := readDMI("/sys/devices/virtual/dmi/id/product_name")

	switch {
	case strings.Contains(productName, "Jupiter"):
		return Target{Name: "steamdeck", Arch: "x86_64"}
	case strings.Contains(productName, "ROG Ally"):
		return Target{Name: "rog-ally", Arch: "x86_64"}
	default:
		return Target{Name: "amd64", Arch: "x86_64"}
	}
}

func readDMI(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
