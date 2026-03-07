package guestapp

type FakeMenuUI struct {
	SelectIndex  int
	SelectAction Action
	SelectFunc   func(items []MenuItem) (int, Action)
	Err          error
	ShowCalls    []MenuShowCall
}

type MenuShowCall struct {
	Items   []MenuItem
	Options MenuOptions
}

func (m *FakeMenuUI) Show(items []MenuItem, options MenuOptions) (int, Action, error) {
	m.ShowCalls = append(m.ShowCalls, MenuShowCall{Items: items, Options: options})
	if m.SelectFunc != nil {
		idx, action := m.SelectFunc(items)
		return idx, action, m.Err
	}
	return m.SelectIndex, m.SelectAction, m.Err
}

type FakeKeyboardUI struct {
	Input    string
	Err      error
	GetCalls []KeyboardOptions
}

func (k *FakeKeyboardUI) GetInput(options KeyboardOptions) (string, error) {
	k.GetCalls = append(k.GetCalls, options)
	return k.Input, k.Err
}

type FakePresenterUI struct {
	Messages []FakeMessageCall
	Progress []FakeProgressCall
	Closed   bool
	Err      error
}

type FakeMessageCall struct {
	Title string
	Text  string
}

type FakeProgressCall struct {
	Title   string
	Percent int
}

func (p *FakePresenterUI) ShowMessage(title, text string) error {
	p.Messages = append(p.Messages, FakeMessageCall{Title: title, Text: text})
	return p.Err
}

func (p *FakePresenterUI) ShowMessageAsync(title, text string) error {
	p.Messages = append(p.Messages, FakeMessageCall{Title: title, Text: text})
	return p.Err
}

func (p *FakePresenterUI) ShowProgress(title string, percent int) error {
	p.Progress = append(p.Progress, FakeProgressCall{Title: title, Percent: percent})
	return p.Err
}

func (p *FakePresenterUI) Close() error {
	p.Closed = true
	return p.Err
}

type FakeUI struct {
	MenuUI      *FakeMenuUI
	KeyboardUI  *FakeKeyboardUI
	PresenterUI *FakePresenterUI
}

func NewFakeUI() *FakeUI {
	return &FakeUI{
		MenuUI:      &FakeMenuUI{},
		KeyboardUI:  &FakeKeyboardUI{},
		PresenterUI: &FakePresenterUI{},
	}
}

func (u *FakeUI) Menu() MenuUI {
	return u.MenuUI
}

func (u *FakeUI) Keyboard() KeyboardUI {
	return u.KeyboardUI
}

func (u *FakeUI) Presenter() PresenterUI {
	return u.PresenterUI
}

var _ UI = (*FakeUI)(nil)
