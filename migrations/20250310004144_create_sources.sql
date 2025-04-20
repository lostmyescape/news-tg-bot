-- +goose Up
CREATE TABLE sources (
                         id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
                         name TEXT NOT NULL,
                         feed_url TEXT NOT NULL,
                         created_at TIMESTAMP NOT NULL DEFAULT NOW(),
                         updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);


-- +goose Down
DROP TABLE IF EXISTS sources;

INSERT INTO sources (name, feed_url) VALUES ('dev.to', 'https://dev.to/feed/tag/golang'), ('hashnode.com', 'https://hashnode.com/n/golang/rss')
