package main

import (
	"context"
	"errors"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/lostmyescape/news-tg-bot/internal/bot"
	"github.com/lostmyescape/news-tg-bot/internal/bot/middleware"
	"github.com/lostmyescape/news-tg-bot/internal/botkit"
	"github.com/lostmyescape/news-tg-bot/internal/config"
	"github.com/lostmyescape/news-tg-bot/internal/fetcher"
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

	// get telegram token
	token := config.Get().TelegramBotToken
	if token == "" {
		logger.Log.Warn("telegram bot token is empty, bot won't be started")
		return
	}

	// create a bot
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
		articleSaver  = storage.NewArticleStorage(db)
		sourceStorage = storage.NewSourceStorage(db)
		f             = fetcher.New(
			articleSaver,
			sourceStorage,
			config.Get().FetchInterval,
			config.Get().FilterKeywords,
		)
		n = notifier.New(
			articleSaver,
			summary.NewOpenAiSummarizer(config.Get().OpenAIKey, config.Get().OpenAIPrompt),
			botAPI,
			config.Get().NotificationInterval,
			2*config.Get().FetchInterval,
			config.Get().TelegramChannelID,
		)
	)

	// application shutdown
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// views registration
	newsBot := botkit.New(botAPI)
	newsBot.RegisterCmdView("start", bot.ViewCmdStart())
	newsBot.RegisterCmdView("addsource", middleware.AdminOnly(config.Get().Admins, bot.ViewCmdAddSource(sourceStorage)))
	newsBot.RegisterCmdView("listsources", middleware.AdminOnly(config.Get().Admins, bot.ViewCmdListSources(sourceStorage)))
	newsBot.RegisterCmdView("editsource", middleware.AdminOnly(config.Get().Admins, bot.ViewCmdEditSource(sourceStorage)))
	newsBot.RegisterCmdView("deletesource", middleware.AdminOnly(config.Get().Admins, bot.ViewCmdDeleteSource(sourceStorage)))

	// start fetcher
	go func(ctx context.Context) {
		if err := f.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Log.Errorw("failed to start fetcher", "err", err)
				return
			}

			logger.Log.Info("fetcher stopped")
		}
	}(ctx)

	// start notifier
	go func(ctx context.Context) {
		if err := n.Start(ctx); err != nil {
			if !errors.Is(err, context.Canceled) {
				logger.Log.Errorw("failed to start notifier", "err", err)
				return
			}

			logger.Log.Info("n stopped")
		}
	}(ctx)

	// start bot
	if err := newsBot.Run(ctx); err != nil {
		if !errors.Is(err, context.Canceled) {
			logger.Log.Errorw("failed to run bot:", "err", err)
			return
		}
		logger.Log.Info("bot stopped")
	}

}
