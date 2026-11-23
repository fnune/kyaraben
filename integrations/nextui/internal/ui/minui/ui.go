package minui

import "github.com/fnune/kyaraben/internal/guestapp"

type UI struct {
	menu      *MenuUI
	keyboard  *KeyboardUI
	presenter *PresenterUI
}

func New(binDir string) *UI {
	return &UI{
		menu:      NewMenuUI(binDir),
		keyboard:  NewKeyboardUI(binDir),
		presenter: NewPresenterUI(binDir),
	}
}

func (u *UI) Menu() guestapp.MenuUI {
	return u.menu
}

func (u *UI) Keyboard() guestapp.KeyboardUI {
	return u.keyboard
}

func (u *UI) Presenter() guestapp.PresenterUI {
	return u.presenter
}

var _ guestapp.UI = (*UI)(nil)
