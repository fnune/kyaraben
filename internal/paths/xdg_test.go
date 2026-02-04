package paths

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStateDir(t *testing.T) {
	t.Run("returns XDG_STATE_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", "/custom/state")

		dir, err := StateDir()
		if err != nil {
			t.Fatalf("StateDir() error = %v", err)
		}
		if dir != "/custom/state" {
			t.Errorf("StateDir() = %q, want %q", dir, "/custom/state")
		}
	})

	t.Run("returns ~/.local/state when XDG_STATE_HOME not set", func(t *testing.T) {
		t.Setenv("XDG_STATE_HOME", "")

		dir, err := StateDir()
		if err != nil {
			t.Fatalf("StateDir() error = %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".local", "state")
		if dir != expected {
			t.Errorf("StateDir() = %q, want %q", dir, expected)
		}
	})
}

func TestConfigDir(t *testing.T) {
	t.Run("returns XDG_CONFIG_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "/custom/config")

		dir, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		if dir != "/custom/config" {
			t.Errorf("ConfigDir() = %q, want %q", dir, "/custom/config")
		}
	})

	t.Run("falls back to os.UserConfigDir when not set", func(t *testing.T) {
		t.Setenv("XDG_CONFIG_HOME", "")

		dir, err := ConfigDir()
		if err != nil {
			t.Fatalf("ConfigDir() error = %v", err)
		}
		expected, _ := os.UserConfigDir()
		if dir != expected {
			t.Errorf("ConfigDir() = %q, want %q", dir, expected)
		}
	})
}

func TestDataDir(t *testing.T) {
	t.Run("returns XDG_DATA_HOME when set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "/custom/data")

		dir, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		if dir != "/custom/data" {
			t.Errorf("DataDir() = %q, want %q", dir, "/custom/data")
		}
	})

	t.Run("returns ~/.local/share when not set", func(t *testing.T) {
		t.Setenv("XDG_DATA_HOME", "")

		dir, err := DataDir()
		if err != nil {
			t.Fatalf("DataDir() error = %v", err)
		}
		home, _ := os.UserHomeDir()
		expected := filepath.Join(home, ".local", "share")
		if dir != expected {
			t.Errorf("DataDir() = %q, want %q", dir, expected)
		}
	})
}

func TestKyarabenStateDir(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/test/state")

	dir, err := KyarabenStateDir()
	if err != nil {
		t.Fatalf("KyarabenStateDir() error = %v", err)
	}
	expected := "/test/state/kyaraben"
	if dir != expected {
		t.Errorf("KyarabenStateDir() = %q, want %q", dir, expected)
	}
}

func TestKyarabenConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/test/config")

	dir, err := KyarabenConfigDir()
	if err != nil {
		t.Fatalf("KyarabenConfigDir() error = %v", err)
	}
	expected := "/test/config/kyaraben"
	if dir != expected {
		t.Errorf("KyarabenConfigDir() = %q, want %q", dir, expected)
	}
}

func TestRetroArchCoresDir(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/test/state")

	dir, err := RetroArchCoresDir()
	if err != nil {
		t.Fatalf("RetroArchCoresDir() error = %v", err)
	}
	if !strings.HasPrefix(dir, "/test/state/kyaraben") {
		t.Errorf("RetroArchCoresDir() = %q, should be under kyaraben state dir", dir)
	}
	if !strings.HasSuffix(dir, filepath.Join("lib", "retroarch", "cores")) {
		t.Errorf("RetroArchCoresDir() = %q, should end with lib/retroarch/cores", dir)
	}
}

func TestMustRetroArchCoresDir(t *testing.T) {
	t.Setenv("XDG_STATE_HOME", "/test/state")

	dir := MustRetroArchCoresDir()
	expected := "/test/state/kyaraben/current/lib/retroarch/cores"
	if dir != expected {
		t.Errorf("MustRetroArchCoresDir() = %q, want %q", dir, expected)
	}
}
