package sync

import (
	"net"
	"testing"

	"github.com/fnune/kyaraben/internal/model"
)

func TestCheckPortsAvailable_AllFree(t *testing.T) {
	cfg := model.SyncthingConfig{
		GUIPort:       18380,
		ListenPort:    22080,
		DiscoveryPort: 21080,
	}

	err := checkPortsAvailable(cfg)
	if err != nil {
		t.Errorf("checkPortsAvailable() error = %v, want nil", err)
	}
}

func TestCheckPortsAvailable_TCPPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:18381")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	cfg := model.SyncthingConfig{
		GUIPort:       18381,
		ListenPort:    22081,
		DiscoveryPort: 21081,
	}

	err = checkPortsAvailable(cfg)
	if err == nil {
		t.Error("checkPortsAvailable() should fail when GUI port in use")
	}
}

func TestCheckPortsAvailable_UDPPortInUse(t *testing.T) {
	conn, err := net.ListenPacket("udp", "127.0.0.1:21082")
	if err != nil {
		t.Fatalf("failed to listen UDP: %v", err)
	}
	defer func() { _ = conn.Close() }()

	cfg := model.SyncthingConfig{
		GUIPort:       18382,
		ListenPort:    22082,
		DiscoveryPort: 21082,
	}

	err = checkPortsAvailable(cfg)
	if err == nil {
		t.Error("checkPortsAvailable() should fail when discovery port in use")
	}
}

func TestCheckPortsAvailable_ListenPortInUse(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:22083")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	cfg := model.SyncthingConfig{
		GUIPort:       18383,
		ListenPort:    22083,
		DiscoveryPort: 21083,
	}

	err = checkPortsAvailable(cfg)
	if err == nil {
		t.Error("checkPortsAvailable() should fail when listen port in use")
	}
}

func TestCheckPorts_CanBeReplaced(t *testing.T) {
	original := CheckPorts
	defer func() { CheckPorts = original }()

	called := false
	CheckPorts = func(_ model.SyncthingConfig) error {
		called = true
		return nil
	}

	cfg := model.SyncthingConfig{GUIPort: 8385}
	_ = CheckPorts(cfg)

	if !called {
		t.Error("CheckPorts replacement was not called")
	}
}
