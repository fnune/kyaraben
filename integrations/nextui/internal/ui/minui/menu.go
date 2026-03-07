package minui

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fnune/kyaraben/internal/guestapp"
)

type MenuUI struct {
	binPath string
}

func NewMenuUI(pakPath string) *MenuUI {
	return &MenuUI{
		binPath: filepath.Join(pakPath, "minui-list"),
	}
}

type jsonItem struct {
	Name     string       `json:"name"`
	Options  []string     `json:"options,omitempty"`
	Selected int          `json:"selected,omitempty"`
	Features jsonFeatures `json:"features,omitempty"`
}

type jsonFeatures struct {
	Unselectable    bool   `json:"unselectable,omitempty"`
	BackgroundColor string `json:"background_color,omitempty"`
	ConfirmText     string `json:"confirm_text,omitempty"`
}

func (m *MenuUI) Show(items []guestapp.MenuItem, options guestapp.MenuOptions) (int, guestapp.Action, error) {
	if len(items) == 0 {
		return -1, guestapp.ActionBack, nil
	}

	outputFile, err := os.CreateTemp("", "minui-list-output-*.txt")
	if err != nil {
		return -1, guestapp.ActionBack, fmt.Errorf("create temp file: %w", err)
	}
	outputPath := outputFile.Name()
	_ = outputFile.Close()
	defer func() { _ = os.Remove(outputPath) }()

	args := []string{
		"--file", "-",
		"--format", "json",
		"--item-key", "items",
		"--write-location", outputPath,
	}
	if options.Title != "" {
		args = append(args, "--title", options.Title)
	}
	if options.StartIndex > 0 {
		args = append(args, "-i", strconv.Itoa(options.StartIndex))
	}

	jsonItems := make([]jsonItem, len(items))
	for i, item := range items {
		jsonItems[i] = jsonItem{
			Name:     item.Label,
			Options:  item.Options,
			Selected: item.Selected,
			Features: jsonFeatures{
				Unselectable:    item.Unselectable,
				BackgroundColor: item.BackgroundColor,
				ConfirmText:     item.ConfirmText,
			},
		}
	}

	wrapper := struct {
		Items []jsonItem `json:"items"`
	}{Items: jsonItems}

	inputData, err := json.Marshal(wrapper)
	if err != nil {
		return -1, guestapp.ActionBack, fmt.Errorf("marshal json: %w", err)
	}

	fmt.Fprintf(os.Stderr, "minui-list JSON input: %s\n", string(inputData))

	inputFile, err := os.CreateTemp("", "minui-list-input-*.json")
	if err != nil {
		return -1, guestapp.ActionBack, fmt.Errorf("create input file: %w", err)
	}
	inputPath := inputFile.Name()
	if _, err := inputFile.Write(inputData); err != nil {
		_ = inputFile.Close()
		_ = os.Remove(inputPath)
		return -1, guestapp.ActionBack, fmt.Errorf("write input file: %w", err)
	}
	_ = inputFile.Close()
	defer func() { _ = os.Remove(inputPath) }()

	args[1] = inputPath // replace "-" with file path

	cmd := exec.Command(m.binPath, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			switch exitErr.ExitCode() {
			case 2:
				return -1, guestapp.ActionBack, nil
			case 3:
				return -1, guestapp.ActionMenu, nil
			}
		}
		return -1, guestapp.ActionBack, fmt.Errorf("minui-list failed: %w", err)
	}

	output, err := os.ReadFile(outputPath)
	if err != nil {
		return -1, guestapp.ActionBack, fmt.Errorf("read output file: %w", err)
	}

	selected := strings.TrimSpace(string(output))
	for i, item := range items {
		if item.Label == selected {
			return i, guestapp.ActionSelect, nil
		}
	}

	return -1, guestapp.ActionBack, nil
}
