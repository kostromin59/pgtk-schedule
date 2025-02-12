package service

import (
	"context"
	"log"
	"pgtk-schedule/internal/models"
	"slices"
	"time"
)

type schedulePortal interface {
	Update() error
	CurrentWeekLessons(stream, substream string) ([]models.Lesson, error)
}

type schedule struct {
	portal schedulePortal
}

func NewSchedule(portal schedulePortal) *schedule {
	return &schedule{
		portal: portal,
	}
}

func (s *schedule) Update() error {
	if err := s.portal.Update(); err != nil {
		return err
	}

	return nil
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

func (s *schedule) dateLessons(stream, substream string, date time.Time) ([]models.Lesson, error) {
	l, err := s.CurrentWeekLessons(stream, substream)
	if err != nil {
		return nil, err
	}

	date = date.Truncate(24 * time.Hour)

	lessons := make([]models.Lesson, 0, len(l))
	for _, lesson := range l {
		lessonDate := lesson.DateStart.Truncate(24 * time.Hour)
		if !date.Equal(lessonDate) {
			continue
		}

		lessons = append(lessons, lesson)
	}

	return lessons, nil
}

func (s *schedule) TodayLessons(stream, substream string) ([]models.Lesson, error) {
	return s.dateLessons(stream, substream, time.Now())
}

func (s *schedule) TomorrowLessons(stream, substream string) ([]models.Lesson, error) {
	return s.dateLessons(stream, substream, time.Now().Add(24*time.Hour))
}

func (s *schedule) CurrentWeekLessons(stream, substream string) ([]models.Lesson, error) {
	lessons, err := s.portal.CurrentWeekLessons(stream, substream)
	if err != nil {
		return nil, err
	}

	slices.SortFunc(lessons, func(a, b models.Lesson) int {
		if a.DateStart.Before(b.DateStart) {
			return -1
		} else if a.DateStart.After(b.DateStart) {
			return 1
		}

		return 0
	})

	return lessons, nil
}
