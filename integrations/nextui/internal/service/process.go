package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ProcessManager struct {
	pidFile  string
	homePath string
}

func NewProcessManager(dataDir string) *ProcessManager {
	return &ProcessManager{
		pidFile:  filepath.Join(dataDir, "syncthing.pid"),
		homePath: filepath.Join(dataDir, "syncthing"),
	}
}

func (p *ProcessManager) IsOurProcess(pid int) bool {
	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return false
	}
	return strings.Contains(string(cmdline), p.homePath)
}

func (p *ProcessManager) GetRunningPID() (int, bool) {
	data, err := os.ReadFile(p.pidFile)
	if err != nil {
		return 0, false
	}

	pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, false
	}

	if !p.IsOurProcess(pid) {
		_ = os.Remove(p.pidFile)
		return 0, false
	}

	return pid, true
}

func (p *ProcessManager) WritePID(pid int) error {
	if err := os.MkdirAll(filepath.Dir(p.pidFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(p.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (p *ProcessManager) RemovePID() error {
	err := os.Remove(p.pidFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (p *ProcessManager) StopProcess() error {
	pid, running := p.GetRunningPID()
	if !running {
		return nil
	}

	proc, err := os.FindProcess(pid)
	if err != nil {
		return p.RemovePID()
	}

	if err := proc.Signal(syscall.SIGTERM); err != nil {
		_ = proc.Kill()
		_ = p.RemovePID()
		return nil
	}

	done := make(chan struct{})
	go func() {
		for i := 0; i < 50; i++ {
			if !p.IsOurProcess(pid) {
				close(done)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		_ = proc.Kill()
	}

	return p.RemovePID()
}

func (p *ProcessManager) HomePath() string {
	return p.homePath
}

func (p *ProcessManager) PIDFile() string {
	return p.pidFile
}
