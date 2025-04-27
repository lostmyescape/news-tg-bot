package storage

import (
	"context"
	"github.com/jmoiron/sqlx"
	"github.com/lostmyescape/news-tg-bot/internal/model"
	"github.com/samber/lo"
	"time"
)

type SourcePostgresStorage struct {
	db *sqlx.DB
}

func NewSourceStorage(db *sqlx.DB) *SourcePostgresStorage {
	return &SourcePostgresStorage{db: db}
}

// Edit edits source by id and returns an id
func (s *SourcePostgresStorage) Edit(ctx context.Context, source model.Source) (int64, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, err
	}
	defer conn.Close()

	var id int64

	row := conn.QueryRowContext(
		ctx,
		`UPDATE sources SET (name, feed_url) = ($1, $2) WHERE id = $3 RETURNING id`,
		source.Name,
		source.FeedURL,
		source.ID,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// Sources returns all from sources
func (s *SourcePostgresStorage) Sources(ctx context.Context) ([]model.Source, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var sources []dbSource
	if err := conn.SelectContext(ctx, &sources, `SELECT * FROM sources`); err != nil {
		return nil, err
	}

	return lo.Map(sources, func(source dbSource, _ int) model.Source { return model.Source(source) }), nil
}

// SourceById selects source by id
func (s *SourcePostgresStorage) SourceById(ctx context.Context, id int64) (*model.Source, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var source dbSource
	if err := conn.GetContext(ctx, &source, `SELECT * FROM sources WHERE id = $1`, id); err != nil {
		return nil, err
	}

	return (*model.Source)(&source), nil

}

// Add adds source to database and returns an id
func (s *SourcePostgresStorage) Add(ctx context.Context, source model.Source) (int64, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	var id int64

	row := conn.QueryRowContext(
		ctx,
		`INSERT INTO sources (name, feed_url, created_at) VALUES ($1, $2, $3) RETURNING id`,
		source.Name,
		source.FeedURL,
		source.CreatedAt,
	)

	if err := row.Err(); err != nil {
		return 0, err
	}

	if err := row.Scan(&id); err != nil {
		return 0, err
	}

	return id, nil
}

// Delete deletes source by id
func (s *SourcePostgresStorage) Delete(ctx context.Context, id int64) (int64, error) {
	conn, err := s.db.Connx(ctx)
	if err != nil {
		return 0, err
	}

	defer conn.Close()

	if _, err := conn.ExecContext(ctx, `DELETE FROM sources WHERE id = $1`, id); err != nil {
		return 0, err
	}

	return id, nil
}

type dbSource struct {
	ID        int64     `db:"id"`
	Name      string    `db:"name"`
	FeedURL   string    `db:"feed_url"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}
