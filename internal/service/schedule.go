package service

import (
	"fmt"
	"pgtk-schedule/internal/models"
	"slices"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

var (
	weekdates = map[time.Weekday]string{
		time.Sunday:    "ВОСКРЕСЕНЬЕ",
		time.Monday:    "ПОНЕДЕЛЬНИК",
		time.Tuesday:   "ВТОРНИК",
		time.Wednesday: "СРЕДА",
		time.Thursday:  "ЧЕТВЕРГ",
		time.Friday:    "ПЯТНИЦА",
		time.Saturday:  "СУББОТА",
	}

	weekdayKeys = [...]string{
		weekdates[time.Monday],
		weekdates[time.Tuesday],
		weekdates[time.Wednesday],
		weekdates[time.Thursday],
		weekdates[time.Friday],
		weekdates[time.Saturday],
		weekdates[time.Sunday],
	}
)

type schedulePortal interface {
	Update() error
	CurrentWeekLessons(stream, substream string) ([]models.Lesson, error)
}

type schedule struct {
	portal schedulePortal
	mu     sync.RWMutex
}

func NewSchedule(portal schedulePortal) *schedule {
	return &schedule{
		portal: portal,
	}
}

func (s *schedule) Update() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	err := s.portal.Update()
	return err
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
	s.mu.RLock()
	defer s.mu.RUnlock()

	lessons, err := s.portal.CurrentWeekLessons(stream, substream)
	if err != nil {
		return nil, err
	}

	if len(lessons) == 0 {
		return nil, models.ErrLessonsAreEmpty
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

func (s *schedule) LessonsToString(lessons []models.Lesson) string {
	mapLessons := make(map[string][]models.Lesson, len(weekdates))

	for _, lesson := range lessons {
		weekdayLessons, ok := mapLessons[weekdates[lesson.DateStart.Weekday()]]
		if !ok {
			mapLessons[weekdates[lesson.DateStart.Weekday()]] = []models.Lesson{lesson}
		} else {
			mapLessons[weekdates[lesson.DateStart.Weekday()]] = append(weekdayLessons, lesson)
		}
	}

	sb := strings.Builder{}

	for _, weekday := range weekdayKeys {
		lessons, ok := mapLessons[weekday]
		if !ok {
			continue
		}

		if len(lessons) == 0 {
			continue
		}

		formattedWeekday := "📆 " + weekday
		sb.Grow(utf8.RuneCountInString(formattedWeekday))
		sb.Grow(3)
		sb.WriteString("<b>")
		sb.WriteString(formattedWeekday)

		sb.Grow(3)
		sb.WriteString(" (")
		sb.WriteString(lessons[0].DateStart.Format("02.01.2006"))
		sb.WriteString(")")

		sb.Grow(4)
		sb.WriteString("</b>")

		sb.Grow(1)
		sb.WriteString("\n")

		for i, l := range lessons {
			stringLesson := fmt.Sprintf("<b>%d)</b> %s (%s)\nПреподаватель: %s\nВремя: %s-%s\nКабинет: %s", i+1, l.Name, l.Type, l.Teacher, l.DateStart.Format("15:04"), l.DateEnd.Format("15:04"), l.Cabinet)
			sb.Grow(utf8.RuneCountInString(stringLesson) + 1)
			sb.WriteString(stringLesson)
			sb.WriteString("\n\n")
		}

		sb.WriteString("\n")
	}

	return sb.String()
}
