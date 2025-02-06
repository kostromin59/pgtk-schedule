package repository

import (
	"context"
	"errors"
	"pgtk-schedule/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type student struct {
	pool *pgxpool.Pool
}

func NewStudent(pool *pgxpool.Pool) *student {
	return &student{
		pool: pool,
	}
}

func (s *student) Create(ctx context.Context, id int64, nickname string) error {
	query := `INSERT INTO students(id, nickname) VALUES ($1, $2);`
	_, err := s.pool.Exec(ctx, query, id, nickname)
	return err
}

func (s *student) FindByID(ctx context.Context, id int64) (models.Student, error) {
	query := `SELECT nickname, stream, substream FROM students WHERE id = $1;`
	row := s.pool.QueryRow(ctx, query, id)
	student := models.Student{
		ID: id,
	}

	err := row.Scan(&student.Nickname, &student.Stream, student.Substream)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return student, models.ErrStudentNotFound
		}

		return student, err
	}

	return student, nil
}

func (s *student) FindAll(ctx context.Context, id int64, limit int) ([]models.Student, int64, error) {
	query := `SELECT id, nickname, stream, substream FROM students
	WHERE id > $1 ORDER BY id LIMIT $2`
	rows, err := s.pool.Query(ctx, query, id, limit)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	students, err := pgx.CollectRows(rows, pgx.RowToStructByName[models.Student])
	if err != nil {
		return nil, 0, err
	}

	lastSeenID := id
	if len(students) > 0 {
		lastSeenID = students[len(students)-1].ID
	}

	return students, lastSeenID, nil
}

func (s *student) UpdateStream(ctx context.Context, id int64, stream string) error {
	query := `UPDATE students SET stream = $1 WHERE id = $2;`
	rows, err := s.pool.Exec(ctx, query, stream, id)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != 1 {
		return models.ErrStudentNotFound
	}

	return nil
}

func (s *student) UpdateSubstream(ctx context.Context, id int64, substream string) error {
	query := `UPDATE students SET substream = $1 WHERE id = $2;`
	rows, err := s.pool.Exec(ctx, query, substream, id)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != 1 {
		return models.ErrStudentNotFound
	}

	return nil
}

func (s *student) UpdateNickname(ctx context.Context, id int64, nickname string) error {
	query := `UPDATE students SET nickname = $1 WHERE id = $2;`
	rows, err := s.pool.Exec(ctx, query, nickname, id)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != 1 {
		return models.ErrStudentNotFound
	}

	return nil
}
