package tg

import (
	"errors"
	"pgtk-schedule/internal/models"

	"gopkg.in/telebot.v4"
)

const KeyStream = "stream"
const KeySubstream = "substream"

var (
	ErrStreamIsInvalid    = errors.New("Группа не указана или указана неверно")
	ErrSubstreamIsInvalid = errors.New("Подгруппа указана неверно")
)

type scheduleService interface {
	CurrentWeekLessons(stream, substream string) ([]models.Lesson, error)
	TodayLessons(stream, substream string) ([]models.Lesson, error)
	TomorrowLessons(stream, substream string) ([]models.Lesson, error)
	LessonsToString(lessons []models.Lesson) string
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
			if errors.Is(err, models.ErrLessonNotFound) {
				return ctx.Reply("Расписание не найдено! Попробуйте ещё раз через пару минут.")
			}
			return err
		}

		return ctx.Send(s.service.LessonsToString(lessons))
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
			if errors.Is(err, models.ErrLessonsAreEmpty) {
				return ctx.Reply("Расписание не найдено! Попробуйте ещё раз через пару минут.")
			}
			return err
		}

		return ctx.Send(s.service.LessonsToString(lessons))
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
			if errors.Is(err, models.ErrLessonNotFound) {
				return ctx.Reply("Расписание не найдено! Попробуйте ещё раз через пару минут.")
			}
			return err
		}

		return ctx.Send(s.service.LessonsToString(lessons))
	}
}
