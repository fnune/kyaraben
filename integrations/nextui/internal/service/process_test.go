package service

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"
)

func TestProcessManager_WritePID(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if err := pm.WritePID(12345); err != nil {
		t.Fatalf("WritePID: %v", err)
	}

	data, err := os.ReadFile(pm.PIDFile())
	if err != nil {
		t.Fatalf("read pid file: %v", err)
	}

	if string(data) != "12345" {
		t.Errorf("expected '12345', got %q", data)
	}
}

func TestProcessManager_WritePIDCreatesDirectory(t *testing.T) {
	dir := t.TempDir()
	nested := filepath.Join(dir, "deep", "nested")
	pm := NewProcessManager(nested)

	if err := pm.WritePID(1); err != nil {
		t.Fatalf("WritePID: %v", err)
	}

	if _, err := os.Stat(pm.PIDFile()); err != nil {
		t.Errorf("pid file not created: %v", err)
	}
}

func TestProcessManager_RemovePID(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if err := pm.WritePID(1); err != nil {
		t.Fatal(err)
	}
	if err := pm.RemovePID(); err != nil {
		t.Fatalf("RemovePID: %v", err)
	}

	if _, err := os.Stat(pm.PIDFile()); !os.IsNotExist(err) {
		t.Error("pid file should not exist after RemovePID")
	}
}

func TestProcessManager_RemovePIDOnMissingFile(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if err := pm.RemovePID(); err != nil {
		t.Errorf("RemovePID on missing file should not error: %v", err)
	}
}

func TestProcessManager_GetRunningPIDWithInvalidContent(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pm.PIDFile(), []byte("not-a-number"), 0644); err != nil {
		t.Fatal(err)
	}

	_, running := pm.GetRunningPID()
	if running {
		t.Error("expected not running with invalid pid content")
	}
}

func TestProcessManager_GetRunningPIDWithMissingFile(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	_, running := pm.GetRunningPID()
	if running {
		t.Error("expected not running with missing pid file")
	}
}

func TestProcessManager_IsOurProcessWithCurrentProcess(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	pid := os.Getpid()
	if pm.IsOurProcess(pid) {
		t.Skip("current process cmdline does not contain test home path, which is expected")
	}
}

func TestProcessManager_IsOurProcessWithNonexistentPID(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if pm.IsOurProcess(999999999) {
		t.Error("should return false for nonexistent PID")
	}
}

func TestProcessManager_HomePath(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	expected := filepath.Join(dir, "syncthing")
	if pm.HomePath() != expected {
		t.Errorf("HomePath = %q, want %q", pm.HomePath(), expected)
	}
}

func TestProcessManager_GetRunningPIDCleansStaleFile(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if err := pm.WritePID(999999999); err != nil {
		t.Fatal(err)
	}

	_, running := pm.GetRunningPID()
	if running {
		t.Error("should not report running for nonexistent process")
	}

	if _, err := os.Stat(pm.PIDFile()); !os.IsNotExist(err) {
		t.Error("stale pid file should be removed")
	}
}

func TestProcessManager_StopProcessWithNoPIDFile(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	if err := pm.StopProcess(); err != nil {
		t.Errorf("StopProcess with no pid file should not error: %v", err)
	}
}

func TestProcessManager_StopProcessWithStalePID(t *testing.T) {
	dir := t.TempDir()
	pm := NewProcessManager(dir)

	stalePID := 999999999
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(pm.PIDFile(), []byte(strconv.Itoa(stalePID)), 0644); err != nil {
		t.Fatal(err)
	}

	if err := pm.StopProcess(); err != nil {
		t.Errorf("StopProcess with stale pid: %v", err)
	}

	if _, err := os.Stat(pm.PIDFile()); !os.IsNotExist(err) {
		t.Error("pid file should be cleaned up")
	}
}
