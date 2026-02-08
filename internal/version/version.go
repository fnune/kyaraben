package version

import "os"

var (
	Version = "dev"
)

func Get() string {
	if override := os.Getenv("KYARABEN_VERSION"); override != "" {
		return override
	}
	return Version
}
