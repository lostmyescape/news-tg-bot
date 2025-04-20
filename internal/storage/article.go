package storage

import (
	"context"
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/lostmyescape/news-tg-bot/internal/model"
	"github.com/lostmyescape/news-tg-bot/logger"
	"github.com/samber/lo"
	"time"
)

type ArticlePostgresStorage struct {
	db *sqlx.DB
}

func NewArticleStorage(db *sqlx.DB) *ArticlePostgresStorage {
	return &ArticlePostgresStorage{db: db}
}

// Store сохранение статьи
func (s *ArticlePostgresStorage) Store(ctx context.Context, article model.Article) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	if _, err := conn.ExecContext(ctx,
		`INSERT INTO articles (source_id, title, link, summary, published_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT DO NOTHING`, article.SourceID, article.Title, article.Link, article.Summary, article.PublishedAt,
	); err != nil {
		return err
	}
	return nil
}

// AllNotPosted покажет статьи, которые еще не были опубликованы
func (s *ArticlePostgresStorage) AllNotPosted(ctx context.Context, since time.Time) ([]model.Article, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var articles []dbArticle

	if err := conn.SelectContext(
		ctx,
		&articles,
		`SELECT * FROM articles
         WHERE posted_at IS NULL
         ORDER BY published_at DESC
         `); err != nil {
		return nil, err
	}

	logger.Log.Infof("notifier: AllNotPosted since %v", since)

	return lo.Map(articles, func(article dbArticle, _ int) model.Article {
		return model.Article{
			ID:          article.ID,
			SourceID:    article.SourceID,
			Title:       article.Title,
			Link:        article.Link,
			Summary:     article.Summary,
			PostedAt:    article.PostedAt.Time,
			PublishedAt: article.PublishedAt,
			CreatedAt:   article.CreatedAt,
		}
	}), nil

}

// MarkAsPosted отметка статьи о том, что она была запощена
func (s *ArticlePostgresStorage) MarkAsPosted(ctx context.Context, article model.Article) error {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()

	if _, err := conn.ExecContext(
		ctx,
		`UPDATE articles SET posted_at = $1::timestamp WHERE id = $2;`,
		time.Now().UTC().Format(time.RFC3339),
		article.ID,
	); err != nil {
		return err
	}

	return nil
}

type dbArticle struct {
	ID          int64        `db:"id"`
	SourceID    int64        `db:"source_id"`
	Title       string       `db:"title"`
	Link        string       `db:"link"`
	Summary     string       `db:"summary"`
	PublishedAt time.Time    `db:"published_at"`
	PostedAt    sql.NullTime `db:"posted_at"`
	CreatedAt   time.Time    `db:"created_at"`
}
