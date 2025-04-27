package fetcher

import (
	"context"
	"github.com/lostmyescape/news-tg-bot/internal/model"
	"github.com/lostmyescape/news-tg-bot/internal/source"
	"github.com/lostmyescape/news-tg-bot/logger"
	"log"
	"strings"
	"sync"
	"time"
)

type ArticleSaver interface {
	Store(ctx context.Context, article model.Article) error
}

type SourceProvider interface {
	Sources(ctx context.Context) ([]model.Source, error)
}

type Source interface {
	ID() int64
	Name() string
	Fetch(ctx context.Context) ([]model.Item, error)
}

type Fetcher struct {
	articles ArticleSaver
	sources  SourceProvider

	fetchInterval  time.Duration
	filterKeywords []string
}

func New(
	articleSaver ArticleSaver,
	sourceProvider SourceProvider,
	fetchInterval time.Duration,
	filterKeywords []string,
) *Fetcher {
	return &Fetcher{
		articles:       articleSaver,
		sources:        sourceProvider,
		fetchInterval:  fetchInterval,
		filterKeywords: filterKeywords,
	}
}

// Start starts the Fetch
func (f *Fetcher) Start(ctx context.Context) error {
	ticker := time.NewTicker(f.fetchInterval)
	defer ticker.Stop()

	if err := f.Fetch(ctx); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			if err := f.Fetch(ctx); err != nil {
				return err
			}
		}
	}
}

// Fetch loads data from all sources, wraps each sources in goroutine,
// parses rss-feed and sends result to processItems
func (f *Fetcher) Fetch(ctx context.Context) error {
	sources, err := f.sources.Sources(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	for _, src := range sources {
		wg.Add(1)

		go func(source Source) {
			defer wg.Done()

			items, err := source.Fetch(ctx)
			if err != nil {
				log.Printf("Error: fetching items from source %s: %v", source.Name(), err)
				return
			}

			if err := f.processItems(ctx, source, items); err != nil {
				log.Printf("Error: processing items from source %s: %v", source.Name(), err)
				return
			}

			logger.Log.Infof("fetcher: processed items for source %s", source.Name())

		}(source.NewRSSSourceFromModel(src))
	}

	wg.Wait()

	return nil
}

// processItems base logic - normalizes the date, filters items, saves article
func (f *Fetcher) processItems(ctx context.Context, source Source, items []model.Item) error {
	for _, item := range items {
		item.Date = item.Date.UTC()

		if f.itemShouldBeSkipped(item) {
			continue
		}

		if err := f.articles.Store(ctx, model.Article{
			SourceID:    source.ID(),
			Title:       item.Title,
			Link:        item.Link,
			Summary:     item.Summary,
			PublishedAt: item.Date,
		}); err != nil {
			return err
		}
	}
	logger.Log.Infof("fetcher: got %d items from %s", len(items), source.Name())

	return nil
}

// itemShouldBeSkipped skips an item if the category or title contains keywords
func (f *Fetcher) itemShouldBeSkipped(item model.Item) bool {
	categoriesSet := make(map[string]struct{})

	for _, category := range item.Categories {
		categoriesSet[category] = struct{}{}
	}

	titleLower := strings.ToLower(item.Title)

	for _, keyword := range f.filterKeywords {
		if _, found := categoriesSet[keyword]; found || strings.Contains(titleLower, keyword) {
			return true
		}

	}

	return false
}
