package tg

import (
	"context"
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

type student struct {
	repo studentRepository
}

func NewStudent(repo studentRepository) *student {
	return &student{
		repo: repo,
	}
}

func (s *student) RegisteredStudent() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			if err := s.validateStudent(ctx); err == nil {
				return next(ctx)
			}

			if err := s.registerStudent(ctx); err != nil {
				return err
			}

			return next(ctx)
		}
	}
}

func (s *student) validateStudent(ctx telebot.Context) error {
	id := ctx.Sender().ID

	student, err := s.repo.FindByID(context.Background(), id)
	if err != nil {
		return err
	}

	if student.Stream == nil {
		return models.ErrStreamIsUnknown
	}
	ctx.Set(KeyStream, *student.Stream)

	if student.Substream != nil {
		ctx.Set(KeySubstream, *student.Substream)
	}

	return nil
}

func (s *student) registerStudent(ctx telebot.Context) error {
	id := ctx.Sender().ID
	nickname := ctx.Sender().Username

	return s.repo.Create(context.Background(), id, nickname)
}

func (s *student) sendMainKeyboard(ctx telebot.Context) error {
	return ctx.Reply("main keyboard - TODO!")
}
