package minui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
)

type PresenterUI struct {
	binPath string
	cmd     *exec.Cmd
}

func NewPresenterUI(pakPath string) *PresenterUI {
	return &PresenterUI{
		binPath: filepath.Join(pakPath, "minui-presenter"),
	}
}

func (p *PresenterUI) ShowMessage(title, text string) error {
	if err := p.Close(); err != nil {
		return err
	}

	message := title
	if text != "" {
		message = title + "\n\n" + text
	}

	args := []string{"--message", message}
	cmd := exec.Command(p.binPath, args...)
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 2 {
			return nil
		}
		return err
	}
	return nil
}

func (p *PresenterUI) ShowMessageAsync(title, text string) error {
	if err := p.Close(); err != nil {
		return err
	}

	message := title
	if text != "" {
		message = title + "\n\n" + text
	}

	args := []string{"--message", message}
	p.cmd = exec.Command(p.binPath, args...)
	p.cmd.Stderr = os.Stderr

	return p.cmd.Start()
}

func (p *PresenterUI) ShowProgress(title string, percent int) error {
	if err := p.Close(); err != nil {
		return err
	}

	message := fmt.Sprintf("%s\n\n%d%%", title, percent)
	args := []string{"--message", message}
	p.cmd = exec.Command(p.binPath, args...)
	p.cmd.Stderr = os.Stderr

	return p.cmd.Start()
}

func (p *PresenterUI) Close() error {
	if p.cmd != nil && p.cmd.Process != nil {
		if err := p.cmd.Process.Kill(); err != nil {
			if !os.IsPermission(err) {
				return fmt.Errorf("failed to kill presenter: %w", err)
			}
		}
		p.cmd = nil
	}
	return nil
}

var _ ui.PresenterUI = (*PresenterUI)(nil)
