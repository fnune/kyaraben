package apply

import (
	"bufio"
	"os"
	"strconv"
	"strings"
	"time"
)

type NetMonitor struct {
	lastBytes int64
	lastTime  time.Time
	onSpeed   func(bytesPerSec int64)
	stop      chan struct{}
}

func NewNetMonitor(onSpeed func(bytesPerSec int64)) *NetMonitor {
	return &NetMonitor{
		onSpeed: onSpeed,
		stop:    make(chan struct{}),
	}
}

func (m *NetMonitor) Start() {
	m.lastBytes = readRxBytes()
	m.lastTime = time.Now()

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				bytes := readRxBytes()
				elapsed := now.Sub(m.lastTime).Seconds()
				if elapsed > 0 && bytes > m.lastBytes {
					speed := int64(float64(bytes-m.lastBytes) / elapsed)
					m.onSpeed(speed)
				}
				m.lastBytes = bytes
				m.lastTime = now
			case <-m.stop:
				return
			}
		}
	}()
}

func (m *NetMonitor) Stop() {
	close(m.stop)
}

func readRxBytes() int64 {
	f, err := os.Open("/proc/net/dev")
	if err != nil {
		return 0
	}
	defer f.Close()

	var total int64
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.Contains(line, ":") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 2 {
			continue
		}
		// Skip loopback
		if strings.HasPrefix(parts[0], "lo:") {
			continue
		}
		// rx_bytes is the first number after the interface name
		rxBytes, _ := strconv.ParseInt(parts[1], 10, 64)
		total += rxBytes
	}
	return total
}
