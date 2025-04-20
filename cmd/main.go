package main

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lostmyescape/news-tg-bot/internal/bot"
	"github.com/lostmyescape/news-tg-bot/internal/botkit"
	"github.com/lostmyescape/news-tg-bot/internal/config"
	fetcher2 "github.com/lostmyescape/news-tg-bot/internal/fetcher"
	"github.com/lostmyescape/news-tg-bot/internal/notifier"
	"github.com/lostmyescape/news-tg-bot/internal/storage"
	"github.com/lostmyescape/news-tg-bot/internal/summary"
	"github.com/lostmyescape/news-tg-bot/logger"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	logger.Init()
	logger.Log.Infow("Logger initialized", "env", "dev")

	token := config.Get().TelegramBotToken
	if token == "" {
		logger.Log.Warn("telegram bot token is empty, bot won't be started")
		return
	}

	botAPI, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		logger.Log.Errorw("failed to create bot", "err", err)
		return
	}

	db, err := sqlx.Connect("postgres", config.Get().DatabaseDSN)
	if err != nil {
		logger.Log.Errorw("failed to connect to database", "err", err)
		return
	}
	defer db.Close()

	var (
		articleStorage = storage.NewArticleStorage(db)
		sourceStorage  = storage.NewSourceStorage(db)
		fetcher        = fetcher2.New(
			articleStorage,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
		n = notifier.New(
			articleStorage,
			summary.NewOpenAiSummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
			botAPI,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	newsBot := botkit.New(botAPI)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())
	// запуск fetcher
	go func(ctx context.Context) {
		if err := fetcher.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Log.Errorw("failed to start fetcher", "err", err)
				return
			}

			logger.Log.Info("fetcher stopped")
		}
	}(ctx)

	// запуск notifier
	go func(ctx context.Context) {
		if err := n.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Log.Errorw("failed to start notifier", "err", err)
				return
			}

			logger.Log.Info("n stopped")
		}
	}(ctx)

	if err := newsBot.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Log.Errorw("failed to run bot:", "err", err)
			return
		}
		logger.Log.Info("bot stopped")
	}

}
