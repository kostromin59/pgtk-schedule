package models

import "time"

// NOTE: do I need this?
type Week struct {
	ID        string
	Text      string
	StartDate time.Time
	EndDate   time.Time
}
