package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lostmyescape/news-tg-bot/internal/botkit"
	"github.com/lostmyescape/news-tg-bot/internal/model"
)

type EditStorage interface {
	Edit(ctx context.Context, source model.Source) (int64, error)
}

// ViewCmdEditSource change name and url from list
func ViewCmdEditSource(storage EditStorage) botkit.ViewFunc {
	type editSourceArgs struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
		URL  string `json:"url"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[editSourceArgs](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		source := model.Source{
			ID:      args.ID,
			Name:    args.Name,
			FeedURL: args.URL,
		}

		sourceID, err := storage.Edit(ctx, source)
		if err != nil {
			return err
		}

		var (
			msgText = fmt.Sprintf("источник с ID: `%d` был изменен\\.", sourceID)

			reply = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = "Markdownv2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
