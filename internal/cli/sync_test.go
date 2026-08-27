package cli

import (
	"testing"

	"github.com/fnune/kyaraben/internal/syncmeta"
)

func TestOnOff(t *testing.T) {
	if got := onOff(true); got != "on" {
		t.Errorf("onOff(true) = %q, want on", got)
	}
	if got := onOff(false); got != "off" {
		t.Errorf("onOff(false) = %q, want off", got)
	}
}

func TestRomDeletionStatus(t *testing.T) {
	markers := map[string]syncmeta.DeviceStatus{
		"PROTECTED": {DeviceID: "PROTECTED", IgnoreDeleteROMs: true},
		"FOLLOWS":   {DeviceID: "FOLLOWS", IgnoreDeleteROMs: false},
	}

	cases := []struct {
		deviceID string
		want     string
	}{
		{"PROTECTED", "Ignores ROM deletions from other devices"},
		{"FOLLOWS", "Follows ROM deletions from other devices"},
		{"UNREPORTED", "ROM deletion setting not yet reported"},
	}

	for _, tc := range cases {
		if got := romDeletionStatus(markers, tc.deviceID); got != tc.want {
			t.Errorf("romDeletionStatus(%q) = %q, want %q", tc.deviceID, got, tc.want)
		}
	}
}
