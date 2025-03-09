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

	"github.com/go-co-op/gocron/v2"
	"gopkg.in/telebot.v4"
)

func Run(cfg configs.Bot) error {
	// Bot
	pref := telebot.Settings{
		Token: cfg.BotToken,
		OnError: func(err error, ctx telebot.Context) {
			log.Println(err.Error(), ctx.Sender().ID)
			ctx.Reply("Что-то пошло не так! Напишите @kostromin59 о проблеме.")
		},
		Poller: &telebot.LongPoller{
			Timeout:        3 * time.Second,
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
	pool, err := database.NewPgx(cfg.DbConn)
	if err != nil {
		return err
	}

	// Repository
	studentRepo := repository.NewStudent(pool)
	notifySettingsRepo := repository.NewNotifySettings(pool)

	// Service
	studentService := service.NewStudent(studentRepo)
	scheduleService := service.NewSchedule(portal)
	teacherService := service.NewTeacher(portal, scheduleService)
	notifySettingsService := service.NewNotifySettings(notifySettingsRepo)

	// Handlers
	studentHandlers := tg.NewStudent(bot, studentService, portal)
	scheduleHandlers := tg.NewSchedule(scheduleService)
	teacherHandlers := tg.NewTeacher(bot, teacherService)
	adminHandlers := tg.NewAdmin(bot, studentService, cfg.AdminID)
	notifyHandlers := tg.NewNotify(bot, studentService, scheduleService, notifySettingsService)

	if err := scheduleService.Update(); err != nil {
		return err
	}

	err = bot.SetCommands([]telebot.Command{
		{
			Text:        "/setstream",
			Description: "Измененить группу и подгруппу",
		},
		{
			Text:        "/findteacher",
			Description: "Найти преподавателя",
		},
		{
			Text:        "/notifysettings",
			Description: "Изменить настройки уведомлений",
		},
		{
			Text:        "/feedback",
			Description: "Связаться с разработчиком",
		},
	})
	if err != nil {
		return err
	}

	markup := bot.NewMarkup()
	weekButton := markup.Text("Получить расписание на неделю")
	todayButton := markup.Text("На сегодня")
	tomorrowButton := markup.Text("На завтра")
	markup.ResizeKeyboard = true
	markup.Reply(telebot.Row{weekButton}, telebot.Row{todayButton, tomorrowButton})

	bot.Handle("/start", func(ctx telebot.Context) error {
		return ctx.Reply("Привет! Вышло обновление бота. Со следующего учебного года поддержка бота будет платной, потому что никто из студентов не хочет поддерживать бота. Необходимо будет оплачивать сервер каждый месяц. Подробнее можно спросить у @kostromin59.\n\nИспользуйте команду /feedback для обратной связи.", markup)
	})
	bot.Handle("/setstream", studentHandlers.SetStream(), studentHandlers.RegisteredStudent())
	bot.Handle("/findteacher", teacherHandlers.Find())
	bot.Handle("/send", adminHandlers.Send(), adminHandlers.ValidateAdmin())
	bot.Handle("/notifysettings", notifyHandlers.Change(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle("/feedback", func(ctx telebot.Context) error {
		return ctx.Reply("Напишите @kostromin59, чтобы сообщить о проблеме, предложить новый функционал или договориться о дальнейшей поддержке бота")
	})

	bot.Handle(&weekButton, scheduleHandlers.CurrentWeekLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&todayButton, scheduleHandlers.TodayLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&tomorrowButton, scheduleHandlers.TomorrowLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())

	// Cron jobs
	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	s.NewJob(gocron.CronJob("TZ=Asia/Yekaterinburg 0 5 * * 1-6", false), gocron.NewTask(notifyHandlers.Morning))
	s.NewJob(gocron.CronJob("TZ=Asia/Yekaterinburg 0 18 * * 1-5", false), gocron.NewTask(notifyHandlers.Evening))
	s.NewJob(gocron.CronJob("TZ=Asia/Yekaterinburg 0 12 * * 0", false), gocron.NewTask(notifyHandlers.Week))

	s.NewJob(gocron.CronJob("0 * * * *", false), gocron.NewTask(func() {
		if err := scheduleService.Update(); err != nil {
			log.Println(err.Error())
		}
		log.Println("schedule has been updated!")
	}))

	s.Start()
	defer s.Shutdown()

	bot.Start()

	return nil
}
