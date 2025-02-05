package models

import "time"

type Week struct {
	Text      string
	StartDate time.Time
	EndDate   time.Time
}
