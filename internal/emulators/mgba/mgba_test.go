package mgba

import (
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestRomCommand(t *testing.T) {
	emu := Definition{}.Emulator()

	tests := []struct {
		name     string
		opts     model.RomLaunchOptions
		expected string
	}{
		{
			name: "basic command",
			opts: model.RomLaunchOptions{
				BinaryPath: "/usr/bin/mgba",
			},
			expected: "/usr/bin/mgba %ROM%",
		},
		{
			name: "fullscreen",
			opts: model.RomLaunchOptions{
				BinaryPath: "/usr/bin/mgba",
				Fullscreen: true,
			},
			expected: "/usr/bin/mgba -f %ROM%",
		},
		{
			name: "with saves dir",
			opts: model.RomLaunchOptions{
				BinaryPath: "/usr/bin/mgba",
				SavesDir:   "/home/user/saves/gb",
			},
			expected: "/usr/bin/mgba -C savegamePath=/home/user/saves/gb %ROM%",
		},
		{
			name: "fullscreen with saves dir",
			opts: model.RomLaunchOptions{
				BinaryPath: "/usr/bin/mgba",
				Fullscreen: true,
				SavesDir:   "/home/user/saves/gbc",
			},
			expected: "/usr/bin/mgba -f -C savegamePath=/home/user/saves/gbc %ROM%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := emu.Launcher.RomCommand(tt.opts)
			if got != tt.expected {
				t.Errorf("RomCommand() = %q, want %q", got, tt.expected)
			}
		})
	}
}
