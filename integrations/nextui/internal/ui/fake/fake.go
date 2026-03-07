package fake

import (
	"github.com/fnune/kyaraben/integrations/nextui/internal/ui"
)

type MenuUI struct {
	SelectIndex  int
	SelectAction ui.Action
	SelectFunc   func(items []ui.MenuItem) (int, ui.Action)
	Err          error
	ShowCalls    []ShowCall
}

type ShowCall struct {
	Items   []ui.MenuItem
	Options ui.MenuOptions
}

func (m *MenuUI) Show(items []ui.MenuItem, options ui.MenuOptions) (int, ui.Action, error) {
	m.ShowCalls = append(m.ShowCalls, ShowCall{Items: items, Options: options})
	if m.SelectFunc != nil {
		idx, action := m.SelectFunc(items)
		return idx, action, m.Err
	}
	return m.SelectIndex, m.SelectAction, m.Err
}

type KeyboardUI struct {
	Input    string
	Err      error
	GetCalls []ui.KeyboardOptions
}

func (k *KeyboardUI) GetInput(options ui.KeyboardOptions) (string, error) {
	k.GetCalls = append(k.GetCalls, options)
	return k.Input, k.Err
}

type PresenterUI struct {
	Messages []MessageCall
	Progress []ProgressCall
	Closed   bool
	Err      error
}

type MessageCall struct {
	Title string
	Text  string
}

type ProgressCall struct {
	Title   string
	Percent int
}

func (p *PresenterUI) ShowMessage(title, text string) error {
	p.Messages = append(p.Messages, MessageCall{Title: title, Text: text})
	return p.Err
}

func (p *PresenterUI) ShowProgress(title string, percent int) error {
	p.Progress = append(p.Progress, ProgressCall{Title: title, Percent: percent})
	return p.Err
}

func (p *PresenterUI) Close() error {
	p.Closed = true
	return p.Err
}

type UI struct {
	MenuUI      *MenuUI
	KeyboardUI  *KeyboardUI
	PresenterUI *PresenterUI
}

func New() *UI {
	return &UI{
		MenuUI:      &MenuUI{},
		KeyboardUI:  &KeyboardUI{},
		PresenterUI: &PresenterUI{},
	}
}

func (u *UI) Menu() ui.MenuUI {
	return u.MenuUI
}

func (u *UI) Keyboard() ui.KeyboardUI {
	return u.KeyboardUI
}

func (u *UI) Presenter() ui.PresenterUI {
	return u.PresenterUI
}

var _ ui.UI = (*UI)(nil)
