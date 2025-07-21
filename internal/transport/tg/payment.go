package tg

import (
	"context"
	"fmt"
	"log"
	"pgtk-schedule/internal/models"

	"gopkg.in/telebot.v4"
)

type studentPayer interface {
	SetIsPaid(ctx context.Context, id int64) error
}

type payment struct {
	bot          *telebot.Bot
	paymentToken string
	studentPayer studentPayer
}

func NewPayment(bot *telebot.Bot, paymentToken string, studentPayer studentPayer) *payment {
	return &payment{
		bot:          bot,
		paymentToken: paymentToken,
		studentPayer: studentPayer,
	}
}

func (p *payment) OnPayment() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		log.Println("got payment query", ctx.Sender().ID)

		err := p.studentPayer.SetIsPaid(context.Background(), ctx.Sender().ID)
		if err != nil {
			_, _ = p.bot.Send(&telebot.User{ID: ctx.Sender().ID}, "Произошла ошибка при обработке платежа. Напишите администратору при помощи команды /feedback.")
			return err
		}

		return nil
	}
}

func (p *payment) OnCheckout() telebot.HandlerFunc {
	return func(ctx telebot.Context) error {
		pre := ctx.PreCheckoutQuery()
		if pre == nil || pre.ID == "" {
			log.Println("checkout is empty")
			return nil
		}

		log.Println("got checkout query", ctx.Sender().ID)

		err := p.bot.Accept(pre)
		if err != nil {
			_, _ = p.bot.Send(&telebot.User{ID: ctx.Sender().ID}, "Произошла ошибка при принятии платежа. Напишите администратору при помощи команды /feedback.")
			return err
		}

		return nil
	}
}

func (p *payment) Validate() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			isAdminRaw := ctx.Get(KeyIsAdmin)
			if isAdmin, ok := isAdminRaw.(bool); ok && isAdmin {
				return next(ctx)
			}

			if student, ok := ctx.Get(KeyStudent).(models.Student); ok && student.IsPaid {
				return next(ctx)
			}

			invoice := telebot.Invoice{
				Title:       "Подписка на весь учебный год",
				Description: "Подписка для доступа к боту на весь учебный год. Подписка будет обнулена в конце следующего учебного года (август 2026).",
				Currency:    "RUB",
				Prices:      []telebot.Price{{Label: "159 рублей", Amount: 15900}},
				Token:       p.paymentToken,
				Total:       15900,
				Payload:     fmt.Sprintf("%d", ctx.Sender().ID),
				NeedEmail:   true,
				SendEmail:   true,
				Data: `{
					"receipt": {
					"items": [
						{
							"description": "Подписка на весь учебный год",
							"quantity": "1.00",
							"amount": {
								"value": "159.00",
								"currency": "RUB"
							},
							"vat_code": 1
						}
					]
				}}`,
			}

			_, err := invoice.Send(p.bot, ctx.Sender(), nil)

			return err
		}
	}
}
