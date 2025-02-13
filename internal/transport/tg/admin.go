package tg

import (
	"log"
	"pgtk-schedule/internal/models"
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

// TODO: take message from args
func (a *admin) Send() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		a.studentService.ForEach(func(student models.Student) error {
			defer time.Sleep(300 * time.Second)
			_, err := a.bot.Send(&telebot.User{ID: student.ID}, "test")
			if err != nil {
				log.Println(err.Error())
			}
			return nil
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
