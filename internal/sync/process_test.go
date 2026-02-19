package sync

import (
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestFindPIDByPort_FindsListeningProcess(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	port := ln.Addr().(*net.TCPAddr).Port

	pid, err := FindPIDByPort(port)
	if err != nil {
		t.Fatalf("FindPIDByPort() error = %v", err)
	}

	if pid != os.Getpid() {
		t.Errorf("FindPIDByPort() = %d, want %d (current process)", pid, os.Getpid())
	}
}

func TestFindPIDByPort_ReturnsZeroForUnusedPort(t *testing.T) {
	pid, err := FindPIDByPort(59999)
	if err != nil {
		t.Fatalf("FindPIDByPort() error = %v", err)
	}

	if pid != 0 {
		t.Errorf("FindPIDByPort() = %d, want 0 for unused port", pid)
	}
}

func TestIsKyarabenSyncthing_MatchesConfigPath(t *testing.T) {
	stateDir := "/home/test/.local/state/kyaraben"
	cmdline := "syncthing\x00serve\x00--config=" + filepath.Join(stateDir, "syncthing", "config") + "\x00--data=/some/path"

	tmpFile := filepath.Join(t.TempDir(), "cmdline")
	if err := os.WriteFile(tmpFile, []byte(cmdline), 0644); err != nil {
		t.Fatalf("failed to write test cmdline: %v", err)
	}

	result := isKyarabenSyncthingFromCmdline(cmdline, stateDir)
	if !result {
		t.Error("isKyarabenSyncthingFromCmdline() = false, want true for matching config path")
	}
}

func TestIsKyarabenSyncthing_RejectsDifferentConfigPath(t *testing.T) {
	stateDir := "/home/test/.local/state/kyaraben"
	cmdline := "syncthing\x00serve\x00--config=/other/path/syncthing/config\x00--data=/some/path"

	result := isKyarabenSyncthingFromCmdline(cmdline, stateDir)
	if result {
		t.Error("isKyarabenSyncthingFromCmdline() = true, want false for different config path")
	}
}

func TestIsKyarabenSyncthing_RejectsNonSyncthing(t *testing.T) {
	stateDir := "/home/test/.local/state/kyaraben"
	cmdline := "python\x00server.py\x00--port=8080"

	result := isKyarabenSyncthingFromCmdline(cmdline, stateDir)
	if result {
		t.Error("isKyarabenSyncthingFromCmdline() = true, want false for non-syncthing process")
	}
}

func TestIsKyarabenInstance_MatchesKyarabenPath(t *testing.T) {
	cmdline := "syncthing\x00serve\x00--config=/home/user/.local/state/kyaraben/syncthing/config"

	result := isKyarabenInstanceFromCmdline(cmdline)
	if !result {
		t.Error("isKyarabenInstanceFromCmdline() = false, want true for kyaraben syncthing")
	}
}

func TestIsKyarabenInstance_MatchesKyarabenInstancePath(t *testing.T) {
	cmdline := "syncthing\x00serve\x00--config=/home/user/.local/state/kyaraben-deck/syncthing/config"

	result := isKyarabenInstanceFromCmdline(cmdline)
	if !result {
		t.Error("isKyarabenInstanceFromCmdline() = false, want true for kyaraben instance syncthing")
	}
}

func TestIsKyarabenInstance_RejectsNonKyaraben(t *testing.T) {
	cmdline := "syncthing\x00serve\x00--config=/home/user/.config/syncthing"

	result := isKyarabenInstanceFromCmdline(cmdline)
	if result {
		t.Error("isKyarabenInstanceFromCmdline() = true, want false for non-kyaraben syncthing")
	}
}

func TestKillProcess_TerminatesGracefully(t *testing.T) {
	cmd := exec.Command("sleep", "60")
	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start sleep process: %v", err)
	}

	pid := cmd.Process.Pid

	done := make(chan error, 1)
	go func() {
		done <- cmd.Wait()
	}()

	err := KillProcess(pid, 2*time.Second)
	if err != nil {
		t.Errorf("KillProcess() error = %v", err)
	}

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Error("process should be dead after KillProcess()")
		_ = cmd.Process.Kill()
	}
}

func TestKillProcess_HandlesAlreadyDead(t *testing.T) {
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to run true: %v", err)
	}

	err := KillProcess(cmd.Process.Pid, 1*time.Second)
	if err != nil {
		t.Errorf("KillProcess() on dead process error = %v", err)
	}
}

func TestWaitForPortRelease_ReturnsImmediatelyForFreePort(t *testing.T) {
	start := time.Now()
	err := WaitForPortRelease(59998, 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("WaitForPortRelease() error = %v", err)
	}

	if elapsed > 100*time.Millisecond {
		t.Errorf("WaitForPortRelease() took %v, should return immediately for free port", elapsed)
	}
}

func TestWaitForPortRelease_WaitsForRelease(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}

	port := ln.Addr().(*net.TCPAddr).Port

	go func() {
		time.Sleep(200 * time.Millisecond)
		_ = ln.Close()
	}()

	start := time.Now()
	err = WaitForPortRelease(port, 5*time.Second)
	elapsed := time.Since(start)

	if err != nil {
		t.Errorf("WaitForPortRelease() error = %v", err)
	}

	if elapsed < 150*time.Millisecond {
		t.Errorf("WaitForPortRelease() returned too quickly: %v", elapsed)
	}
}

func TestWaitForPortRelease_TimesOut(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	port := ln.Addr().(*net.TCPAddr).Port

	err = WaitForPortRelease(port, 100*time.Millisecond)
	if err == nil {
		t.Error("WaitForPortRelease() should timeout when port stays in use")
	}
}

func isKyarabenSyncthingFromCmdline(cmdline, stateDir string) bool {
	if !containsString(cmdline, "syncthing") {
		return false
	}

	configArg := "--config=" + filepath.Join(stateDir, "syncthing", "config")
	return containsString(cmdline, configArg)
}

func isKyarabenInstanceFromCmdline(cmdline string) bool {
	if !containsString(cmdline, "syncthing") {
		return false
	}
	return containsString(cmdline, "kyaraben")
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
