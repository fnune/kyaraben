package collections

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestRetroArchCommanderWaitsForBind(t *testing.T) {
	procNetUDP := filepath.Join(t.TempDir(), "udp")
	header := "  sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode ref pointer drops\n"
	if err := os.WriteFile(procNetUDP, []byte(header), 0o644); err != nil {
		t.Fatal(err)
	}
	c := &RetroArchCommander{ProcNetUDP: procNetUDP}
	if err := c.Send(47420, "MENU_TOGGLE", 300*time.Millisecond); err == nil {
		t.Error("expected a timeout while the port is not bound")
	}

	bound := header + "  1: 00000000:B93C 00000000:0000 07 00000000:00000000 00:00000000 00000000  1000        0 1 2 0000000000000000 0\n"
	if err := os.WriteFile(procNetUDP, []byte(bound), 0o644); err != nil {
		t.Fatal(err)
	}
	if !c.bound(47420) {
		t.Error("port 47420 (0xB93C) should count as bound")
	}
}
