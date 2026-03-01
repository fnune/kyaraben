package model

import "testing"

type fakeResolver struct {
	configDir string
	dataDir   string
	homeDir   string
}

func (r fakeResolver) UserConfigDir() (string, error) { return r.configDir, nil }
func (r fakeResolver) UserDataDir() (string, error)   { return r.dataDir, nil }
func (r fakeResolver) UserHomeDir() (string, error)   { return r.homeDir, nil }

func TestConfigTargetResolve(t *testing.T) {
	resolver := fakeResolver{
		configDir: "/home/user/.config",
		dataDir:   "/home/user/.local/share",
		homeDir:   "/home/user",
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
			want: "/home/user/.config/duckstation/settings.ini",
		},
		{
			name: "user data dir",
			target: ConfigTarget{
				RelPath: "retroarch/retroarch.cfg",
				Format:  ConfigFormatCFG,
				BaseDir: ConfigBaseDirUserData,
			},
			want: "/home/user/.local/share/retroarch/retroarch.cfg",
		},
		{
			name: "home dir",
			target: ConfigTarget{
				RelPath: ".emulatorrc",
				Format:  ConfigFormatINI,
				BaseDir: ConfigBaseDirHome,
			},
			want: "/home/user/.emulatorrc",
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
			got, err := tt.target.ResolveWith(resolver)
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
	resolver := fakeResolver{
		configDir: "/home/user/.config",
		dataDir:   "/home/user/.local/share",
		homeDir:   "/home/user",
	}

	target := ConfigTarget{
		RelPath: "emulator/subdir/deep/config.ini",
		Format:  ConfigFormatINI,
		BaseDir: ConfigBaseDirUserConfig,
	}

	got, err := target.ResolveWith(resolver)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	want := "/home/user/.config/emulator/subdir/deep/config.ini"
	if got != want {
		t.Errorf("Resolve() = %q, want %q", got, want)
	}
}

func TestConfigTargetResolveDir(t *testing.T) {
	resolver := fakeResolver{
		configDir: "/home/user/.config",
		dataDir:   "/home/user/.local/share",
		homeDir:   "/home/user",
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
			want: "/home/user/.config/duckstation",
		},
		{
			name: "extracts top-level dir from nested path",
			target: ConfigTarget{
				RelPath: "retroarch/config/bsnes/bsnes.cfg",
				BaseDir: ConfigBaseDirUserConfig,
			},
			want: "/home/user/.config/retroarch",
		},
		{
			name: "extracts top-level dir from user data",
			target: ConfigTarget{
				RelPath: "ppsspp/PSP/SYSTEM/ppsspp.ini",
				BaseDir: ConfigBaseDirUserData,
			},
			want: "/home/user/.local/share/ppsspp",
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
			got, err := tt.target.ResolveDirWith(resolver)
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
