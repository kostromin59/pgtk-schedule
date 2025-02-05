package schedule

import "pgtk-schedule/internal/models"

type portal interface {
	Update() error
	Streams() []models.Stream
	Lessons(stream string) ([]models.Lesson, error)
	TodayLessons(stream string) ([]models.Lesson, error)
	TomorrowLessons(stream string) ([]models.Lesson, error)
}

type schedule struct {
	portal portal
}

func NewSchedule(portal portal) *schedule {
	return &schedule{
		portal: portal,
	}
}
