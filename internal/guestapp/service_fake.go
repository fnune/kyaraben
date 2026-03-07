package guestapp

import "context"

type FakeServiceManager struct {
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

func NewFakeServiceManager() *FakeServiceManager {
	return &FakeServiceManager{}
}

func (f *FakeServiceManager) Start(ctx context.Context) error {
	f.StartCalls++
	if f.StartErr != nil {
		return f.StartErr
	}
	f.Running = true
	return nil
}

func (f *FakeServiceManager) Stop() error {
	f.StopCalls++
	if f.StopErr != nil {
		return f.StopErr
	}
	f.Running = false
	return nil
}

func (f *FakeServiceManager) IsRunning(ctx context.Context) bool {
	return f.Running
}

func (f *FakeServiceManager) EnableAutostart() error {
	f.EnableCalls++
	if f.EnableErr != nil {
		return f.EnableErr
	}
	f.AutostartEnabled = true
	return nil
}

func (f *FakeServiceManager) DisableAutostart() error {
	f.DisableCalls++
	if f.DisableErr != nil {
		return f.DisableErr
	}
	f.AutostartEnabled = false
	return nil
}

func (f *FakeServiceManager) IsAutostartEnabled() bool {
	return f.AutostartEnabled
}

var _ ServiceManager = (*FakeServiceManager)(nil)
