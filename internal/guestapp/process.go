package guestapp

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type ProcessController interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning(ctx context.Context) bool
}

type PIDProcessController struct {
	pidFile   string
	markerStr string
	startCmd  func(ctx context.Context) (*exec.Cmd, error)
}

type PIDProcessConfig struct {
	PIDFile   string
	MarkerStr string
	StartCmd  func(ctx context.Context) (*exec.Cmd, error)
}

func NewPIDProcessController(cfg PIDProcessConfig) *PIDProcessController {
	return &PIDProcessController{
		pidFile:   cfg.PIDFile,
		markerStr: cfg.MarkerStr,
		startCmd:  cfg.StartCmd,
	}
}

func (p *PIDProcessController) Start(ctx context.Context) error {
	if p.IsRunning(ctx) {
		return nil
	}

	cmd, err := p.startCmd(ctx)
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	return p.WritePID(cmd.Process.Pid)
}

func (p *PIDProcessController) Stop(ctx context.Context) error {
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

func (p *PIDProcessController) IsRunning(ctx context.Context) bool {
	_, running := p.GetRunningPID()
	return running
}

func (p *PIDProcessController) IsOurProcess(pid int) bool {
	cmdline, err := os.ReadFile(fmt.Sprintf("/proc/%d/cmdline", pid))
	if err != nil {
		return false
	}
	return strings.Contains(string(cmdline), p.markerStr)
}

func (p *PIDProcessController) GetRunningPID() (int, bool) {
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

func (p *PIDProcessController) WritePID(pid int) error {
	if err := os.MkdirAll(filepath.Dir(p.pidFile), 0755); err != nil {
		return err
	}
	return os.WriteFile(p.pidFile, []byte(strconv.Itoa(pid)), 0644)
}

func (p *PIDProcessController) RemovePID() error {
	err := os.Remove(p.pidFile)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func (p *PIDProcessController) PIDFile() string {
	return p.pidFile
}

var _ ProcessController = (*PIDProcessController)(nil)
