package collections

import (
	"bytes"
	"fmt"
	"path/filepath"
	"text/template"

	"github.com/twpayne/go-vfs/v5"

	"github.com/fnune/kyaraben/internal/sync"
)

const socketTemplate = `[Unit]
Description=Kyaraben collections launcher socket

[Socket]
ListenStream=127.0.0.1:{{.Port}}

[Install]
WantedBy=sockets.target
`

// KillMode=process: games share the service's cgroup and must survive its restarts.
const serviceTemplate = `[Unit]
Description=Kyaraben collections launcher
Requires={{.Name}}.socket

[Service]
Type=simple
ExecStart={{.Binary}}{{if .Instance}} --instance={{.Instance}}{{end}} collections serve --port={{.Port}}
KillMode=process
Restart=on-failure
`

type unitParams struct {
	Name     string
	Binary   string
	Instance string
	Port     int
}

type Units struct {
	fs       vfs.FS
	service  sync.ServiceManager
	unitDir  string
	name     string
	instance string
}

func NewUnits(fs vfs.FS, service sync.ServiceManager, unitDir, name, instance string) *Units {
	return &Units{fs: fs, service: service, unitDir: unitDir, name: name, instance: instance}
}

func (u *Units) socketPath() string  { return filepath.Join(u.unitDir, u.name+".socket") }
func (u *Units) servicePath() string { return filepath.Join(u.unitDir, u.name+".service") }

func (u *Units) Ensure(binary string, port int) error {
	params := unitParams{Name: u.name, Binary: binary, Instance: u.instance, Port: port}
	socketChanged, err := u.write(u.socketPath(), socketTemplate, params)
	if err != nil {
		return err
	}
	if _, err := u.write(u.servicePath(), serviceTemplate, params); err != nil {
		return err
	}
	if err := u.service.DaemonReload(); err != nil {
		return fmt.Errorf("daemon-reload: %w", err)
	}
	socket := u.name + ".socket"
	if socketChanged && u.service.IsEnabled(socket) {
		_ = u.service.Stop(u.name + ".service")
		if err := u.service.Restart(socket); err != nil {
			return fmt.Errorf("restarting %s: %w", socket, err)
		}
	} else if err := u.service.Enable(socket); err != nil {
		return fmt.Errorf("enabling %s: %w", socket, err)
	}
	if state := u.service.State(u.name + ".service"); state == "active" { // it reads config only at start
		if err := u.service.Restart(u.name + ".service"); err != nil {
			return fmt.Errorf("restarting %s.service: %w", u.name, err)
		}
	}
	return nil
}

func (u *Units) Installed() bool {
	_, err := u.fs.Stat(u.socketPath())
	return err == nil
}

func (u *Units) Remove() error {
	_ = u.service.Disable(u.name + ".socket")
	_ = u.service.Stop(u.name + ".service")
	for _, p := range []string{u.socketPath(), u.servicePath()} {
		if err := u.fs.Remove(p); err != nil && !isNotExist(err) {
			return err
		}
	}
	return u.service.DaemonReload()
}

func (u *Units) write(path, tmpl string, params unitParams) (bool, error) {
	var buf bytes.Buffer
	if err := template.Must(template.New("unit").Parse(tmpl)).Execute(&buf, params); err != nil {
		return false, err
	}
	current, err := u.fs.ReadFile(path)
	if err == nil && bytes.Equal(current, buf.Bytes()) {
		return false, nil
	}
	if err := vfs.MkdirAll(u.fs, u.unitDir, 0o755); err != nil {
		return false, err
	}
	if err := u.fs.WriteFile(path, buf.Bytes(), 0o644); err != nil {
		return false, fmt.Errorf("writing %s: %w", path, err)
	}
	return true, nil
}
