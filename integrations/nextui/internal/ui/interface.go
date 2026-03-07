package ui

import "github.com/fnune/kyaraben/internal/guestapp"

type Action = guestapp.Action

const (
	ActionSelect = guestapp.ActionSelect
	ActionBack   = guestapp.ActionBack
	ActionMenu   = guestapp.ActionMenu
)

type MenuItem = guestapp.MenuItem
type MenuOptions = guestapp.MenuOptions
type KeyboardOptions = guestapp.KeyboardOptions
type MenuUI = guestapp.MenuUI
type KeyboardUI = guestapp.KeyboardUI
type PresenterUI = guestapp.PresenterUI
type UI = guestapp.UI
