package models

import "time"

type Week struct {
	ID        string
	Text      string
	StartDate time.Time
	EndDate   time.Time
}
