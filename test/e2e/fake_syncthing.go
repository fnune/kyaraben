package e2e

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
)

type FakeSyncthing struct {
	mu       sync.Mutex
	server   *http.Server
	listener net.Listener
	port     int

	deviceID    string
	devices     map[string]fakeDevice
	folders     map[string]fakeFolder
	connections map[string]fakeConnection
}

type fakeDevice struct {
	DeviceID    string   `json:"deviceID"`
	Name        string   `json:"name"`
	Addresses   []string `json:"addresses"`
	Compression string   `json:"compression"`
}

type fakeFolder struct {
	ID      string             `json:"id"`
	Path    string             `json:"path"`
	Type    string             `json:"type"`
	Devices []fakeFolderDevice `json:"devices"`
}

type fakeFolderDevice struct {
	DeviceID string `json:"deviceID"`
}

type fakeConnection struct {
	Connected bool   `json:"connected"`
	Address   string `json:"address"`
	Paused    bool   `json:"paused"`
}

func NewFakeSyncthing(deviceID string) *FakeSyncthing {
	return &FakeSyncthing{
		deviceID:    deviceID,
		devices:     make(map[string]fakeDevice),
		folders:     make(map[string]fakeFolder),
		connections: make(map[string]fakeConnection),
	}
}

func (f *FakeSyncthing) SetDeviceID(id string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.deviceID = id
}

func (f *FakeSyncthing) AddFolder(id, path, folderType string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.folders[id] = fakeFolder{
		ID:      id,
		Path:    path,
		Type:    folderType,
		Devices: []fakeFolderDevice{{DeviceID: f.deviceID}},
	}
}

func (f *FakeSyncthing) Start() error {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return err
	}
	f.listener = listener
	f.port = listener.Addr().(*net.TCPAddr).Port

	mux := http.NewServeMux()
	mux.HandleFunc("/rest/system/ping", f.handlePing)
	mux.HandleFunc("/rest/system/status", f.handleStatus)
	mux.HandleFunc("/rest/system/connections", f.handleConnections)
	mux.HandleFunc("/rest/config/devices/", f.handleDevice)
	mux.HandleFunc("/rest/config/devices", f.handleDevices)
	mux.HandleFunc("/rest/config/folders/", f.handleFolder)
	mux.HandleFunc("/rest/config/folders", f.handleFolders)
	mux.HandleFunc("/rest/db/status", f.handleFolderStatus)
	mux.HandleFunc("/rest/cluster/pending/devices", f.handlePendingDevices)
	mux.HandleFunc("/rest/cluster/pending/folders", f.handlePendingFolders)
	mux.HandleFunc("/rest/system/discovery", f.handleDiscovery)

	f.server = &http.Server{Handler: mux}
	go func() { _ = f.server.Serve(listener) }()
	return nil
}

func (f *FakeSyncthing) Stop() {
	if f.server != nil {
		_ = f.server.Close()
	}
}

func (f *FakeSyncthing) Port() int {
	return f.port
}

func (f *FakeSyncthing) BaseURL() string {
	return fmt.Sprintf("http://127.0.0.1:%d", f.port)
}

func (f *FakeSyncthing) handlePing(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{"ping":"pong"}`))
}

func (f *FakeSyncthing) handleStatus(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"myID":   f.deviceID,
		"uptime": 60,
	})
}

func (f *FakeSyncthing) handleConnections(w http.ResponseWriter, _ *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"connections": f.connections,
	})
}

func (f *FakeSyncthing) handleDevices(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.Method == http.MethodGet {
		devices := []fakeDevice{{
			DeviceID:    f.deviceID,
			Name:        "self",
			Addresses:   []string{"dynamic"},
			Compression: "metadata",
		}}
		for _, d := range f.devices {
			devices = append(devices, d)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(devices)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (f *FakeSyncthing) handleDevice(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	deviceID := strings.TrimPrefix(r.URL.Path, "/rest/config/devices/")

	switch r.Method {
	case http.MethodPut:
		var dev fakeDevice
		if err := json.NewDecoder(r.Body).Decode(&dev); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.devices[deviceID] = dev
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))

	case http.MethodDelete:
		delete(f.devices, deviceID)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))

	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (f *FakeSyncthing) handleFolders(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	if r.Method == http.MethodGet {
		folders := make([]fakeFolder, 0, len(f.folders))
		for _, folder := range f.folders {
			folders = append(folders, folder)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(folders)
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (f *FakeSyncthing) handleFolder(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()

	folderID := strings.TrimPrefix(r.URL.Path, "/rest/config/folders/")

	if r.Method == http.MethodPut {
		var folder fakeFolder
		if err := json.NewDecoder(r.Body).Decode(&folder); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		f.folders[folderID] = folder
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
		return
	}

	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

func (f *FakeSyncthing) handleFolderStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{
		"state":       "idle",
		"globalBytes": 0,
		"needBytes":   0,
	})
}

func (f *FakeSyncthing) handlePendingDevices(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{}`))
}

func (f *FakeSyncthing) handlePendingFolders(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{}`))
}

func (f *FakeSyncthing) handleDiscovery(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte(`{}`))
}
