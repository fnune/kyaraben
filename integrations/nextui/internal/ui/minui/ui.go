package minui

import "github.com/fnune/kyaraben/integrations/nextui/internal/ui"

type UI struct {
	menu      *MenuUI
	keyboard  *KeyboardUI
	presenter *PresenterUI
}

func New(pakPath string) *UI {
	return &UI{
		menu:      NewMenuUI(pakPath),
		keyboard:  NewKeyboardUI(pakPath),
		presenter: NewPresenterUI(pakPath),
	}
}

func (u *UI) Menu() ui.MenuUI {
	return u.menu
}

func (u *UI) Keyboard() ui.KeyboardUI {
	return u.keyboard
}

func (u *UI) Presenter() ui.PresenterUI {
	return u.presenter
}

var _ ui.UI = (*UI)(nil)
