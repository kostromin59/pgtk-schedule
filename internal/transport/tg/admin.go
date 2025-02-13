package tg

import "gopkg.in/telebot.v4"

type admin struct {
	adminId int64
}

func NewAdmin(adminId int64) *admin {
	return &admin{
		adminId: adminId,
	}
}

func (a *admin) ValidateAdmin() telebot.MiddlewareFunc {
	return func(next telebot.HandlerFunc) telebot.HandlerFunc {
		return func(ctx telebot.Context) error {
			if ctx.Sender().ID != a.adminId {
				return nil
			}

			return next(ctx)
		}
	}
}
