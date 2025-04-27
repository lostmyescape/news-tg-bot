package middleware

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lostmyescape/news-tg-bot/internal/botkit"
)

// AdminOnly middleware provides to commands for admins only
func AdminOnly(adminsID []int64, next botkit.ViewFunc) botkit.ViewFunc {
	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {

		// check if the person who sent the command is in the list of admins
		for _, admin := range adminsID {
			if admin == update.Message.From.ID {
				return next(ctx, bot, update)
			}
		}

		if _, err := bot.Send(tgbotapi.NewMessage(
			update.Message.Chat.ID,
			"У вас нет прав на выполнение данных действий",
		)); err != nil {
			return err
		}
		return nil
	}
}
