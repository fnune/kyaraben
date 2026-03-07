package pairing

import (
	"crypto/rand"
	"strings"
	"sync"
	"time"
)

const (
	DefaultTTL       = 5 * time.Minute
	CleanupInterval  = 1 * time.Minute
	CodeLength       = 6
	MaxSessionsPerIP = 5
	MaxTotalSessions = 10000
)

type Session struct {
	Code              string
	InitiatorDeviceID string
	ResponderDeviceID string
	CreatedAt         time.Time
	ExpiresAt         time.Time
	CreatorIP         string
}

func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

func (s *Session) HasResponse() bool {
	return s.ResponderDeviceID != ""
}

type Store struct {
	mu       sync.RWMutex
	sessions map[string]*Session
	stopCh   chan struct{}
	ttl      time.Duration
}

func NewStore(ttl time.Duration) *Store {
	s := &Store{
		sessions: make(map[string]*Session),
		stopCh:   make(chan struct{}),
		ttl:      ttl,
	}
	go s.cleanupLoop()
	return s
}

func NewDefaultStore() *Store {
	return NewStore(DefaultTTL)
}

func (s *Store) Close() {
	close(s.stopCh)
}

func (s *Store) Create(initiatorDeviceID, creatorIP string) (*Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.sessions) >= MaxTotalSessions {
		return nil, ErrTooManySessions
	}

	ipCount := 0
	for _, sess := range s.sessions {
		if sess.CreatorIP == creatorIP {
			ipCount++
		}
	}
	if ipCount >= MaxSessionsPerIP {
		return nil, ErrTooManySessionsForIP
	}

	code := generateCode()
	for s.sessions[code] != nil {
		code = generateCode()
	}

	now := time.Now()
	session := &Session{
		Code:              code,
		InitiatorDeviceID: initiatorDeviceID,
		CreatedAt:         now,
		ExpiresAt:         now.Add(s.ttl),
		CreatorIP:         creatorIP,
	}
	s.sessions[code] = session
	return session, nil
}

func (s *Store) Get(code string) (*Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	session := s.sessions[normalizeCode(code)]
	if session == nil {
		return nil, ErrSessionNotFound
	}
	if session.IsExpired() {
		return nil, ErrSessionExpired
	}
	return session, nil
}

func (s *Store) SetResponse(code, responderDeviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := s.sessions[normalizeCode(code)]
	if session == nil {
		return ErrSessionNotFound
	}
	if session.IsExpired() {
		return ErrSessionExpired
	}
	if session.ResponderDeviceID != "" {
		return ErrResponseAlreadySet
	}
	session.ResponderDeviceID = responderDeviceID
	return nil
}

func (s *Store) Delete(code string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.sessions[normalizeCode(code)] == nil {
		return ErrSessionNotFound
	}
	delete(s.sessions, normalizeCode(code))
	return nil
}

func (s *Store) cleanupLoop() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupExpired()
		case <-s.stopCh:
			return
		}
	}
}

func (s *Store) cleanupExpired() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for code, session := range s.sessions {
		if now.After(session.ExpiresAt) {
			delete(s.sessions, code)
		}
	}
}

func (s *Store) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.sessions)
}

// Excludes visually ambiguous characters: 0, 1, I, L, O
const codeAlphabet = "ABCDEFGHJKMNPQRSTUVWXYZ23456789"

func generateCode() string {
	b := make([]byte, CodeLength)
	_, _ = rand.Read(b)
	code := make([]byte, CodeLength)
	for i := 0; i < CodeLength; i++ {
		code[i] = codeAlphabet[int(b[i])%len(codeAlphabet)]
	}
	return string(code)
}

func normalizeCode(code string) string {
	return strings.ToUpper(strings.TrimSpace(code))
}
