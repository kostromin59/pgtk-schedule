package middleware

import (
	"context"
	"pgtk-schedule/internal/models"
	"pgtk-schedule/internal/transport/tg"

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

func (s *student) ValidateStudent() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			id := ctx.Sender().ID

			student, err := s.repo.FindByID(context.Background(), id)
			if err != nil {
				return err
			}

			if student.Stream == nil {
				return models.ErrStreamIsUnknown
			}
			ctx.Set(tg.KeyStream, *student.Stream)

			if student.Substream != nil {
				ctx.Set(tg.KeySubstream, *student.Substream)
			}

			return next(ctx)
		}
	}
}
