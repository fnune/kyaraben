package sync

type FakeUnit struct {
	Enabled bool
	Active  string
	Logs    string
}

type FakeServiceManager struct {
	Units          map[string]*FakeUnit
	DaemonReloaded bool
	Errors         map[string]error
}

func NewFakeServiceManager() *FakeServiceManager {
	return &FakeServiceManager{
		Units:  make(map[string]*FakeUnit),
		Errors: make(map[string]error),
	}
}

func (m *FakeServiceManager) getOrCreate(unit string) *FakeUnit {
	if m.Units[unit] == nil {
		m.Units[unit] = &FakeUnit{Active: "inactive"}
	}
	return m.Units[unit]
}

func (m *FakeServiceManager) DaemonReload() error {
	if err := m.Errors["daemon-reload"]; err != nil {
		return err
	}
	m.DaemonReloaded = true
	return nil
}

func (m *FakeServiceManager) Start(unit string) error {
	if err := m.Errors["start:"+unit]; err != nil {
		return err
	}
	u := m.getOrCreate(unit)
	u.Active = "active"
	return nil
}

func (m *FakeServiceManager) Stop(unit string) error {
	if err := m.Errors["stop:"+unit]; err != nil {
		return err
	}
	u := m.getOrCreate(unit)
	u.Active = "inactive"
	return nil
}

func (m *FakeServiceManager) Restart(unit string) error {
	if err := m.Errors["restart:"+unit]; err != nil {
		return err
	}
	u := m.getOrCreate(unit)
	u.Active = "active"
	return nil
}

func (m *FakeServiceManager) Enable(unit string) error {
	if err := m.Errors["enable:"+unit]; err != nil {
		return err
	}
	u := m.getOrCreate(unit)
	u.Enabled = true
	u.Active = "active"
	return nil
}

func (m *FakeServiceManager) Disable(unit string) error {
	if err := m.Errors["disable:"+unit]; err != nil {
		return err
	}
	u := m.getOrCreate(unit)
	u.Enabled = false
	u.Active = "inactive"
	return nil
}

func (m *FakeServiceManager) IsEnabled(unit string) bool {
	u := m.Units[unit]
	return u != nil && u.Enabled
}

func (m *FakeServiceManager) State(unit string) string {
	u := m.Units[unit]
	if u == nil {
		return ""
	}
	return u.Active
}

func (m *FakeServiceManager) Logs(unit string, _ int) string {
	u := m.Units[unit]
	if u == nil {
		return ""
	}
	return u.Logs
}

func (m *FakeServiceManager) SetState(unit, state string) {
	m.getOrCreate(unit).Active = state
}

func (m *FakeServiceManager) SetLogs(unit, logs string) {
	m.getOrCreate(unit).Logs = logs
}

func (m *FakeServiceManager) SetEnabled(unit string, enabled bool) {
	m.getOrCreate(unit).Enabled = enabled
}
