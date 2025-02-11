package bot

import (
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
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return err
	}

	// Api
	portal := portal.New()
	err = portal.Update()
	if err != nil {
		return err
	}

	// Database
	pool, err := database.NewPgx(cfg.DB_CONN)
	if err != nil {
		return err
	}

	// Repository
	studentRepo := repository.NewStudent(pool)

	// Service
	studentService := service.NewStudent(studentRepo)

	// Handlers
	studentHandlers := tg.NewStudent(bot, studentService, portal)

	bot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Reply("start command")
	}, studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())

	bot.Handle("/setstream", studentHandlers.SetStream(), studentHandlers.RegisteredStudent())

	bot.Handle("Получить расписание на неделю", func(ctx telebot.Context) error {
		return ctx.Reply("week lessons command")
	}, studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())

	bot.Start()

	return nil
}
