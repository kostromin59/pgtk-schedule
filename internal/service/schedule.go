package service

import (
	"context"
	"pgtk-schedule/internal/models"
)

type portal interface {
	Update() error
	Streams() []models.Stream
	Lessons(stream string) ([]models.Lesson, error)
	TodayLessons(stream string) ([]models.Lesson, error)
	TomorrowLessons(stream string) ([]models.Lesson, error)
}

type studentRepository interface {
	Create(ctx context.Context, id int64, nickname string) error
	FindByID(ctx context.Context, id int64) (models.Student, error)
	UpdateStream(ctx context.Context, stream string) error
	UpdateSubstream(ctx context.Context, substream string) error
	UpdateNickname(ctx context.Context, substream string) error
}

type schedule struct {
	portal      portal
	studentRepo studentRepository
}

func NewSchedule(portal portal, studentRepo studentRepository) *schedule {
	return &schedule{
		portal:      portal,
		studentRepo: studentRepo,
	}
}
