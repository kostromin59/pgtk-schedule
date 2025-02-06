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

type schedule struct {
	portal portal
}

func NewSchedule(portal portal) *schedule {
	return &schedule{
		portal: portal,
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

func (s *schedule) Streams() []models.Stream {
	return s.portal.Streams()
}

func (s *schedule) TodayLessons(stream string) ([]models.Lesson, error) {
	return s.portal.TodayLessons(stream)
}

func (s *schedule) TomorrowLessons(stream string) ([]models.Lesson, error) {
	return s.portal.TomorrowLessons(stream)
}

func (s *schedule) CurrentWeekLessons(stream string) ([]models.Lesson, error) {
	return s.portal.Lessons(stream)
}
