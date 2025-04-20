package source

import (
	"context"
	"github.com/SlyMarbo/rss"
	"github.com/lostmyescape/news-tg-bot/internal/model"
	"github.com/samber/lo"
)

type RSSSource struct {
	URL        string
	SourceID   int64
	SourceName string
}

// NewRSSSourceFromModel принимает модельку из слоя данных, и на ее основе создает RSSSource
func NewRSSSourceFromModel(m model.Source) RSSSource {
	return RSSSource{
		URL:        m.FeedURL,
		SourceID:   m.ID,
		SourceName: m.Name,
	}
}

// Fetch загружает RSS-ленту по s.URL, преобразовывает каждый rss. Item в model. Item, возвращает слайс этих айтемов
func (s RSSSource) Fetch(ctx context.Context) ([]model.Item, error) {
	feed, err := s.loadFeed(ctx, s.URL)
	if err != nil {
		return nil, err
	}

	return lo.Map(feed.Items, func(item *rss.Item, _ int) model.Item {
		return model.Item{
			Title:      item.Title,
			Categories: item.Categories,
			Link:       item.Link,
			Date:       item.Date,
			Summary:    item.Summary,
			SourceName: s.SourceName,
		}
	}), nil
}

// loadFeed делает асинхронную загрузку RSS-ленты
func (s RSSSource) loadFeed(ctx context.Context, url string) (*rss.Feed, error) {
	var (
		feedCh = make(chan *rss.Feed)
		errCh  = make(chan error)
	)

	go func() {
		feed, err := rss.Fetch(url)
		if err != nil {
			errCh <- err
			return
		}

		feedCh <- feed
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case err := <-errCh:
		return nil, err
	case feed := <-feedCh:
		return feed, nil
	}
}

func (s RSSSource) ID() int64 {
	return s.SourceID
}

func (s RSSSource) Name() string {
	return s.SourceName
}
