package ui

type Action int

const (
	ActionSelect Action = iota
	ActionBack
	ActionMenu
)

type MenuItem struct {
	Label   string
	Value   string
	Enabled bool
}

type MenuOptions struct {
	Title      string
	ShowBack   bool
	ShowMenu   bool
	StartIndex int
}

type KeyboardOptions struct {
	Title       string
	Placeholder string
	MaxLength   int
	Uppercase   bool
}

type MenuUI interface {
	Show(items []MenuItem, options MenuOptions) (selected int, action Action, err error)
}

type KeyboardUI interface {
	GetInput(options KeyboardOptions) (string, error)
}

type PresenterUI interface {
	ShowMessage(title, text string) error
	ShowProgress(title string, percent int) error
	Close() error
}

type UI interface {
	Menu() MenuUI
	Keyboard() KeyboardUI
	Presenter() PresenterUI
}
