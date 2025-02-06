package service

import (
	"context"
	"log"
	"pgtk-schedule/internal/models"
	"sync"
	"time"
)

type portal interface {
	Update() error
	Streams() []models.Stream
	CurrentWeekLessons(stream string) ([]models.Lesson, error)
}

type schedule struct {
	portal portal

	// Stream lessons cache by date
	cache map[time.Time]map[string][]models.Lesson
	mu    sync.RWMutex
}

func NewSchedule(portal portal) *schedule {
	return &schedule{
		portal: portal,
		cache:  make(map[time.Time]map[string][]models.Lesson),
	}
}

func (s *schedule) Update() error {
	if err := s.portal.Update(); err != nil {
		return err
	}

	// Clear cache
	s.cache = make(map[time.Time]map[string][]models.Lesson)

	return nil
}

func (s *schedule) RunUpdater(ctx context.Context, d time.Duration) {
	go func() {
		if err := s.Update(); err != nil {
			log.Printf("error while updating portal: %s", err.Error())
		}

		ticker := time.NewTicker(d)
		for {
			select {
			case <-ticker.C:
				if err := s.Update(); err != nil {
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

func (s *schedule) dateLessonsWithCache(stream string, date time.Time) ([]models.Lesson, error) {
	s.mu.RLock()
	dateCache, ok := s.cache[date]
	s.mu.RUnlock()

	if !ok {
		lessons, err := s.dateLessons(stream, date)
		if err != nil {
			return nil, err
		}

		// Save into cache
		s.mu.Lock()
		defer s.mu.Unlock()
		s.cache[date] = map[string][]models.Lesson{
			stream: lessons,
		}

		return lessons, nil
	}

	s.mu.RLock()
	streamCache, ok := dateCache[stream]
	s.mu.RUnlock()

	if !ok {
		lessons, err := s.dateLessons(stream, date)
		if err != nil {
			return nil, err
		}

		// Save into cache
		s.mu.Lock()
		defer s.mu.Unlock()
		dateCache[stream] = lessons
		return lessons, nil
	}

	return streamCache, nil
}

func (s *schedule) TodayLessons(stream string) ([]models.Lesson, error) {
	return s.dateLessonsWithCache(stream, time.Now())
}

func (s *schedule) TomorrowLessons(stream string) ([]models.Lesson, error) {
	return s.dateLessonsWithCache(stream, time.Now().Add(24*time.Hour))
}

func (s *schedule) CurrentWeekLessons(stream string) ([]models.Lesson, error) {
	return s.portal.CurrentWeekLessons(stream)
}
