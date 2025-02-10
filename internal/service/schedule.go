package service

import (
	"context"
	"log"
	"pgtk-schedule/internal/models"
	"time"
)

type portal interface {
	Update() error
	CurrentWeekLessons(stream, substream string) ([]models.Lesson, error)
}

type schedule struct {
	portal portal

	onStreamLessonsChange func(stream string, lessons []models.Lesson)
	previousLessons       map[string]models.Lesson
}

func NewSchedule(portal portal) *schedule {
	return &schedule{
		portal:          portal,
		previousLessons: make(map[string]models.Lesson),
	}
}

func (s *schedule) Update() error {
	if err := s.portal.Update(); err != nil {
		return err
	}

	// TODO: check changes
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
	l, err := s.portal.CurrentWeekLessons(stream, substream)
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
	return s.portal.CurrentWeekLessons(stream, substream)
}

// func (s *schedule) OnStreamLessonsChange(fn func(stream string, lessons []models.Lesson)) {
// 	s.onStreamLessonsChange = fn
// }
