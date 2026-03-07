package service

import "context"

type FakeManager struct {
	Running          bool
	AutostartEnabled bool
	StartErr         error
	StopErr          error
	EnableErr        error
	DisableErr       error

	StartCalls   int
	StopCalls    int
	EnableCalls  int
	DisableCalls int
}

func NewFakeManager() *FakeManager {
	return &FakeManager{}
}

func (f *FakeManager) Start(ctx context.Context) error {
	f.StartCalls++
	if f.StartErr != nil {
		return f.StartErr
	}
	f.Running = true
	return nil
}

func (f *FakeManager) Stop() error {
	f.StopCalls++
	if f.StopErr != nil {
		return f.StopErr
	}
	f.Running = false
	return nil
}

func (f *FakeManager) IsRunning(ctx context.Context) bool {
	return f.Running
}

func (f *FakeManager) EnableAutostart() error {
	f.EnableCalls++
	if f.EnableErr != nil {
		return f.EnableErr
	}
	f.AutostartEnabled = true
	return nil
}

func (f *FakeManager) DisableAutostart() error {
	f.DisableCalls++
	if f.DisableErr != nil {
		return f.DisableErr
	}
	f.AutostartEnabled = false
	return nil
}

func (f *FakeManager) IsAutostartEnabled() bool {
	return f.AutostartEnabled
}

var _ ServiceManager = (*FakeManager)(nil)
