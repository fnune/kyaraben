package sync

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/fnune/kyaraben/internal/model"
)

type testInstance struct {
	t          *testing.T
	name       string
	configDir  string
	dataDir    string
	syncDir    string
	guiPort    int
	listenPort int
	apiKey     string
	deviceID   string
	cmd        *exec.Cmd
	client     *Client
}

func newTestInstance(t *testing.T, name string, guiPort, listenPort int) *testInstance {
	t.Helper()

	tmpDir := t.TempDir()

	inst := &testInstance{
		t:          t,
		name:       name,
		configDir:  filepath.Join(tmpDir, "config"),
		dataDir:    filepath.Join(tmpDir, "data"),
		syncDir:    filepath.Join(tmpDir, "sync"),
		guiPort:    guiPort,
		listenPort: listenPort,
		apiKey:     fmt.Sprintf("test-api-key-%s", name),
	}

	for _, dir := range []string{inst.configDir, inst.dataDir, inst.syncDir} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("creating dir %s: %v", dir, err)
		}
	}

	return inst
}

func (inst *testInstance) writeConfigXML(cfg *SyncthingXMLConfig) error {
	configPath := filepath.Join(inst.configDir, "config.xml")

	data, err := xml.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	xmlHeader := []byte(xml.Header)
	data = append(xmlHeader, data...)

	return os.WriteFile(configPath, data, 0600)
}

func (inst *testInstance) generate() error {
	syncthingPath, err := exec.LookPath("syncthing")
	if err != nil {
		return fmt.Errorf("syncthing not in PATH: %w", err)
	}

	cmd := exec.Command(syncthingPath,
		"generate",
		fmt.Sprintf("--home=%s", inst.configDir),
	)
	cmd.Env = append(os.Environ(), "STNODEFAULTFOLDER=1")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("generating config: %w\noutput: %s", err, output)
	}

	return inst.readDeviceID()
}

func (inst *testInstance) readDeviceID() error {
	syncthingPath, err := exec.LookPath("syncthing")
	if err != nil {
		return err
	}

	cmd := exec.Command(syncthingPath,
		"device-id",
		fmt.Sprintf("--home=%s", inst.configDir),
	)

	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("reading device ID: %w", err)
	}

	inst.deviceID = string(output[:len(output)-1])
	return nil
}

func (inst *testInstance) start(ctx context.Context) error {
	syncthingPath, err := exec.LookPath("syncthing")
	if err != nil {
		return fmt.Errorf("syncthing not in PATH: %w", err)
	}

	inst.cmd = exec.CommandContext(ctx, syncthingPath,
		"serve",
		"--no-browser",
		fmt.Sprintf("--home=%s", inst.configDir),
		fmt.Sprintf("--gui-address=127.0.0.1:%d", inst.guiPort),
		fmt.Sprintf("--gui-apikey=%s", inst.apiKey),
	)

	inst.cmd.Env = append(os.Environ(),
		"STNODEFAULTFOLDER=1",
		"STNOUPGRADE=1",
	)

	if err := inst.cmd.Start(); err != nil {
		return fmt.Errorf("starting syncthing: %w", err)
	}

	inst.client = NewClient(model.SyncConfig{
		Syncthing: model.SyncthingConfig{GUIPort: inst.guiPort},
	})
	inst.client.SetAPIKey(inst.apiKey)

	return inst.waitReady(ctx)
}

func (inst *testInstance) waitReady(ctx context.Context) error {
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		if inst.client.IsRunning(ctx) {
			deviceID, err := inst.client.GetDeviceID(ctx)
			if err == nil && deviceID != "" {
				inst.deviceID = deviceID
				return nil
			}
		}
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
	return fmt.Errorf("syncthing did not become ready")
}

func (inst *testInstance) stop() {
	if inst.cmd != nil && inst.cmd.Process != nil {
		_ = inst.cmd.Process.Signal(os.Interrupt)
		done := make(chan error, 1)
		go func() { done <- inst.cmd.Wait() }()
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			_ = inst.cmd.Process.Kill()
			<-done
		}
		inst.cmd = nil
	}
}
