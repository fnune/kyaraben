package syncthing

import (
	"context"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestVersioningConfig(t *testing.T) {
	got := versioningConfig(&FolderVersioning{
		Type:   "staggered",
		Params: map[string]string{"maxAge": "2592000"},
	})

	if got["type"] != "staggered" {
		t.Errorf("type = %v, want staggered", got["type"])
	}
	params, ok := got["params"].(map[string]string)
	if !ok || params["maxAge"] != "2592000" {
		t.Errorf("params = %v, want maxAge=2592000", got["params"])
	}
	for _, key := range []string{"cleanupIntervalS", "fsPath", "fsType"} {
		if _, ok := got[key]; !ok {
			t.Errorf("versioning config missing %q field Syncthing expects", key)
		}
	}
}

func TestAddFolders_Versioning(t *testing.T) {
	staggered := &FolderVersioning{Type: "staggered", Params: map[string]string{"maxAge": "2592000"}}

	existing := []map[string]any{
		{"id": "kyaraben-saves-snes", "path": "/c/saves/snes", "type": "sendreceive",
			"versioning": map[string]any{"type": "", "cleanupIntervalS": float64(3600)}},
		{"id": "kyaraben-roms-snes", "path": "/c/roms/snes", "type": "sendreceive",
			"versioning": map[string]any{"type": ""}},
	}
	defaults := map[string]any{
		"rescanIntervalS":  float64(3600),
		"fsWatcherEnabled": true,
		"versioning":       map[string]any{"type": ""},
	}

	var putBody []map[string]any
	mux := http.NewServeMux()
	mux.HandleFunc("/rest/system/status", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"myID": "THIS-DEVICE"})
	})
	mux.HandleFunc("/rest/config/defaults/folder", func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(defaults)
	})
	mux.HandleFunc("/rest/config/folders", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			_ = json.NewEncoder(w).Encode(existing)
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &putBody); err != nil {
				t.Errorf("decoding PUT body: %v", err)
			}
			w.WriteHeader(http.StatusOK)
		}
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	client := NewClient(Config{BaseURL: server.URL})
	client.SetAPIKey("key")

	requests := []FolderCreateRequest{
		{ID: "kyaraben-saves-snes", Label: "saves", Path: "/c/saves/snes", Type: "sendreceive", Versioning: staggered},
		{ID: "kyaraben-roms-snes", Label: "roms", Path: "/c/roms/snes", Type: "sendreceive"},
		{ID: "kyaraben-saves-psx", Label: "saves", Path: "/c/saves/psx", Type: "sendreceive", Versioning: staggered},
		{ID: "kyaraben-roms-psx", Label: "roms", Path: "/c/roms/psx", Type: "sendreceive"},
	}

	if err := client.AddFolders(context.Background(), requests); err != nil {
		t.Fatalf("AddFolders() error = %v", err)
	}

	byID := make(map[string]map[string]any)
	for _, f := range putBody {
		id, _ := f["id"].(string)
		byID[id] = f
	}
	if len(byID) != 4 {
		t.Fatalf("PUT wrote %d folders, want 4: %v", len(byID), putBody)
	}

	t.Run("existing saves folder is migrated to staggered versioning", func(t *testing.T) {
		f := byID["kyaraben-saves-snes"]
		if got := versioningType(f); got != "staggered" {
			t.Errorf("versioning type = %q, want staggered", got)
		}
		if got := versioningMaxAge(f); got != "2592000" {
			t.Errorf("maxAge = %q, want 2592000", got)
		}
		if got, _ := f["path"].(string); got != "/c/saves/snes" {
			t.Errorf("reconcile clobbered path = %q, want /c/saves/snes", got)
		}
	})

	t.Run("existing roms folder is left unversioned", func(t *testing.T) {
		if got := versioningType(byID["kyaraben-roms-snes"]); got != "" {
			t.Errorf("ROM versioning type = %q, want empty", got)
		}
	})

	t.Run("new saves folder is created with staggered versioning", func(t *testing.T) {
		f := byID["kyaraben-saves-psx"]
		if got := versioningType(f); got != "staggered" {
			t.Errorf("versioning type = %q, want staggered", got)
		}
		if got := versioningMaxAge(f); got != "2592000" {
			t.Errorf("maxAge = %q, want 2592000", got)
		}
	})

	t.Run("new roms folder is created without versioning", func(t *testing.T) {
		if got := versioningType(byID["kyaraben-roms-psx"]); got != "" {
			t.Errorf("ROM versioning type = %q, want empty", got)
		}
	})
}

func versioningType(folder map[string]any) string {
	v, ok := folder["versioning"].(map[string]any)
	if !ok {
		return ""
	}
	t, _ := v["type"].(string)
	return t
}

func versioningMaxAge(folder map[string]any) string {
	v, ok := folder["versioning"].(map[string]any)
	if !ok {
		return ""
	}
	params, ok := v["params"].(map[string]any)
	if !ok {
		return ""
	}
	maxAge, _ := params["maxAge"].(string)
	return maxAge
}

func TestExtractHost(t *testing.T) {
	tests := []struct {
		name     string
		addr     string
		expected string
	}{
		{"tcp URL", "tcp://192.168.1.100:22000", "192.168.1.100:22000"},
		{"quic URL", "quic://192.168.1.100:22000", "192.168.1.100:22000"},
		{"tcp URL with IPv6", "tcp://[::1]:22000", "[::1]:22000"},
		{"bare host:port", "192.168.1.100:22000", "192.168.1.100:22000"},
		{"bare IPv6", "[::1]:22000", "[::1]:22000"},
		{"no port", "192.168.1.100", ""},
		{"empty", "", ""},
		{"invalid URL", "tcp://[invalid", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractHost(tt.addr)
			if got != tt.expected {
				t.Errorf("extractHost(%q) = %q, want %q", tt.addr, got, tt.expected)
			}
		})
	}
}

func TestCheckPortReachable_Reachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer func() { _ = listener.Close() }()

	addr := listener.Addr().String()
	if !CheckPortReachable(addr) {
		t.Errorf("CheckPortReachable(%q) = false, want true", addr)
	}

	if !CheckPortReachable("tcp://" + addr) {
		t.Errorf("CheckPortReachable(%q) = false, want true", "tcp://"+addr)
	}
}

func TestCheckPortReachable_Unreachable(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	addr := listener.Addr().String()
	_ = listener.Close()

	if CheckPortReachable(addr) {
		t.Errorf("CheckPortReachable(%q) = true, want false (port closed)", addr)
	}
}

func TestCheckPortReachable_InvalidAddress(t *testing.T) {
	tests := []string{
		"",
		"not-an-address",
		"tcp://[invalid",
	}

	for _, addr := range tests {
		if CheckPortReachable(addr) {
			t.Errorf("CheckPortReachable(%q) = true, want false", addr)
		}
	}
}
