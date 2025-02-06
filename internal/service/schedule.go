package service

import (
	"context"
	"log"
	"pgtk-schedule/internal/models"
	"time"
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
	UpdateStream(ctx context.Context, id int64, stream string) error
	UpdateSubstream(ctx context.Context, id int64, substream string) error
	UpdateNickname(ctx context.Context, id int64, nickname string) error
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

func (s *schedule) RunUpdater(ctx context.Context, d time.Duration) {
	go func() {
		if err := s.portal.Update(); err != nil {
			log.Printf("error while updating portal: %s", err.Error())
		}

		ticker := time.NewTicker(d)
		for {
			select {
			case <-ticker.C:
				if err := s.portal.Update(); err != nil {
					log.Printf("error while updating portal: %s", err.Error())
				}
			case <-ctx.Done():
				return
			}
		}
	}()
}
