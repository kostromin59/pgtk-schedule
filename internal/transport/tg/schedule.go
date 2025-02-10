package tg

import (
	"errors"
	"fmt"
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
}

type schedule struct {
	service scheduleService
}

func NewSchedule() *schedule {
	return &schedule{}
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
			return ErrStreamIsInvalid
		}

		lessons, err := s.service.CurrentWeekLessons(stream, substream)
		if err != nil {
			return err
		}

		// TODO: Rate limiter + formatting
		return ctx.Send(s.lessonsToMessage(lessons))
	}
}

func (s *schedule) lessonsToMessage(lessons []models.Lesson) string {
	return fmt.Sprintf("count: %d", len(lessons))
}
