//go:build integration

package repository

import (
	"os"
	"pgtk-schedule/internal/models"
	"pgtk-schedule/internal/repository"
	"pgtk-schedule/pkg/database"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStudentNotifySettings(t *testing.T) {
	dbConn := os.Getenv("DB_CONN")
	if dbConn == "" {
		t.Skip("DB_CONN not found")
	}

	pool, err := database.NewPgx(dbConn)
	require.NoError(t, err)

	_, err = pool.Exec(t.Context(), "DELETE FROM students")
	require.NoError(t, err)

	studentRepo := repository.NewStudent(pool)
	notifySettingsRepo := repository.NewNotifySettings(pool)

	t.Run("create student", func(t *testing.T) {
		err := studentRepo.Create(t.Context(), 1, "test")
		require.NoError(t, err)
		err = studentRepo.Create(t.Context(), 2, "test2")
		require.NoError(t, err)
	})

	t.Run("notify settings are created", func(t *testing.T) {
		notifySettings, err := notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)

		assert.Equal(t, int64(1), notifySettings.StudentID)
		assert.True(t, notifySettings.Morning)
		assert.True(t, notifySettings.Evening)
		assert.True(t, notifySettings.Week)
	})

	t.Run("find student by id", func(t *testing.T) {
		student, err := studentRepo.FindByID(t.Context(), 1)
		require.NoError(t, err)
		nickname := "test"
		assert.Equal(t, models.Student{
			ID:       1,
			Nickname: &nickname,
		}, student)
	})

	t.Run("student not found", func(t *testing.T) {
		_, err := studentRepo.FindByID(t.Context(), -1)
		assert.ErrorIs(t, err, models.ErrStudentNotFound)
	})

	t.Run("find all with limit", func(t *testing.T) {
		students, lastId, err := studentRepo.FindAll(t.Context(), 0, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(1), lastId)
		assert.Len(t, students, 1)

		students, lastId, err = studentRepo.FindAll(t.Context(), lastId, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), lastId)
		assert.Len(t, students, 1)

		students, lastId, err = studentRepo.FindAll(t.Context(), lastId, 1)
		require.NoError(t, err)
		assert.Equal(t, int64(2), lastId)
		assert.Len(t, students, 0)
	})

	t.Run("update stream", func(t *testing.T) {
		err := studentRepo.UpdateStream(t.Context(), 1, "updated")
		require.NoError(t, err)

		student, err := studentRepo.FindByID(t.Context(), 1)
		require.NoError(t, err)
		require.NotNil(t, student.Stream)
		assert.Equal(t, "updated", *student.Stream)
	})

	t.Run("update substream", func(t *testing.T) {
		err := studentRepo.UpdateSubstream(t.Context(), 1, "updated")
		require.NoError(t, err)

		student, err := studentRepo.FindByID(t.Context(), 1)
		require.NoError(t, err)
		require.NotNil(t, student.Substream)
		assert.Equal(t, "updated", *student.Substream)
	})

	t.Run("update nickname", func(t *testing.T) {
		err := studentRepo.UpdateNickname(t.Context(), 1, "updated")
		require.NoError(t, err)

		student, err := studentRepo.FindByID(t.Context(), 1)
		require.NoError(t, err)
		require.NotNil(t, student.Nickname)
		assert.Equal(t, "updated", *student.Nickname)
	})

	t.Run("notify settings toggle morning", func(t *testing.T) {
		err := notifySettingsRepo.ToggleMorning(t.Context(), 1)
		require.NoError(t, err)

		notifySettings, err := notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, false, notifySettings.Morning)

		err = notifySettingsRepo.ToggleMorning(t.Context(), 1)
		require.NoError(t, err)

		notifySettings, err = notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, true, notifySettings.Morning)
	})

	t.Run("notify settings toggle evening", func(t *testing.T) {
		err := notifySettingsRepo.ToggleEvening(t.Context(), 1)
		require.NoError(t, err)

		notifySettings, err := notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, false, notifySettings.Evening)

		err = notifySettingsRepo.ToggleEvening(t.Context(), 1)
		require.NoError(t, err)

		notifySettings, err = notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, true, notifySettings.Evening)
	})

	t.Run("notify settings toggle week", func(t *testing.T) {
		err := notifySettingsRepo.ToggleWeek(t.Context(), 1)
		require.NoError(t, err)

		notifySettings, err := notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, false, notifySettings.Week)

		err = notifySettingsRepo.ToggleWeek(t.Context(), 1)
		require.NoError(t, err)

		notifySettings, err = notifySettingsRepo.FindByStudentID(t.Context(), 1)
		require.NoError(t, err)
		assert.Equal(t, true, notifySettings.Week)
	})
}
