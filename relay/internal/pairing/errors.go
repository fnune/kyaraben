package pairing

import "errors"

var (
	ErrSessionNotFound      = errors.New("session not found")
	ErrSessionExpired       = errors.New("session expired")
	ErrResponseAlreadySet   = errors.New("response already set")
	ErrTooManySessions      = errors.New("too many active sessions")
	ErrTooManySessionsForIP = errors.New("too many active sessions for this IP")
)
