package schedule

import "pgtk-schedule/internal/models"

type portal interface {
	Update() error
	CurrentWeek() (models.Week, error)
	Streams() []models.Stream
	Lessons() []models.Lesson
	StreamLessons(string) []models.Lesson
	TodayStreamLessons(string) []models.Lesson
}

type schedule struct {
	portal portal
}

func NewSchedule(portal portal) *schedule {
	return &schedule{
		portal: portal,
	}
}
