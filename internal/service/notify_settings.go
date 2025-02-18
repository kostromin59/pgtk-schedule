package service

import (
	"context"
	"pgtk-schedule/internal/models"
)

type notifySettingsRepository interface {
	FindByStudentID(ctx context.Context, studentId int64) (models.NotifySettings, error)
	ToggleMorning(ctx context.Context, studentId int64) error
	ToggleEvening(ctx context.Context, studentId int64) error
	ToggleWeek(ctx context.Context, studentId int64) error
}

type notifySettings struct {
	repo notifySettingsRepository
}

func NewNotifySettings(repo notifySettingsRepository) *notifySettings {
	return &notifySettings{
		repo: repo,
	}
}

func (ns *notifySettings) FindByStudentID(ctx context.Context, studentId int64) (models.NotifySettings, error) {
	return ns.repo.FindByStudentID(ctx, studentId)
}

func (ns *notifySettings) ToggleMorning(ctx context.Context, studentId int64) error {
	return ns.repo.ToggleMorning(ctx, studentId)
}

func (ns *notifySettings) ToggleEvening(ctx context.Context, studentId int64) error {
	return ns.repo.ToggleEvening(ctx, studentId)
}

func (ns *notifySettings) ToggleWeek(ctx context.Context, studentId int64) error {
	return ns.repo.ToggleWeek(ctx, studentId)
}
