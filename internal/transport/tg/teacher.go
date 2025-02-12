package tg

import (
	"errors"
	"fmt"
	"pgtk-schedule/internal/models"
	"slices"
	"strings"

	"gopkg.in/telebot.v4"
)

const (
	actionFindTeacher = "findTeacher"
)

type teacherService interface {
	TodayList() ([]string, error)
	Find(teacher string) (models.Lesson, error)
}

type teacher struct {
	bot            *telebot.Bot
	teacherService teacherService
}

func NewTeacher(bot *telebot.Bot, teacherService teacherService) *teacher {
	return &teacher{
		bot:            bot,
		teacherService: teacherService,
	}
}

func (t *teacher) Find() telebot.HandlerFunc {
	t.bot.Handle("\f"+actionFindTeacher, func(ctx telebot.Context) error {
		teacher := ctx.Callback().Data

		lesson, err := t.teacherService.Find(teacher)
		if err != nil {
			if errors.Is(err, models.ErrLessonNotFound) {
				_, err = t.bot.Edit(ctx.Callback().Message, "Пара не найдена!")
				return err
			}

			return err
		}

		msg := fmt.Sprintf("<b>Ближайшая пара на сегодня у преподавателя %s:</b>\nПара: %s\nКабинет: %s\nВремя: %s-%s", lesson.Teacher, lesson.Name, lesson.Cabinet, lesson.DateStart.Format("15:04"), lesson.DateEnd.Format("15:04"))

		_, err = t.bot.Edit(ctx.Callback().Message, msg)
		return err
	})

	return func(ctx telebot.Context) error {
		teachers, err := t.teacherService.TodayList()
		if err != nil {
			return err
		}

		slices.Sort(teachers)

		r := t.bot.NewMarkup()

		fmt.Println(teachers)

		btns := make([]telebot.Row, 0, len(teachers))
		for _, teacher := range teachers {
			if teacher == "" {
				continue
			}

			splitted := strings.Split(teacher, " ")
			data := splitted[0]
			if len(splitted) > 1 {
				data += " "
				data += splitted[1]
			}
			fmt.Println(data)
			b := r.Data(teacher, actionFindTeacher, data)
			btns = append(btns, r.Row(b))
		}
		r.Inline(btns...)

		return ctx.Reply("Список преподавателей, у которых сегодня есть пары:", r)
	}
}
