package sync

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
)

func FindPIDByPort(port int) (int, error) {
	inode, err := findInodeByPort(port)
	if err != nil {
		return 0, err
	}
	if inode == 0 {
		return 0, nil
	}

	return findPIDByInode(inode)
}

func findInodeByPort(port int) (uint64, error) {
	file, err := os.Open("/proc/net/tcp")
	if err != nil {
		return 0, fmt.Errorf("opening /proc/net/tcp: %w", err)
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	scanner.Scan()

	targetAddr := fmt.Sprintf("0100007F:%04X", port)

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		localAddr := fields[1]
		if localAddr == targetAddr {
			inode, err := strconv.ParseUint(fields[9], 10, 64)
			if err != nil {
				continue
			}
			return inode, nil
		}
	}

	return 0, nil
}

func findPIDByInode(inode uint64) (int, error) {
	procDir, err := os.Open("/proc")
	if err != nil {
		return 0, fmt.Errorf("opening /proc: %w", err)
	}
	defer func() { _ = procDir.Close() }()

	entries, err := procDir.Readdirnames(-1)
	if err != nil {
		return 0, fmt.Errorf("reading /proc: %w", err)
	}

	targetSocket := fmt.Sprintf("socket:[%d]", inode)

	for _, entry := range entries {
		pid, err := strconv.Atoi(entry)
		if err != nil {
			continue
		}

		fdDir := filepath.Join("/proc", entry, "fd")
		fds, err := os.ReadDir(fdDir)
		if err != nil {
			continue
		}

		for _, fd := range fds {
			linkPath := filepath.Join(fdDir, fd.Name())
			link, err := os.Readlink(linkPath)
			if err != nil {
				continue
			}
			if link == targetSocket {
				return pid, nil
			}
		}
	}

	return 0, nil
}

func IsKyarabenSyncthing(pid int, stateDir string) bool {
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return false
	}

	cmdline := string(data)

	if !strings.Contains(cmdline, "syncthing") {
		return false
	}

	configArg := "--config=" + filepath.Join(stateDir, "syncthing", "config")
	if strings.Contains(cmdline, configArg) {
		return true
	}

	parts := strings.Split(cmdline, "\x00")
	for i, part := range parts {
		if part == "--config" && i+1 < len(parts) {
			configPath := parts[i+1]
			expectedPath := filepath.Join(stateDir, "syncthing", "config")
			if configPath == expectedPath {
				return true
			}
		}
	}

	return false
}

func IsKyarabenInstance(pid int) bool {
	cmdlinePath := fmt.Sprintf("/proc/%d/cmdline", pid)
	data, err := os.ReadFile(cmdlinePath)
	if err != nil {
		return false
	}

	cmdline := string(data)

	if !strings.Contains(cmdline, "syncthing") {
		return false
	}

	return strings.Contains(cmdline, "kyaraben")
}

func KillProcess(pid int, timeout time.Duration) error {
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("finding process %d: %w", pid, err)
	}

	if err := process.Signal(syscall.SIGTERM); err != nil {
		if err == os.ErrProcessDone {
			return nil
		}
		return fmt.Errorf("sending SIGTERM to %d: %w", pid, err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return nil
		}
		time.Sleep(100 * time.Millisecond)
	}

	if err := process.Signal(syscall.SIGKILL); err != nil {
		if err == os.ErrProcessDone {
			return nil
		}
		return fmt.Errorf("sending SIGKILL to %d: %w", pid, err)
	}

	for i := 0; i < 10; i++ {
		if err := process.Signal(syscall.Signal(0)); err != nil {
			return nil
		}
		time.Sleep(50 * time.Millisecond)
	}

	return nil
}

func WaitForPortRelease(port int, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	interval := 20 * time.Millisecond

	for time.Now().Before(deadline) {
		pid, err := FindPIDByPort(port)
		if err != nil {
			return err
		}
		if pid == 0 {
			return nil
		}
		time.Sleep(interval)
		if interval < 200*time.Millisecond {
			interval = interval * 2
		}
	}
	return fmt.Errorf("port %d still in use after %v", port, timeout)
}
