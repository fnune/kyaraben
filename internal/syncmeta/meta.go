package syncmeta

import (
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"time"

	"github.com/twpayne/go-vfs/v5"
)

const (
	schemaVersion = 1
	FolderDirName = ".kyaraben"
)

type DeviceStatus struct {
	DeviceID         string
	IgnoreDeleteROMs bool
	PublishedAt      string
	KyarabenVersion  string
}

type markerFile struct {
	SchemaVersion   int        `json:"schemaVersion"`
	DeviceID        string     `json:"deviceId"`
	PublishedAt     string     `json:"publishedAt"`
	KyarabenVersion string     `json:"kyarabenVersion"`
	ROM             romSection `json:"rom"`
}

type romSection struct {
	IgnoreDelete *bool `json:"ignoreDelete"`
}

type Store struct {
	fs  vfs.FS
	dir string
}

func NewStore(filesystem vfs.FS, collectionRoot string) *Store {
	return &Store{fs: filesystem, dir: filepath.Join(collectionRoot, FolderDirName)}
}

func NewDefaultStore(collectionRoot string) *Store {
	return NewStore(vfs.OSFS, collectionRoot)
}

func (s *Store) Publish(deviceID, kyarabenVersion string, ignoreDeleteROMs bool, now time.Time) error {
	if deviceID == "" {
		return errors.New("syncmeta: empty device id")
	}

	filename := deviceID + ".json"
	path := filepath.Join(s.dir, filename)

	if existing, ok := s.readFile(path, filename); ok &&
		existing.IgnoreDeleteROMs == ignoreDeleteROMs &&
		existing.KyarabenVersion == kyarabenVersion {
		return nil
	}

	ignore := ignoreDeleteROMs
	data, err := json.MarshalIndent(markerFile{
		SchemaVersion:   schemaVersion,
		DeviceID:        deviceID,
		PublishedAt:     now.UTC().Format(time.RFC3339),
		KyarabenVersion: kyarabenVersion,
		ROM:             romSection{IgnoreDelete: &ignore},
	}, "", "  ")
	if err != nil {
		return err
	}

	if err := vfs.MkdirAll(s.fs, s.dir, 0o755); err != nil {
		return err
	}
	return s.fs.WriteFile(path, append(data, '\n'), 0o644)
}

func (s *Store) ReadAll() (map[string]DeviceStatus, error) {
	entries, err := s.fs.ReadDir(s.dir)
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]DeviceStatus{}, nil
	}
	if err != nil {
		return nil, err
	}

	out := make(map[string]DeviceStatus)
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".json") {
			continue
		}
		if status, ok := s.readFile(filepath.Join(s.dir, name), name); ok {
			out[status.DeviceID] = status
		}
	}
	return out, nil
}

func (s *Store) Remove(deviceID string) error {
	err := s.fs.Remove(filepath.Join(s.dir, deviceID+".json"))
	if errors.Is(err, fs.ErrNotExist) {
		return nil
	}
	return err
}

func (s *Store) readFile(path, filename string) (DeviceStatus, bool) {
	data, err := s.fs.ReadFile(path)
	if err != nil {
		return DeviceStatus{}, false
	}

	var probe struct {
		SchemaVersion int `json:"schemaVersion"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return DeviceStatus{}, false
	}
	if probe.SchemaVersion < 1 || probe.SchemaVersion > schemaVersion {
		return DeviceStatus{}, false
	}

	var marker markerFile
	if err := json.Unmarshal(data, &marker); err != nil {
		return DeviceStatus{}, false
	}
	if marker.DeviceID == "" || marker.DeviceID+".json" != filename || marker.ROM.IgnoreDelete == nil {
		return DeviceStatus{}, false
	}

	return DeviceStatus{
		DeviceID:         marker.DeviceID,
		IgnoreDeleteROMs: *marker.ROM.IgnoreDelete,
		PublishedAt:      marker.PublishedAt,
		KyarabenVersion:  marker.KyarabenVersion,
	}, true
}
