package service

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAutostartManager_EnableCreatesAutoSh(t *testing.T) {
	dir := t.TempDir()
	mgr := NewAutostartManager(dir, "tg5040", "/mnt/SDCARD/Tools/tg5040/Kyaraben.pak", dir+"/logs")

	if err := mgr.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "auto.sh"))
	if err != nil {
		t.Fatalf("read auto.sh: %v", err)
	}

	if !strings.Contains(string(content), markerStart) {
		t.Error("auto.sh missing start marker")
	}
	if !strings.Contains(string(content), markerEnd) {
		t.Error("auto.sh missing end marker")
	}
	if !strings.Contains(string(content), "kyaraben_start_syncthing") {
		t.Error("auto.sh missing startup function")
	}
}

func TestAutostartManager_EnableIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	mgr := NewAutostartManager(dir, "tg5040", "/pak", dir+"/logs")

	if err := mgr.Enable(); err != nil {
		t.Fatalf("first Enable: %v", err)
	}
	if err := mgr.Enable(); err != nil {
		t.Fatalf("second Enable: %v", err)
	}

	content, err := os.ReadFile(filepath.Join(dir, "auto.sh"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}

	count := strings.Count(string(content), markerStart)
	if count != 1 {
		t.Errorf("expected 1 marker block, got %d", count)
	}
}

func TestAutostartManager_EnablePreservesExistingContent(t *testing.T) {
	dir := t.TempDir()
	autoSh := filepath.Join(dir, "auto.sh")

	existing := "#!/bin/sh\necho 'user script'\n"
	if err := os.WriteFile(autoSh, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}

	mgr := NewAutostartManager(dir, "tg5040", "/pak", dir+"/logs")
	if err := mgr.Enable(); err != nil {
		t.Fatalf("Enable: %v", err)
	}

	content, _ := os.ReadFile(autoSh)
	if !strings.Contains(string(content), "echo 'user script'") {
		t.Error("existing content was overwritten")
	}
	if !strings.Contains(string(content), markerStart) {
		t.Error("kyaraben block not added")
	}
}

func TestAutostartManager_DisableRemovesBlock(t *testing.T) {
	dir := t.TempDir()
	mgr := NewAutostartManager(dir, "tg5040", "/pak", dir+"/logs")

	if err := mgr.Enable(); err != nil {
		t.Fatal(err)
	}
	if !mgr.IsEnabled() {
		t.Fatal("expected IsEnabled=true after Enable")
	}

	if err := mgr.Disable(); err != nil {
		t.Fatal(err)
	}
	if mgr.IsEnabled() {
		t.Error("expected IsEnabled=false after Disable")
	}

	content, _ := os.ReadFile(filepath.Join(dir, "auto.sh"))
	if strings.Contains(string(content), markerStart) {
		t.Error("marker still present after Disable")
	}
}

func TestAutostartManager_DisablePreservesOtherContent(t *testing.T) {
	dir := t.TempDir()
	autoSh := filepath.Join(dir, "auto.sh")

	existing := "#!/bin/sh\necho 'before'\n"
	if err := os.WriteFile(autoSh, []byte(existing), 0755); err != nil {
		t.Fatal(err)
	}

	mgr := NewAutostartManager(dir, "tg5040", "/pak", dir+"/logs")
	if err := mgr.Enable(); err != nil {
		t.Fatal(err)
	}
	if err := mgr.Disable(); err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(autoSh)
	if !strings.Contains(string(content), "echo 'before'") {
		t.Error("existing content was removed")
	}
}

func TestAutostartManager_DisableOnMissingFile(t *testing.T) {
	dir := t.TempDir()
	mgr := NewAutostartManager(dir, "tg5040", "/pak", dir+"/logs")

	if err := mgr.Disable(); err != nil {
		t.Errorf("Disable on missing file should not error: %v", err)
	}
}

func TestAutostartManager_GeneratedBlockContainsPaths(t *testing.T) {
	dir := t.TempDir()
	pakPath := "/mnt/SDCARD/Tools/tg5040/Kyaraben.pak"
	logsPath := "/mnt/SDCARD/.userdata/tg5040/logs"
	mgr := NewAutostartManager(dir, "tg5040", pakPath, logsPath)

	if err := mgr.Enable(); err != nil {
		t.Fatal(err)
	}

	content, _ := os.ReadFile(filepath.Join(dir, "auto.sh"))
	s := string(content)

	if !strings.Contains(s, pakPath+"/syncthing") {
		t.Error("missing syncthing path")
	}
	if !strings.Contains(s, filepath.Join(dir, "kyaraben", "syncthing.pid")) {
		t.Error("missing pid file path")
	}
	if !strings.Contains(s, filepath.Join(dir, "kyaraben", "syncthing")) {
		t.Error("missing home path")
	}
	if !strings.Contains(s, filepath.Join(logsPath, "kyaraben-syncthing.log")) {
		t.Error("missing log file path")
	}
}
