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
	outputFile, err := os.CreateTemp("", "minui-keyboard-*.txt")
	if err != nil {
		return "", fmt.Errorf("create temp file: %w", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()
	defer func() { _ = os.Remove(outputPath) }()

	args := []string{"--write-location", outputPath}
	if options.Title != "" {
		args = append(args, "--title", options.Title)
	}
	if options.InitialValue != "" {
		args = append(args, "--initial-value", options.InitialValue)
	}

	cmd := exec.Command(k.binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			if exitErr.ExitCode() == 2 || exitErr.ExitCode() == 3 {
				return "", nil
			}
		}
		return "", fmt.Errorf("minui-keyboard failed: %w", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		return "", fmt.Errorf("read output: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}
