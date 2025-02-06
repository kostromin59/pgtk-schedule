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
	CurrentWeekLessons(stream string) ([]models.Lesson, error)
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

func (s *schedule) dateLessons(stream string, date time.Time) ([]models.Lesson, error) {
	l, err := s.portal.CurrentWeekLessons(stream)
	if err != nil {
		return nil, err
	}

	date = date.Truncate(24 * time.Hour)

	lessons := make([]models.Lesson, 0, len(l))
	for _, lesson := range l {
		lessonDate := lesson.DateStart.Truncate(24 * time.Hour)
		if date.Equal(lessonDate) {
			lessons = append(lessons, lesson)
		}
	}

	return lessons, nil
}

func (s *schedule) TodayLessons(stream string) ([]models.Lesson, error) {
	return s.dateLessons(stream, time.Now())
}

func (s *schedule) TomorrowLessons(stream string) ([]models.Lesson, error) {
	return s.dateLessons(stream, time.Now().Add(24*time.Hour))
}

func (s *schedule) CurrentWeekLessons(stream string) ([]models.Lesson, error) {
	return s.portal.CurrentWeekLessons(stream)
}
