// Package paths provides XDG Base Directory paths for kyaraben.
//
// Kyaraben stores all its operational data in XDG_STATE_HOME (~/.local/state/kyaraben/):
//
//	~/.local/state/kyaraben/
//	├── bin/              # wrapper scripts for emulator binaries (add to PATH)
//	├── cores/            # symlinks to RetroArch cores (libretro_directory)
//	├── syncthing/        # sync device pairings and identity (user data)
//	├── kyaraben.log      # application log
//	└── build/            # regenerable via 'kyaraben apply'
//	    └── manifest.json # tracks installed emulators and configs
//
// We use STATE rather than DATA because this data is:
//   - Machine-specific (installed package paths are local to the machine)
//   - Mostly regenerable (kyaraben apply rebuilds build/ from config)
//   - Not user data (the actual user data is in ~/Emulation and ~/.config/<emulator>)
//
// User configuration lives in XDG_CONFIG_HOME (~/.config/kyaraben/config.toml).
// User emulation data (ROMs, saves) lives in a user-chosen location (default ~/Emulation).
// Emulator configs live in their standard locations (~/.config/<emulator>/).
package paths

import (
	"hash/fnv"
	"os"
	"path/filepath"
)

func StateDir() (string, error) {
	if dir := os.Getenv("XDG_STATE_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "state"), nil
}

func ConfigDir() (string, error) {
	if dir := os.Getenv("XDG_CONFIG_HOME"); dir != "" {
		return dir, nil
	}
	return os.UserConfigDir()
}

func DataDir() (string, error) {
	if dir := os.Getenv("XDG_DATA_HOME"); dir != "" {
		return dir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func KyarabenStateDir() (string, error) {
	base, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "kyaraben"), nil
}

type Paths struct {
	Instance string
}

func NewPaths(instance string) *Paths {
	return &Paths{Instance: instance}
}

func DefaultPaths() *Paths {
	return NewPaths("")
}

func (p *Paths) DirName() string {
	if p.Instance != "" {
		return "kyaraben-" + p.Instance
	}
	return "kyaraben"
}

func (p *Paths) StateDir() (string, error) {
	base, err := StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, p.DirName()), nil
}

func (p *Paths) ConfigDir() (string, error) {
	base, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, p.DirName()), nil
}

func (p *Paths) ConfigPath() (string, error) {
	dir, err := p.ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.toml"), nil
}

func (p *Paths) ManifestPath() (string, error) {
	state, err := p.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(state, "build", "manifest.json"), nil
}

func (p *Paths) DesktopFileName() string {
	return p.DirName() + ".desktop"
}

func (p *Paths) AppBinaryName() string {
	if p.Instance != "" {
		return "kyaraben-ui-" + p.Instance
	}
	return "kyaraben-ui"
}

func (p *Paths) CLIBinaryName() string {
	return p.DirName()
}

func (p *Paths) CLIInstallPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "bin", p.CLIBinaryName()), nil
}

func (p *Paths) CoresDir() (string, error) {
	state, err := p.StateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(state, "cores"), nil
}

func (p *Paths) InstancePortOffset() int {
	if p.Instance == "" {
		return 0
	}
	h := fnv.New32a()
	h.Write([]byte(p.Instance))
	return int(h.Sum32()%100) + 1
}

func KyarabenConfigDir() (string, error) {
	base, err := ConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(base, "kyaraben"), nil
}

func RetroArchCoresDir() (string, error) {
	stateDir, err := KyarabenStateDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(stateDir, "cores"), nil
}

func MustRetroArchCoresDir() string {
	dir, err := RetroArchCoresDir()
	if err != nil {
		panic("cannot determine retroarch cores directory: " + err.Error())
	}
	return dir
}
