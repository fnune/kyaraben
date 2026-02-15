package model

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigTargetResolve(t *testing.T) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("failed to get user config dir: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name    string
		target  ConfigTarget
		want    string
		wantErr bool
	}{
		{
			name: "user config dir",
			target: ConfigTarget{
				RelPath: "duckstation/settings.ini",
				Format:  ConfigFormatINI,
				BaseDir: ConfigBaseDirUserConfig,
			},
			want: filepath.Join(configDir, "duckstation", "settings.ini"),
		},
		{
			name: "user data dir",
			target: ConfigTarget{
				RelPath: "retroarch/retroarch.cfg",
				Format:  ConfigFormatCFG,
				BaseDir: ConfigBaseDirUserData,
			},
			want: filepath.Join(homeDir, ".local", "share", "retroarch", "retroarch.cfg"),
		},
		{
			name: "home dir",
			target: ConfigTarget{
				RelPath: ".emulatorrc",
				Format:  ConfigFormatINI,
				BaseDir: ConfigBaseDirHome,
			},
			want: filepath.Join(homeDir, ".emulatorrc"),
		},
		{
			name: "unknown base dir",
			target: ConfigTarget{
				RelPath: "test",
				Format:  ConfigFormatINI,
				BaseDir: ConfigBaseDir("invalid"),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.target.Resolve()
			if (err != nil) != tt.wantErr {
				t.Errorf("Resolve() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("Resolve() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConfigTargetResolveNested(t *testing.T) {
	configDir, _ := os.UserConfigDir()

	target := ConfigTarget{
		RelPath: "emulator/subdir/deep/config.ini",
		Format:  ConfigFormatINI,
		BaseDir: ConfigBaseDirUserConfig,
	}

	got, err := target.Resolve()
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	want := filepath.Join(configDir, "emulator", "subdir", "deep", "config.ini")
	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestConfigTargetResolveDir(t *testing.T) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		t.Fatalf("failed to get user config dir: %v", err)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Fatalf("failed to get home dir: %v", err)
	}

	tests := []struct {
		name    string
		target  ConfigTarget
		want    string
		wantErr bool
	}{
		{
			name: "extracts top-level dir from user config",
			target: ConfigTarget{
				RelPath: "duckstation/settings.ini",
				BaseDir: ConfigBaseDirUserConfig,
			},
			want: filepath.Join(configDir, "duckstation"),
		},
		{
			name: "extracts top-level dir from nested path",
			target: ConfigTarget{
				RelPath: "retroarch/config/bsnes/bsnes.cfg",
				BaseDir: ConfigBaseDirUserConfig,
			},
			want: filepath.Join(configDir, "retroarch"),
		},
		{
			name: "extracts top-level dir from user data",
			target: ConfigTarget{
				RelPath: "ppsspp/PSP/SYSTEM/ppsspp.ini",
				BaseDir: ConfigBaseDirUserData,
			},
			want: filepath.Join(homeDir, ".local", "share", "ppsspp"),
		},
		{
			name: "rejects single-component path for safety",
			target: ConfigTarget{
				RelPath: ".emulatorrc",
				BaseDir: ConfigBaseDirHome,
			},
			wantErr: true,
		},
		{
			name: "rejects empty RelPath",
			target: ConfigTarget{
				RelPath: "",
				BaseDir: ConfigBaseDirUserConfig,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.target.ResolveDir()
			if (err != nil) != tt.wantErr {
				t.Errorf("ResolveDir() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ResolveDir() = %q, want %q", got, tt.want)
			}
		})
	}
}
