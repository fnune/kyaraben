package collections

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"time"
)

type Process interface {
	Wait() error
}

type Starter interface {
	Start(command []string, logPath string) (Process, error)
}

type ExecStarter struct{}

// Games run in their own session so that restarting the launcher service
// (or closing the terminal that ran `kyaraben launch`) does not kill them.
func (ExecStarter) Start(command []string, logPath string) (Process, error) {
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return nil, err
	}
	logFile, err := os.Create(logPath)
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(command[0], command[1:]...)
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if err := cmd.Start(); err != nil {
		_ = logFile.Close()
		return nil, err
	}
	return &execProcess{cmd: cmd, logFile: logFile}, nil
}

type execProcess struct {
	cmd     *exec.Cmd
	logFile *os.File
}

func (p *execProcess) Wait() error {
	defer func() { _ = p.logFile.Close() }()
	return p.cmd.Wait()
}

type Session struct {
	ID         string     `json:"id"`
	Collection string     `json:"collection"`
	Game       string     `json:"game"`
	Title      string     `json:"title"`
	Emulator   string     `json:"emulator"`
	Set        string     `json:"set"`
	Running    bool       `json:"running"`
	StartedAt  time.Time  `json:"startedAt"`
	EndedAt    *time.Time `json:"endedAt,omitempty"`
	Error      string     `json:"error,omitempty"`
}

var ErrGameRunning = errors.New("a game is already running")

// Sessions tracks launched games so a client can tell when a play session
// ends, for example to offer logging a run.
type Sessions struct {
	mu       sync.Mutex
	starter  Starter
	logDir   string
	now      func() time.Time
	sessions map[string]*Session
	order    []string
}

const maxRememberedSessions = 50

func NewSessions(starter Starter, logDir string) *Sessions {
	return &Sessions{starter: starter, logDir: logDir, now: time.Now, sessions: map[string]*Session{}}
}

func (s *Sessions) Launch(plan *LaunchPlan) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.sessions {
		if existing.Running {
			return Session{}, fmt.Errorf("%w: %s", ErrGameRunning, existing.Title)
		}
	}

	logPath := filepath.Join(s.logDir, plan.CollectionID+"-"+plan.GameID+".log")
	proc, err := s.starter.Start(plan.Command, logPath)
	if err != nil {
		return Session{}, fmt.Errorf("starting %s: %w", plan.Title, err)
	}

	session := &Session{
		ID:         newSessionID(),
		Collection: plan.CollectionID,
		Game:       plan.GameID,
		Title:      plan.Title,
		Emulator:   string(plan.Resolution.Emulator),
		Set:        plan.Resolution.Set,
		Running:    true,
		StartedAt:  s.now(),
	}
	s.remember(session)

	go func() {
		waitErr := proc.Wait()
		s.mu.Lock()
		defer s.mu.Unlock()
		ended := s.now()
		session.Running = false
		session.EndedAt = &ended
		if waitErr != nil {
			session.Error = waitErr.Error()
		}
	}()

	return *session, nil
}

func (s *Sessions) Get(id string) (Session, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	session, ok := s.sessions[id]
	if !ok {
		return Session{}, false
	}
	return *session, true
}

func (s *Sessions) remember(session *Session) {
	s.sessions[session.ID] = session
	s.order = append(s.order, session.ID)
	for len(s.order) > maxRememberedSessions {
		oldest := s.order[0]
		if s.sessions[oldest].Running {
			break
		}
		delete(s.sessions, oldest)
		s.order = s.order[1:]
	}
}

func newSessionID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
