package tg

import (
	"errors"
	"fmt"
	"pgtk-schedule/internal/models"
	"strings"
	"time"
	"unicode/utf8"

	"gopkg.in/telebot.v4"
)

const KeyStream = "stream"
const KeySubstream = "substream"

var (
	ErrStreamIsInvalid    = errors.New("Группа не указана или указана неверно")
	ErrSubstreamIsInvalid = errors.New("Подгруппа указана неверно")

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

type scheduleService interface {
	CurrentWeekLessons(stream, substream string) ([]models.Lesson, error)
	TodayLessons(stream, substream string) ([]models.Lesson, error)
	TomorrowLessons(stream, substream string) ([]models.Lesson, error)
}

type schedule struct {
	service scheduleService
}

func NewSchedule(service scheduleService) *schedule {
	return &schedule{
		service: service,
	}
}

func (s *schedule) CurrentWeekLessons() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		streamCtx := ctx.Get(KeyStream)
		substreamCtx := ctx.Get(KeySubstream)

		stream, ok := streamCtx.(string)
		if !ok || stream == "" {
			return ErrStreamIsInvalid
		}

		substream, ok := substreamCtx.(string)
		if !ok {
			return ErrSubstreamIsInvalid
		}

		lessons, err := s.service.CurrentWeekLessons(stream, substream)
		if err != nil {
			return err
		}

		// TODO: Rate limiter + formatting
		return ctx.Send(s.lessonsToMessage(lessons))
	}
}

func (s *schedule) TodayLessons() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		streamCtx := ctx.Get(KeyStream)
		substreamCtx := ctx.Get(KeySubstream)

		stream, ok := streamCtx.(string)
		if !ok || stream == "" {
			return ErrStreamIsInvalid
		}

		substream, ok := substreamCtx.(string)
		if !ok {
			return ErrSubstreamIsInvalid
		}

		lessons, err := s.service.TodayLessons(stream, substream)
		if err != nil {
			return err
		}

		// TODO: Rate limiter + formatting
		return ctx.Send(s.lessonsToMessage(lessons))
	}
}

func (s *schedule) TomorrowLessons() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		streamCtx := ctx.Get(KeyStream)
		substreamCtx := ctx.Get(KeySubstream)

		stream, ok := streamCtx.(string)
		if !ok || stream == "" {
			return ErrStreamIsInvalid
		}

		substream, ok := substreamCtx.(string)
		if !ok {
			return ErrSubstreamIsInvalid
		}

		lessons, err := s.service.TomorrowLessons(stream, substream)
		if err != nil {
			return err
		}

		// TODO: Rate limiter + formatting
		return ctx.Send(s.lessonsToMessage(lessons))
	}
}

func (s *schedule) lessonsToMessage(lessons []models.Lesson) string {
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
