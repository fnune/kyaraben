// Package paths provides XDG Base Directory paths for kyaraben.
//
// Kyaraben stores all its operational data in XDG_STATE_HOME (~/.local/state/kyaraben/):
//
//	~/.local/state/kyaraben/
//	├── bin/              # wrapper scripts for emulator binaries (add to PATH)
//	├── cores/            # symlinks to RetroArch cores (libretro_directory)
//	├── current           # symlink to active nix profile (real store path)
//	├── current-gc-root   # nix-managed symlink (indirect GC root, virtual store path)
//	├── syncthing/        # sync device pairings and identity (user data)
//	├── kyaraben.log      # application log
//	└── build/            # regenerable via 'kyaraben apply'
//	    ├── nix/          # nix-portable store (multi-GB)
//	    ├── flake/        # generated flake.nix
//	    └── manifest.json # tracks installed emulators and configs
//
// We use STATE rather than DATA because this data is:
//   - Machine-specific (nix store paths contain hashes that won't work elsewhere)
//   - Mostly regenerable (kyaraben apply rebuilds build/ from config)
//   - Not user data (the actual user data is in ~/Emulation and ~/.config/<emulator>)
//
// User configuration lives in XDG_CONFIG_HOME (~/.config/kyaraben/config.toml).
// User emulation data (ROMs, saves) lives in a user-chosen location (default ~/Emulation).
// Emulator configs live in their standard locations (~/.config/<emulator>/).
package paths

import (
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
