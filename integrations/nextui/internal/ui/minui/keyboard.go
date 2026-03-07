package minui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
)

type KeyboardUI struct {
	binPath string
}

func NewKeyboardUI(pakPath string) *KeyboardUI {
	return &KeyboardUI{
		binPath: filepath.Join(pakPath, "minui-keyboard"),
	}
}

func (k *KeyboardUI) GetInput(options ui.KeyboardOptions) (string, error) {
	args := []string{}
	if options.Title != "" {
		args = append(args, "-t", options.Title)
	}
	if options.Placeholder != "" {
		args = append(args, "-p", options.Placeholder)
	}
	if options.Uppercase {
		args = append(args, "-u")
	}

	cmd := exec.Command(k.binPath, args...)
	cmd.Stderr = os.Stderr

	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 2 {
				return "", nil
			}
		}
		return "", fmt.Errorf("minui-keyboard failed: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
