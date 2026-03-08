package syncguest

import (
	"errors"
	"strings"
)

var (
	ErrInvalidCode    = errors.New("invalid pairing code")
	ErrCodeExpired    = errors.New("pairing code expired")
	ErrConnectionLost = errors.New("connection lost, try again")
)

func FriendlyPairingError(err error) error {
	if err == nil {
		return nil
	}

	msg := err.Error()

	switch {
	case strings.Contains(msg, "session not found"):
		return ErrInvalidCode
	case strings.Contains(msg, "session expired"):
		return ErrCodeExpired
	case strings.Contains(msg, "context deadline exceeded"):
		return ErrConnectionLost
	case strings.Contains(msg, "connection refused"):
		return ErrConnectionLost
	}

	return err
}
