package repository

import (
	"context"
	"pgtk-schedule/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type notifySettings struct {
	pool *pgxpool.Pool
}

func NewNotifySettings(pool *pgxpool.Pool) *notifySettings {
	return &notifySettings{
		pool: pool,
	}
}

func (ns *notifySettings) FindByStudentID(ctx context.Context, studentId int64) (models.NotifySettings, error) {
	query := `SELECT id, morning, evening, week FROM notify_settings WHERE student_id = $1`
	row := ns.pool.QueryRow(ctx, query, studentId)
	notifySettings := models.NotifySettings{
		StudentID: studentId,
	}

	err := row.Scan(&notifySettings.ID, &notifySettings.Morning, &notifySettings.Evening, &notifySettings.Week)
	return notifySettings, err
}

func (ns *notifySettings) ToggleMorning(ctx context.Context, studentId int64) error {
	query := `UPDATE notify_settings SET morning = NOT morning WHERE student_id = $1`
	rows, err := ns.pool.Exec(ctx, query, studentId)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != 1 {
		return models.ErrNotifySettingsNotFound
	}

	return nil
}

func (ns *notifySettings) ToggleEvening(ctx context.Context, studentId int64) error {
	query := `UPDATE notify_settings SET evening = NOT evening WHERE student_id = $1`
	rows, err := ns.pool.Exec(ctx, query, studentId)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != 1 {
		return models.ErrNotifySettingsNotFound
	}

	return nil
}

func (ns *notifySettings) ToggleWeek(ctx context.Context, studentId int64) error {
	query := `UPDATE notify_settings SET week = NOT week WHERE student_id = $1`
	rows, err := ns.pool.Exec(ctx, query, studentId)
	if err != nil {
		return err
	}

	if rows.RowsAffected() != 1 {
		return models.ErrNotifySettingsNotFound
	}

	return nil
}
