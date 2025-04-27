package botkit

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lostmyescape/news-tg-bot/logger"
	"runtime/debug"
	"time"
)

type Bot struct {
	api      *tgbotapi.BotAPI
	cmdViews map[string]ViewFunc
}

type ViewFunc func(ctx context.Context, bot *tgbotapi.BotAPI, update tgbotapi.Update) error

func New(api *tgbotapi.BotAPI) *Bot {
	return &Bot{
		api: api,
	}
}

func (b *Bot) RegisterCmdView(cmd string, view ViewFunc) {
	if b.cmdViews == nil {
		b.cmdViews = make(map[string]ViewFunc)
	}

	b.cmdViews[cmd] = view
}

// Run runs bot, check an updates from channel
func (b *Bot) Run(ctx context.Context) error {

	// webhook deleted to perform updates via api
	_, err := b.api.Request(tgbotapi.DeleteWebhookConfig{})
	if err != nil {
		logger.Log.Errorw("failed to remove webhook", "err", err)
		return err
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for {
		select {
		case update := <-updates:
			updateCtx, updateCancel := context.WithTimeout(ctx, 5*time.Second)
			b.handleUpdate(updateCtx, update)
			updateCancel()
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// handleUpdate processes a message from the user
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	defer func() {
		if p := recover(); p != nil {
			logger.Log.Panic("panic recovered:", p, string(debug.Stack()))
		}
	}()

	if update.Message == nil || !update.Message.IsCommand() {
		return
	}

	var view ViewFunc

	if !update.Message.IsCommand() {
		return
	}

	cmd := update.Message.Command()

	cmdView, ok := b.cmdViews[cmd]
	if !ok {
		return
	}

	view = cmdView

	if err := view(ctx, b.api, update); err != nil {
		logger.Log.Errorw("failed to handle update:", "err", err)

		if _, err := b.api.Send(
			tgbotapi.NewMessage(update.Message.Chat.ID, "internal error"),
		); err != nil {
			logger.Log.Errorw("failed to handle update:", "err", err)
		}
	}
}
