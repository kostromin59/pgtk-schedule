package models

import "time"

type Lesson struct {
	ID        string
	Name      string
	Cabinet   string
	Type      string
	Teacher   string
	Stream    string
	Substream string
	DateStart time.Time
	DateEnd   time.Time
}
