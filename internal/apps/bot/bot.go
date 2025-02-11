package bot

import (
	"context"
	"log"
	"pgtk-schedule/configs"
	"pgtk-schedule/internal/api/portal"
	"pgtk-schedule/internal/repository"
	"pgtk-schedule/internal/service"
	"pgtk-schedule/internal/transport/tg"
	"pgtk-schedule/pkg/database"
	"time"

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

	// Handlers
	studentHandlers := tg.NewStudent(bot, studentService, portal)
	scheduleHandlers := tg.NewSchedule(scheduleService)

	scheduleService.RunUpdater(context.Background(), 1*time.Hour)

	err = bot.SetCommands([]telebot.Command{{
		Text:        "/setstream",
		Description: "Изменение группы и подгруппы",
	}})
	if err != nil {
		return err
	}

	r := bot.NewMarkup()
	weekButton := r.Text("Получить расписание на неделю")
	todayButton := r.Text("На сегодня")
	tomorrowButton := r.Text("На завтра")

	r.Reply(telebot.Row{weekButton}, telebot.Row{todayButton, tomorrowButton})

	bot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Reply("Привет! Вышло обновление бота. Со следующего учебного года поддержка бота будет платной, потому что никто из студентов не хочет поддерживать бота. Необходимо будет оплачивать сервер каждый месяц. Подробнее можно спросить у @kostromin59.", r)
	})
	bot.Handle("/setstream", studentHandlers.SetStream(), studentHandlers.RegisteredStudent())
	bot.Handle("/send", func(ctx telebot.Context) error {
		return nil
	})

	bot.Handle(&weekButton, scheduleHandlers.CurrentWeekLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&todayButton, scheduleHandlers.TodayLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&tomorrowButton, scheduleHandlers.TomorrowLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())

	bot.Start()

	return nil
}
