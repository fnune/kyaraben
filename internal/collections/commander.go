package collections

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fnune/kyaraben/internal/logging"
)

var log = logging.New("collections")

type Commander interface {
	Send(port int, command string, timeout time.Duration) error
}

// RetroArchCommander talks to RetroArch's network command interface. A
// datagram sent before RetroArch binds the port is lost, so it waits for
// the bind first.
type RetroArchCommander struct {
	ProcNetUDP string
}

func NewRetroArchCommander() *RetroArchCommander {
	return &RetroArchCommander{ProcNetUDP: "/proc/net/udp"}
}

func (c *RetroArchCommander) Send(port int, command string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	for !c.bound(port) {
		if time.Now().After(deadline) {
			return fmt.Errorf("RetroArch did not open command port %d", port)
		}
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(time.Second) // the port opens before the content finishes loading
	conn, err := net.Dial("udp", net.JoinHostPort("127.0.0.1", strconv.Itoa(port)))
	if err != nil {
		return err
	}
	defer func() { _ = conn.Close() }()
	_, err = conn.Write([]byte(command))
	return err
}

func (c *RetroArchCommander) bound(port int) bool {
	f, err := os.Open(c.ProcNetUDP)
	if err != nil {
		return false
	}
	defer func() { _ = f.Close() }()
	suffix := fmt.Sprintf(":%04X", port)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) > 1 && strings.HasSuffix(fields[1], suffix) {
			return true
		}
	}
	return false
}
