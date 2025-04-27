package notifier

import (
	"context"
	"fmt"
	readability "github.com/go-shiori/go-readability"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/lostmyescape/news-tg-bot/internal/botkit/markup"
	"github.com/lostmyescape/news-tg-bot/internal/model"
	"github.com/lostmyescape/news-tg-bot/logger"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type ArticleProvider interface {
	AllNotPosted(ctx context.Context) ([]model.Article, error)
	MarkAsPosted(ctx context.Context, article model.Article) error
}

type Summarizer interface {
	Summarize(ctx context.Context, text string) (string, error)
}

type Notifier struct {
	articles         ArticleProvider
	summarizer       Summarizer
	bot              *tgbotapi.BotAPI
	sendInterval     time.Duration
	lookupTimeWindow time.Duration
	channelID        int64
}

func New(
	articleProvider ArticleProvider,
	summarizer Summarizer,
	bot *tgbotapi.BotAPI,
	sendInterval time.Duration,
	lookupTimeWindow time.Duration,
	channelID int64,
) *Notifier {
	return &Notifier{
		articles:         articleProvider,
		summarizer:       summarizer,
		bot:              bot,
		sendInterval:     sendInterval,
		lookupTimeWindow: lookupTimeWindow,
		channelID:        channelID,
	}
}

func (n *Notifier) Start(ctx context.Context) error {
	logger.Log.Info("notifier started")
	ticker := time.NewTicker(n.sendInterval)
	defer ticker.Stop()

	if err := n.SelectAndSendArticle(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ticker.C:
			if err := n.SelectAndSendArticle(ctx); err != nil {
				return err
			}
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// SelectAndSendArticle selects an article that has not yet been published
func (n *Notifier) SelectAndSendArticle(ctx context.Context) error {
	topOneArticles, err := n.articles.AllNotPosted(ctx)
	if err != nil {
		return err
	}

	logger.Log.Infof("notifier: found %d articles to send", len(topOneArticles))

	if len(topOneArticles) == 0 {
		return nil
	}

	article := topOneArticles[0]

	summary, err := n.extractSummary(ctx, article)
	if err != nil {
		return err
	}

	if err := n.sendArticle(article, summary); err != nil {
		return err
	}

	return n.articles.MarkAsPosted(ctx, article)
}

// extractSummary считывает html из новости
func (n *Notifier) extractSummary(ctx context.Context, article model.Article) (string, error) {
	var r io.Reader

	if article.Summary != "" {
		r = strings.NewReader(article.Summary)
	} else {
		resp, err := http.Get(article.Link)
		if err != nil {
			return "", err
		}
		defer resp.Body.Close()

		r = resp.Body
	}

	doc, err := readability.FromReader(r, nil)
	if err != nil {
		return "", err
	}

	summary, err := n.summarizer.Summarize(ctx, cleanText(doc.TextContent))

	if err != nil {
		return "", err
	}

	return "\n\n" + summary, nil
}

// sendArticle отправляет статью
func (n *Notifier) sendArticle(article model.Article, summary string) error {
	const msgFormat = "*%s*%s\n\n%s"

	msg := tgbotapi.NewMessage(n.channelID, fmt.Sprintf(msgFormat, markup.EscapeForMarkdown(article.Title), markup.EscapeForMarkdown(summary), markup.EscapeForMarkdown(article.Link)))
	msg.ParseMode = tgbotapi.ModeMarkdownV2

	_, err := n.bot.Send(msg)

	logger.Log.Infof("notifier: sending message to channel %d", n.channelID)
	if err != nil {
		logger.Log.Errorw("telegram send error", "err", err)
	}

	return nil
}

var redundantNewLines = regexp.MustCompile(`\n{3,}`)

func cleanText(text string) string {
	return redundantNewLines.ReplaceAllString(text, "\n")
}
