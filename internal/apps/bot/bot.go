package bot

import (
	"log"
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
	adminHandlers := tg.NewAdmin(bot, studentService, cfg.AdminID)

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
			Text:        "/feedback",
			Description: "Связаться с разработчиком",
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

	bot.Handle("/send", adminHandlers.Send(), adminHandlers.ValidateAdmin())

	bot.Handle(&weekButton, scheduleHandlers.CurrentWeekLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&todayButton, scheduleHandlers.TodayLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())
	bot.Handle(&tomorrowButton, scheduleHandlers.TomorrowLessons(), studentHandlers.RegisteredStudent(), studentHandlers.ValidateStudent())

	// Cron jobs
	s, err := gocron.NewScheduler()
	if err != nil {
		return err
	}

	s.NewJob(gocron.CronJob("TZ=Asia/Yekaterinburg 0 5 * * 1-6", false), gocron.NewTask(func() {
		studentService.ForEach(func(student models.Student) error {
			defer time.Sleep(300 * time.Millisecond)
			if err := studentService.Validate(student); err != nil {
				return err
			}

			substream := ""
			if student.Substream != nil {
				substream = *student.Substream
			}

			weekLessons, err := scheduleService.CurrentWeekLessons(*student.Stream, substream)
			if err != nil {
				return err
			}

			if len(weekLessons) == 0 {
				return nil
			}

			lessons, err := scheduleService.TodayLessons(*student.Stream, substream)
			if err != nil {
				return err
			}

			msg := "Сегодня нет пар! Хорошего дня!"
			if len(lessons) != 0 {
				msg = "<b>Присылаю пары на сегодня. Расписание может измениться в любой момент, не забывай обновлять его!:</b>\n\n" + scheduleService.LessonsToString(lessons)
			}

			_, err = bot.Send(&telebot.User{ID: student.ID}, msg)
			return err
		})
	}))

	s.NewJob(gocron.CronJob("TZ=Asia/Yekaterinburg 0 18 * * 1-6", false), gocron.NewTask(func() {
		studentService.ForEach(func(student models.Student) error {
			defer time.Sleep(300 * time.Millisecond)
			if err := studentService.Validate(student); err != nil {
				return err
			}

			substream := ""
			if student.Substream != nil {
				substream = *student.Substream
			}

			lessons, err := scheduleService.TomorrowLessons(*student.Stream, substream)
			if err != nil {
				return err
			}

			if len(lessons) == 0 {
				return nil
			}

			msg := "<b>Присылаю пары на завтра. Расписание может измениться в любой момент, не забывай обновлять его!:</b>\n\n" + scheduleService.LessonsToString(lessons)

			_, err = bot.Send(&telebot.User{ID: student.ID}, msg)
			return err
		})
	}))

	s.NewJob(gocron.CronJob("TZ=Asia/Yekaterinburg 0 12 * * 0", false), gocron.NewTask(func() {
		studentService.ForEach(func(student models.Student) error {
			defer time.Sleep(300 * time.Millisecond)
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

			msg := "<b>Пары на следующую неделю. Расписание может измениться в любой момент, не забывай обновлять его!</b>\n\n" + scheduleService.LessonsToString(lessons)

			_, err = bot.Send(&telebot.User{ID: student.ID}, msg)
			return err
		})
	}))

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
