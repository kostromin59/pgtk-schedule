package tg

import (
	"context"
	"errors"
	"fmt"
	"pgtk-schedule/internal/models"

	"gopkg.in/telebot.v4"
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
			student, err := s.validateStudent(ctx)
			if err == nil {
				ctx.Set(KeyStream, *student.Stream)

				if student.Substream != nil {
					ctx.Set(KeySubstream, *student.Substream)
				}

				return next(ctx)
			}

			if errors.Is(err, models.ErrStreamIsUnknown) {
				return s.fillStream(ctx)
			}

			if errors.Is(err, models.ErrStudentNotFound) {
				if err := s.registerStudent(ctx); err != nil {
					return err
				}
			}

			return err
		}
	}
}

func (s *student) validateStudent(ctx telebot.Context) (models.Student, error) {
	id := ctx.Sender().ID

	student, err := s.repo.FindByID(context.Background(), id)
	if err != nil {
		return student, err
	}

	if student.Stream == nil {
		return student, models.ErrStreamIsUnknown
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

func (s *student) fillStream(ctx telebot.Context) error {
	streams := s.portal.Streams()
	// r := s.bot.NewMarkup()
	// b := r.Data("test", "set test", "1")
	// r.Inline(r.Row(b))

	// s.bot.Handle(&b, func(ctx telebot.Context) error {
	// 	fmt.Println("fff")
	// 	return ctx.Respond()
	// })

	return ctx.Reply(fmt.Sprintf("there are %d streams in portal", len(streams)))
}
