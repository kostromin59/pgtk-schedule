package schedule

import "pgtk-schedule/internal/models"

type portal interface {
	Update() error
	// CurrentWeek() (models.Week, error) // Do I need this?
	Streams() []models.Stream
	Lessons(stream string) []models.Lesson
	TodayLessons(stream string) []models.Lesson
}

type schedule struct {
	portal portal
}

func NewSchedule(portal portal) *schedule {
	return &schedule{
		portal: portal,
	}
}
