package models

import "errors"

var (
	ErrNotifySettingsNotFound = errors.New("notify settings not found")
)

type NotifySettings struct {
	ID        int64
	StudentID int64
	Morning   bool
	Evening   bool
	Week      bool
}
