package bot

import (
	"log"
	"pgtk-schedule/configs"
	"pgtk-schedule/internal/repository"
	"pgtk-schedule/internal/transport/tg"

	"gopkg.in/telebot.v4"
)

func Run(cfg configs.Bot) error {
	pref := telebot.Settings{
		Token: cfg.TgBotToken,
		OnError: func(err error, ctx telebot.Context) {
			log.Println(err.Error(), ctx.Sender().ID)
			ctx.Reply("Что-то пошло не так!")
		},
	}

	bot, err := telebot.NewBot(pref)
	if err != nil {
		return err
	}

	studentRepo := repository.NewStudent(nil)

	studentHandlers := tg.NewStudent(studentRepo)

	bot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Reply("start command")
	}, studentHandlers.RegisteredStudent())

	bot.Handle("Получить расписание на неделю", func(ctx telebot.Context) error {
		return ctx.Reply("week lessons command")
	}, studentHandlers.RegisteredStudent())

	bot.Start()

	return nil
}
