package service

import "github.com/fnune/kyaraben/internal/guestapp"

type ServiceManager = guestapp.ServiceManager

var _ ServiceManager = (*Manager)(nil)
