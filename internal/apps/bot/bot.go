package bot

import (
	"context"
	"fmt"
	"log"
	"math"
	"pgtk-schedule/configs"
	"pgtk-schedule/internal/api/portal"
	"pgtk-schedule/internal/models"
	"pgtk-schedule/internal/repository"
	"pgtk-schedule/internal/service"
	"pgtk-schedule/internal/transport/tg"
	"pgtk-schedule/pkg/database"
	"time"

	"github.com/go-co-op/gocron/v2"
	"gopkg.in/telebot.v4"
)

func Run(cfg configs.Bot) error {
	// Bot
	pref := telebot.Settings{
		Token: cfg.BotToken,
		OnError: func(err error, ctx telebot.Context) {
			log.Println(err.Error(), ctx.Sender().ID)
			ctx.Reply("Что-то пошло не так!")
		},
		Poller: &telebot.LongPoller{
			Timeout:        10 * time.Second,
			AllowedUpdates: []string{"message", "chat_member", "callback_query", "poll", "inline_query"},
		},
		ParseMode: telebot.ModeHTML,
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return err
	}

	// Api
	portal := portal.New()

	// Database
	pool, err := database.NewPgx(cfg.DB_CONN)
	if err != nil {
		return err
	}

	// Repository
	studentRepo := repository.NewStudent(pool)

	// Service
	studentService := service.NewStudent(studentRepo)
	scheduleService := service.NewSchedule(portal)
	teacherService := service.NewTeacher(portal, scheduleService)

	// Handlers
	studentHandlers := tg.NewStudent(bot, studentService, portal)
	scheduleHandlers := tg.NewSchedule(scheduleService)
	teacherHandlers := tg.NewTeacher(bot, teacherService)

	scheduleService.RunUpdater(context.Background(), 1*time.Hour)

	err = bot.SetCommands([]telebot.Command{
		{
			Text:        "/setstream",
			Description: "Изменение группы и подгруппы",
		},
		{
			Text:        "/findteacher",
			Description: "Найти преподавателя",
		},
		{
			Text:        "/feedback",
			Description: "Найти преподавателя",
		},
	})
	if err != nil {
		return err
	}

	r := bot.NewMarkup()
	weekButton := r.Text("Получить расписание на неделю")
	todayButton := r.Text("На сегодня")
	tomorrowButton := r.Text("На завтра")
	r.ResizeKeyboard = true

	r.Reply(telebot.Row{weekButton}, telebot.Row{todayButton, tomorrowButton})

	bot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Reply("Привет! Вышло обновление бота. Со следующего учебного года поддержка бота будет платной, потому что никто из студентов не хочет поддерживать бота. Необходимо будет оплачивать сервер каждый месяц. Подробнее можно спросить у @kostromin59.\n\nИспользуйте команду /feedback для обратной связи.", r)
	})
	bot.Handle("/setstream", studentHandlers.SetStream(), studentHandlers.RegisteredStudent())
	bot.Handle("/findteacher", teacherHandlers.Find())
	bot.Handle("/feedback", func(ctx telebot.Context) error {
		return ctx.Reply("Напишите @kostromin59, чтобы сообщить о проблеме, предложить новый функционал или договориться о дальнейшей поддержке бота")
	})

	// TODO: move to handlers
	// TODO: take message from args
	// TODO: add admin validation
	bot.Handle("/send", func(ctx telebot.Context) error {
		var lastId int64 = math.MinInt64
		for {
			students, currentLastId, err := studentService.FindAll(context.Background(), lastId, 25)
			if err != nil {
				log.Println(err.Error())
				ctx.Reply(fmt.Sprintf("Ошибка в сервисе: %s", err.Error()))
				return err
			}

			if currentLastId == lastId {
				return nil
			}

			lastId = currentLastId

			for _, student := range students {
				_, err := bot.Send(&telebot.User{ID: student.ID}, "test")
				if err != nil {
					log.Println(err.Error())
				}
				time.Sleep(300 * time.Millisecond)
			}
		}
	})

	bot.Handle(&weekButton, scheduleHandlers.CurrentWeekLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&todayButton, scheduleHandlers.TodayLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&tomorrowButton, scheduleHandlers.TomorrowLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())

	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	s.NewJob(gocron.CronJob("0 5 * * 1-6", false), gocron.NewTask(func() {
		studentHandlers.ForEachStudent(func(bot *telebot.Bot, student models.Student) error {
			if err := studentService.Validate(student); err != nil {
				return err
			}

			substream := ""
			if student.Substream != nil {
				substream = *student.Substream
			}

			lessons, err := scheduleService.CurrentWeekLessons(*student.Stream, substream)
			if err != nil {
				return err
			}

			msg := "Сегодня нет пар! Хорошего дня!"
			if len(lessons) != 0 {
				msg = fmt.Sprintf("<b>У тебя сегодня %d пар:</b>\n", len(lessons)) + scheduleService.LessonsToString(lessons)
			}

			_, err = bot.Send(&telebot.User{ID: student.ID}, msg)
			return err
		})
	}))

	s.NewJob(gocron.CronJob("0 12 * * 0", false), gocron.NewTask(func() {
		studentHandlers.ForEachStudent(func(bot *telebot.Bot, student models.Student) error {
			if err := studentService.Validate(student); err != nil {
				return err
			}

			substream := ""
			if student.Substream != nil {
				substream = *student.Substream
			}

			lessons, err := scheduleService.CurrentWeekLessons(*student.Stream, substream)
			if err != nil {
				return err
			}

			if len(lessons) == 0 {
				return nil
			}

			msg := "<b>Пары на следующую неделю:</b>\n" + scheduleService.LessonsToString(lessons)

			_, err = bot.Send(&telebot.User{ID: student.ID}, msg)
			return err
		})
	}))

	s.Start()
	defer s.Shutdown()

	bot.Start()

	return nil
}
