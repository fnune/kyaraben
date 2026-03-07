package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const (
	markerStart = "# BEGIN KYARABEN"
	markerEnd   = "# END KYARABEN"
)

type AutostartManager struct {
	userdataPath string
	platform     string
	pakPath      string
	logsPath     string
}

func NewAutostartManager(userdataPath, platform, pakPath, logsPath string) *AutostartManager {
	return &AutostartManager{
		userdataPath: userdataPath,
		platform:     platform,
		pakPath:      pakPath,
		logsPath:     logsPath,
	}
}

func (a *AutostartManager) autoShPath() string {
	return filepath.Join(a.userdataPath, "auto.sh")
}

func (a *AutostartManager) IsEnabled() bool {
	content, err := os.ReadFile(a.autoShPath())
	if err != nil {
		return false
	}
	return strings.Contains(string(content), markerStart)
}

func (a *AutostartManager) Enable() error {
	if a.IsEnabled() {
		return nil
	}

	block := a.generateBlock()

	content, err := os.ReadFile(a.autoShPath())
	if os.IsNotExist(err) {
		content = []byte("#!/bin/sh\n")
	} else if err != nil {
		return err
	}

	newContent := string(content)
	if !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += "\n" + block

	return os.WriteFile(a.autoShPath(), []byte(newContent), 0755)
}

func (a *AutostartManager) Disable() error {
	content, err := os.ReadFile(a.autoShPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}

	newContent := a.removeBlock(string(content))
	return os.WriteFile(a.autoShPath(), []byte(newContent), 0755)
}

func (a *AutostartManager) generateBlock() string {
	dataDir := filepath.Join(a.userdataPath, "kyaraben")
	pidFile := filepath.Join(dataDir, "syncthing.pid")
	homePath := filepath.Join(dataDir, "syncthing")
	logFile := filepath.Join(a.logsPath, "kyaraben-syncthing.log")

	return fmt.Sprintf(`%s
kyaraben_start_syncthing() {
    SYNCTHING="%s/syncthing"
    PIDFILE="%s"
    HOME_PATH="%s"
    LOGFILE="%s"

    [ ! -x "$SYNCTHING" ] && return

    if [ -f "$PIDFILE" ]; then
        PID=$(cat "$PIDFILE")
        if [ -d "/proc/$PID" ] && grep -q "$HOME_PATH" "/proc/$PID/cmdline" 2>/dev/null; then
            return
        fi
        rm -f "$PIDFILE"
    fi

    mkdir -p "$(dirname "$PIDFILE")"
    "$SYNCTHING" \
        --home="$HOME_PATH" \
        --no-browser \
        --no-upgrade \
        --gui-address="0.0.0.0:8484" \
        > "$LOGFILE" 2>&1 &
    echo $! > "$PIDFILE"
}
kyaraben_start_syncthing
%s
`, markerStart, a.pakPath, pidFile, homePath, logFile, markerEnd)
}

func (a *AutostartManager) removeBlock(content string) string {
	startIdx := strings.Index(content, markerStart)
	if startIdx == -1 {
		return content
	}

	endIdx := strings.Index(content, markerEnd)
	if endIdx == -1 {
		return content
	}

	endIdx += len(markerEnd)
	for endIdx < len(content) && content[endIdx] == '\n' {
		endIdx++
	}

	for startIdx > 0 && content[startIdx-1] == '\n' {
		startIdx--
	}
	if startIdx > 0 {
		startIdx++
	}

	return content[:startIdx] + content[endIdx:]
}
