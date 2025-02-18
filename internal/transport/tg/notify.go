package tg

import (
	"errors"
	"pgtk-schedule/internal/models"
	"time"

	"gopkg.in/telebot.v4"
)

type studentServiceForNotify interface {
	ForEach(func(student models.Student) error)
}

type scheduleServiceForNotify interface {
	CurrentWeekLessons(stream, substream string) ([]models.Lesson, error)
	TodayLessons(stream, substream string) ([]models.Lesson, error)
	TomorrowLessons(stream, substream string) ([]models.Lesson, error)
	LessonsToString(lessons []models.Lesson) string
}

type notify struct {
	bot             *telebot.Bot
	studentService  studentServiceForNotify
	scheduleService scheduleService
}

func NewNotify(bot *telebot.Bot, studentService studentServiceForNotify, scheduleService scheduleService) *notify {
	return &notify{
		bot:             bot,
		studentService:  studentService,
		scheduleService: scheduleService,
	}
}

func (n *notify) Morning() {
	n.studentService.ForEach(func(student models.Student) error {
		defer time.Sleep(300 * time.Millisecond)
		if err := n.validate(student); err != nil {
			return err
		}

		substream := ""
		if student.Substream != nil {
			substream = *student.Substream
		}

		_, err := n.scheduleService.CurrentWeekLessons(*student.Stream, substream)
		if err != nil {
			if errors.Is(err, models.ErrLessonsAreEmpty) {
				return nil
			}
			return err
		}

		lessons, err := n.scheduleService.TodayLessons(*student.Stream, substream)
		var msg string
		if err != nil {
			if errors.Is(err, models.ErrLessonsAreEmpty) {
				msg = "Сегодня нет пар! Хорошего дня!"
			} else {
				return err
			}
		} else {
			msg = "<b>Присылаю пары на сегодня. Расписание может измениться в любой момент, не забывай обновлять его!:</b>\n\n" + n.scheduleService.LessonsToString(lessons)
		}

		_, err = n.bot.Send(&telebot.User{ID: student.ID}, msg)
		return err
	})
}

func (n *notify) Evening() {
	n.studentService.ForEach(func(student models.Student) error {
		defer time.Sleep(300 * time.Millisecond)
		if err := n.validate(student); err != nil {
			return err
		}

		substream := ""
		if student.Substream != nil {
			substream = *student.Substream
		}

		lessons, err := n.scheduleService.TomorrowLessons(*student.Stream, substream)
		if err != nil {
			if errors.Is(err, models.ErrLessonsAreEmpty) {
				return nil
			}
			return err
		}

		msg := "<b>Присылаю пары на завтра. Расписание может измениться в любой момент, не забывай обновлять его!:</b>\n\n" + n.scheduleService.LessonsToString(lessons)

		_, err = n.bot.Send(&telebot.User{ID: student.ID}, msg)
		return err
	})
}

func (n *notify) Week() {
	n.studentService.ForEach(func(student models.Student) error {
		defer time.Sleep(300 * time.Millisecond)
		if err := n.validate(student); err != nil {
			return err
		}

		substream := ""
		if student.Substream != nil {
			substream = *student.Substream
		}

		lessons, err := n.scheduleService.CurrentWeekLessons(*student.Stream, substream)
		if err != nil {
			if errors.Is(err, models.ErrLessonsAreEmpty) {
				return nil
			}
			return err
		}

		msg := "<b>Пары на следующую неделю. Расписание может измениться в любой момент, не забывай обновлять его!</b>\n\n" + n.scheduleService.LessonsToString(lessons)

		_, err = n.bot.Send(&telebot.User{ID: student.ID}, msg)
		return err
	})
}

func (*notify) validate(student models.Student) error {
	if student.ID == 0 {
		return models.ErrStudentNotFound
	}

	if student.Stream == nil {
		return models.ErrStudentStreamMissed
	}

	return nil
}
