package schedule

import "pgtk-schedule/internal/models"

type portal interface {
	Update() error
	CurrentWeek() (models.Week, error)
	Streams() []models.Stream
}

type schedule struct{}
