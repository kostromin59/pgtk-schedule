package tg

import (
	"pgtk-schedule/internal/models"
	"strings"
	"time"

	"gopkg.in/telebot.v4"
)

type adminStudentService interface {
	ForEach(func(models.Student) error)
}

type admin struct {
	bot            *telebot.Bot
	studentService adminStudentService
	adminId        int64
}

func NewAdmin(bot *telebot.Bot, studentService adminStudentService, adminId int64) *admin {
	return &admin{
		bot:            bot,
		studentService: studentService,
		adminId:        adminId,
	}
}

func (a *admin) Send() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		msg := strings.Join(ctx.Args(), " ")

		a.studentService.ForEach(func(student models.Student) error {
			defer time.Sleep(300 * time.Millisecond)
			_, err := a.bot.Send(&telebot.User{ID: student.ID}, msg)
			return err
		})

		return nil
	}
}

func (a *admin) ValidateAdmin() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			if ctx.Sender().ID != a.adminId {
				return nil
			}

			return next(ctx)
		}
	}
}
