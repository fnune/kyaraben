package syncthing

import (
	"net"
	"testing"
)

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
