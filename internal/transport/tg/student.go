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
)

type studentRepository interface {
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
	repo   studentRepository
	portal portal
	bot    *telebot.Bot
}

func NewStudent(bot *telebot.Bot, repo studentRepository, portal portal) *student {
	return &student{
		repo:   repo,
		portal: portal,
		bot:    bot,
	}
}

func (s *student) RegisteredStudent() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			student, err := s.findStudent(ctx)
			if err != nil {
				if errors.Is(err, models.ErrStudentNotFound) {
					if err := s.registerStudent(ctx); err != nil {
						return err
					}

					return next(ctx)
				}

				return err
			}

			// Set data
			ctx.Set(KeyStream, *student.Stream)

			if student.Substream != nil {
				ctx.Set(KeySubstream, *student.Substream)
			}

			return next(ctx)
		}
	}
}

func (s *student) ValidateStudent() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			student, err := s.findStudent(ctx)
			if err != nil {
				return err
			}

			if student.Stream == nil {
				return ctx.Reply("Укажите группу с помощью команды /setstream")
			}

			return next(ctx)
		}
	}
}

func (s *student) findStudent(ctx telebot.Context) (models.Student, error) {
	id := ctx.Sender().ID

	student, err := s.repo.FindByID(context.Background(), id)
	if err != nil {
		return student, err
	}

	return student, nil
}

func (s *student) registerStudent(ctx telebot.Context) error {
	id := ctx.Sender().ID
	nickname := ctx.Sender().Username

	return s.repo.Create(context.Background(), id, nickname)
}

func (s *student) SendMainKeyboard(ctx telebot.Context) error {
	return ctx.Reply("main keyboard - TODO!")
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

		err := s.repo.UpdateStream(context.Background(), ctx.Callback().Sender.ID, stream)
		if err != nil {
			return err
		}

		if len(foundStream.Substreams) == 0 {
			_, err := s.bot.Edit(ctx.Callback().Message, fmt.Sprintf("Группа %s установлена!", foundStream.Name))
			return err
		}

		if len(foundStream.Substreams) == 1 {
			err := s.repo.UpdateSubstream(context.Background(), ctx.Callback().Sender.ID, foundStream.Substreams[0])
			if err != nil {
				return err
			}
			_, err = s.bot.Edit(ctx.Callback().Message, fmt.Sprintf("Подгруппа %s установлена!", foundStream.Substreams[0]))
			return err
		}

		r := s.bot.NewMarkup()

		btns := make([]telebot.Row, 0, len(streams))
		for _, substream := range foundStream.Substreams {
			b := r.Data(substream, actionSetSubstream, substream)
			btns = append(btns, r.Row(b))
		}
		r.Inline(btns...)

		_, err = s.bot.Edit(ctx.Callback().Message, "Выберите подгруппу:", r)
		return err
	})

	s.bot.Handle("\f"+actionSetSubstream, func(ctx telebot.Context) error {
		substream := ctx.Callback().Data

		err := s.repo.UpdateSubstream(context.Background(), ctx.Callback().Sender.ID, substream)
		if err != nil {
			return err
		}

		_, err = s.bot.Edit(ctx.Callback().Message, fmt.Sprintf("Подгруппа %s установлена!", substream))
		return err
	})

	return func(ctx telebot.Context) error {
		streams := s.portal.Streams()
		r := s.bot.NewMarkup()

		btns := make([]telebot.Row, 0, len(streams))
		for _, stream := range streams {
			b := r.Data(stream.Name, actionSetStream, stream.ID)
			btns = append(btns, r.Row(b))
		}
		r.Inline(btns...)

		return ctx.Reply("Выберите группу:", r)
	}
}
