package service

import "context"

type ServiceManager interface {
	Start(ctx context.Context) error
	Stop() error
	IsRunning(ctx context.Context) bool
	EnableAutostart() error
	DisableAutostart() error
	IsAutostartEnabled() bool
}

var _ ServiceManager = (*Manager)(nil)
