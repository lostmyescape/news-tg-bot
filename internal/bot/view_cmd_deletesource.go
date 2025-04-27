package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lostmyescape/news-tg-bot/internal/botkit"
	"github.com/lostmyescape/news-tg-bot/internal/model"
)

type SourceDeleter interface {
	Delete(ctx context.Context, id int64) (int64, error)
}

// ViewCmdDeleteSource delete source from list
func ViewCmdDeleteSource(storage SourceDeleter) botkit.ViewFunc {
	type deleteSourceStorage struct {
		ID int64 `json:"id"`
	}

	return func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error {
		args, err := botkit.ParseJSON[deleteSourceStorage](update.Message.CommandArguments())
		if err != nil {
			return err
		}

		source := model.Source{
			ID: args.ID,
		}

		sourceID, err := storage.Delete(ctx, source.ID)
		if err != nil {
			return err
		}

		var (
			msgText = fmt.Sprintf("Источник `%d` был удален\\.", sourceID)

			reply = tgbotapi.NewMessage(update.Message.Chat.ID, msgText)
		)

		reply.ParseMode = "Markdownv2"

		if _, err := bot.Send(reply); err != nil {
			return err
		}

		return nil
	}
}
