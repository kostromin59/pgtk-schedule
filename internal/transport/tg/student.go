package tg

import (
	"context"
	"errors"
	"fmt"
	"pgtk-schedule/internal/models"

	"gopkg.in/telebot.v4"
)

const (
	actionSetStream    = "setStream"
	actionSetSubstream = "setSubstream"

	KeyStudent = "student"
)

type studentService interface {
	Create(ctx context.Context, id int64, nickname string) error
	FindByID(ctx context.Context, id int64) (models.Student, error)
	UpdateStream(ctx context.Context, id int64, stream string) error
	UpdateSubstream(ctx context.Context, id int64, substream string) error
	UpdateNickname(ctx context.Context, id int64, nickname string) error
}

type portal interface {
	Streams() []models.Stream
}

type student struct {
	service studentService
	portal  portal
	bot     *telebot.Bot
}

func NewStudent(bot *telebot.Bot, service studentService, portal portal) *student {
	return &student{
		service: service,
		portal:  portal,
		bot:     bot,
	}
}

func (s *student) RegisteredStudent() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			student, err := s.service.FindByID(context.Background(), ctx.Sender().ID)
			if err != nil {
				if errors.Is(err, models.ErrStudentNotFound) {
					if err := s.service.Create(context.Background(), ctx.Sender().ID, ctx.Sender().Username); err != nil {
						return err
					}

					ctx.Set(KeyStudent, models.Student{
						ID:       ctx.Sender().ID,
						Nickname: &ctx.Sender().Username,
					})

					return next(ctx)
				}

				return err
			}

			ctx.Set(KeyStudent, student)

			if student.Stream != nil {
				ctx.Set(KeyStream, *student.Stream)
			}

			if student.Substream != nil {
				ctx.Set(KeySubstream, *student.Substream)
			}

			return next(ctx)
		}
	}
}

func (s *student) validate(student models.Student) error {
	if student.ID == 0 {
		return models.ErrStudentNotFound
	}

	if student.Stream == nil {
		return models.ErrStudentStreamMissed
	}

	return nil
}

func (s *student) ValidateStudent() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			student := ctx.Get(KeyStudent)

			modelStudent, ok := student.(models.Student)
			if !ok {
				return models.ErrStudentNotFound
			}

			err := s.validate(modelStudent)
			if err != nil {
				if errors.Is(err, models.ErrStudentStreamMissed) {
					return ctx.Reply("Укажите группу с помощью команды /setstream")
				}
				return err
			}

			return next(ctx)
		}
	}
}

func (s *student) SetStream() telebot.HandlerFunc {
	s.bot.Handle("\f"+actionSetStream, func(ctx telebot.Context) error {
		stream := ctx.Callback().Data

		streams := s.portal.Streams()
		var foundStream models.Stream
		for _, s := range streams {
			if s.ID == stream {
				foundStream = s
				break
			}
		}

		if foundStream.ID == "" {
			return models.ErrStreamIsUnknown
		}

		err := s.service.UpdateStream(context.Background(), ctx.Callback().Sender.ID, stream)
		if err != nil {
			return err
		}

		if len(foundStream.Substreams) == 0 {
			_, err := s.bot.Edit(ctx.Callback().Message, fmt.Sprintf("Группа %s установлена!", foundStream.Name))
			return err
		}

		if len(foundStream.Substreams) == 1 {
			err := s.service.UpdateSubstream(context.Background(), ctx.Callback().Sender.ID, foundStream.Substreams[0])
			if err != nil {
				return err
			}
			_, err = s.bot.Edit(ctx.Callback().Message, fmt.Sprintf("Подгруппа %s установлена!", foundStream.Substreams[0]))
			return err
		}

		markup := s.bot.NewMarkup()

		btns := make([]telebot.Row, 0, len(streams))
		for _, substream := range foundStream.Substreams {
			b := markup.Data(substream, actionSetSubstream, substream)
			btns = append(btns, markup.Row(b))
		}
		markup.Inline(btns...)

		_, err = s.bot.Edit(ctx.Callback().Message, "Выберите подгруппу:", markup)
		return err
	})

	s.bot.Handle("\f"+actionSetSubstream, func(ctx telebot.Context) error {
		substream := ctx.Callback().Data

		err := s.service.UpdateSubstream(context.Background(), ctx.Callback().Sender.ID, substream)
		if err != nil {
			return err
		}

		_, err = s.bot.Edit(ctx.Callback().Message, fmt.Sprintf("Подгруппа %s установлена!", substream))
		return err
	})

	return func(ctx telebot.Context) error {
		streams := s.portal.Streams()
		markup := s.bot.NewMarkup()

		btns := make([]telebot.Row, 0, len(streams))
		for _, stream := range streams {
			b := markup.Data(stream.Name, actionSetStream, stream.ID)
			btns = append(btns, markup.Row(b))
		}
		markup.Inline(btns...)

		return ctx.Reply("Выберите группу:", markup)
	}
}
