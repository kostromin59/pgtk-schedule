package tg

import (
	"context"
	"errors"
	"pgtk-schedule/internal/models"
	"time"

	"gopkg.in/telebot.v4"
)

const (
	actionToggleMorning = "toggleMorning"
	actionToggleEvening = "toggleEvening"
	actionToggleWeek    = "toggleWeek"
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

type notifySettingsService interface {
	FindByStudentID(ctx context.Context, studentId int64) (models.NotifySettings, error)
	ToggleMorning(ctx context.Context, studentId int64) error
	ToggleEvening(ctx context.Context, studentId int64) error
	ToggleWeek(ctx context.Context, studentId int64) error
}

type notify struct {
	bot                   *telebot.Bot
	studentService        studentServiceForNotify
	scheduleService       scheduleServiceForNotify
	notifySettingsSerivce notifySettingsService
}

func NewNotify(bot *telebot.Bot, studentService studentServiceForNotify, scheduleService scheduleService, notifySettingsService notifySettingsService) *notify {
	return &notify{
		bot:                   bot,
		studentService:        studentService,
		scheduleService:       scheduleService,
		notifySettingsSerivce: notifySettingsService,
	}
}

func (n *notify) Change() telebot.HandlerFunc {
	n.bot.Handle("\f"+actionToggleMorning, func(ctx telebot.Context) error {
		err := n.notifySettingsSerivce.ToggleMorning(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		markup := n.buildMarkup(settings)

		_, err = n.bot.Edit(ctx.Callback().Message, "Изменение настроек уведомлений:", markup)
		return err
	})

	n.bot.Handle("\f"+actionToggleEvening, func(ctx telebot.Context) error {
		err := n.notifySettingsSerivce.ToggleEvening(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		markup := n.buildMarkup(settings)

		_, err = n.bot.Edit(ctx.Callback().Message, "Изменение настроек уведомлений:", markup)
		return err
	})

	n.bot.Handle("\f"+actionToggleWeek, func(ctx telebot.Context) error {
		err := n.notifySettingsSerivce.ToggleWeek(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		markup := n.buildMarkup(settings)

		_, err = n.bot.Edit(ctx.Callback().Message, "Изменение настроек уведомлений:", markup)
		return err
	})

	return func(ctx telebot.Context) error {
		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), ctx.Sender().ID)
		if err != nil {
			return err
		}

		markup := n.buildMarkup(settings)

		return ctx.Reply("Изменение настроек уведомлений:", markup)
	}
}

func (n *notify) buildMarkup(settings models.NotifySettings) *telebot.ReplyMarkup {
	state := [...]struct {
		Text   string
		Action string
		State  bool
	}{
		{Text: "утренние уведомления", Action: actionToggleMorning, State: settings.Morning},
		{Text: "вечерние уведомления", Action: actionToggleEvening, State: settings.Evening},
		{Text: "недельные уведомления", Action: actionToggleWeek, State: settings.Week},
	}

	markup := n.bot.NewMarkup()

	btns := make([]telebot.Row, 0, len(state))
	for _, s := range state {
		var text string
		if s.State {
			text = "Выключить "
		} else {
			text = "Включить "
		}

		text += s.Text

		b := markup.Data(text, s.Action)
		btns = append(btns, markup.Row(b))
	}

	markup.Inline(btns...)

	return markup
}

func (n *notify) Morning() {
	n.studentService.ForEach(func(student models.Student) error {
		defer time.Sleep(300 * time.Millisecond)
		if err := n.validate(student); err != nil {
			return err
		}

		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), student.ID)
		if err != nil {
			return err
		}
		if !settings.Morning {
			return nil
		}

		substream := ""
		if student.Substream != nil {
			substream = *student.Substream
		}

		_, err = n.scheduleService.CurrentWeekLessons(*student.Stream, substream)
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
			msg = "<b>Присылаю пары на сегодня. Расписание может измениться в любой момент, не забывай обновлять его!</b>\n\n" + n.scheduleService.LessonsToString(lessons)
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

		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), student.ID)
		if err != nil {
			return err
		}
		if !settings.Evening {
			return nil
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

		msg := "<b>Присылаю пары на завтра. Расписание может измениться в любой момент, не забывай обновлять его!</b>\n\n" + n.scheduleService.LessonsToString(lessons)

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

		settings, err := n.notifySettingsSerivce.FindByStudentID(context.Background(), student.ID)
		if err != nil {
			return err
		}
		if !settings.Week {
			return nil
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
