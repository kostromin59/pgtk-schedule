package service

import (
	"context"
	"pgtk-schedule/internal/models"
)

type studentRepository interface {
	Create(ctx context.Context, id int64, nickname string) error
	FindByID(ctx context.Context, id int64) (models.Student, error)
	UpdateStream(ctx context.Context, id int64, stream string) error
	UpdateSubstream(ctx context.Context, id int64, substream string) error
	UpdateNickname(ctx context.Context, id int64, nickname string) error
	FindAll(ctx context.Context, id int64, limit int) ([]models.Student, int64, error)
}

type student struct {
	repo studentRepository
}

func NewStudent(repo studentRepository) *student {
	return &student{
		repo: repo,
	}
}

func (s *student) Validate(student models.Student) error {
	if student.ID == 0 {
		return models.ErrStudentNotFound
	}

	if student.Stream == nil {
		return models.ErrStudentStreamMissed
	}

	return nil
}

func (s *student) Create(ctx context.Context, id int64, nickname string) error {
	return s.repo.Create(ctx, id, nickname)
}

func (s *student) FindByID(ctx context.Context, id int64) (models.Student, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *student) UpdateStream(ctx context.Context, id int64, stream string) error {
	return s.repo.UpdateStream(ctx, id, stream)
}

func (s *student) UpdateSubstream(ctx context.Context, id int64, substream string) error {
	return s.repo.UpdateSubstream(ctx, id, substream)
}

func (s *student) UpdateNickname(ctx context.Context, id int64, nickname string) error {
	return s.repo.UpdateNickname(ctx, id, nickname)
}

func (s *student) FindAll(ctx context.Context, id int64, limit int) ([]models.Student, int64, error) {
	return s.repo.FindAll(ctx, id, limit)
}
