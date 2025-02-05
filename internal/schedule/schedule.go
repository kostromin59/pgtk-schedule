package schedule

import "pgtk-schedule/internal/models"

type portal interface {
	Update() error
	CurrentWeek() (models.Week, error)
}

type schedule struct{}
