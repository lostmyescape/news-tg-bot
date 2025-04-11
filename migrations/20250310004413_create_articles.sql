-- +goose Up
CREATE TABLE articles (
                          id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                          source_id INT NOT NULL,
                          title TEXT NOT NULL,
                          link TEXT NOT NULL UNIQUE,
                          summary TEXT NOT NULL,
                          published_at TIMESTAMP NOT NULL,
                          created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                          posted_at TIMESTAMP
);

-- goose Down
DROP TABLE IF EXISTS articles;
