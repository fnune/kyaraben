package minui

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
)

type MenuUI struct {
	binPath string
}

func NewMenuUI(pakPath string) *MenuUI {
	return &MenuUI{
		binPath: filepath.Join(pakPath, "minui-list"),
	}
}

func (m *MenuUI) Show(items []ui.MenuItem, options ui.MenuOptions) (int, ui.Action, error) {
	if len(items) == 0 {
		return -1, ui.ActionBack, nil
	}

	args := []string{}
	if options.Title != "" {
		args = append(args, "-t", options.Title)
	}
	if options.StartIndex > 0 {
		args = append(args, "-i", strconv.Itoa(options.StartIndex))
	}

	var input bytes.Buffer
	for _, item := range items {
		input.WriteString(item.Label)
		input.WriteString("\n")
	}

	cmd := exec.Command(m.binPath, args...)
	cmd.Stdin = &input
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 2:
				return -1, ui.ActionBack, nil
			case 3:
				return -1, ui.ActionMenu, nil
			}
		}
		return -1, ui.ActionBack, fmt.Errorf("minui-list failed: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	for i, item := range items {
		if item.Label == selected {
			return i, ui.ActionSelect, nil
		}
	}

	return -1, ui.ActionBack, nil
}
