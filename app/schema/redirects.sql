-- +goose Up
CREATE TABLE IF NOT EXISTS redirects
(
    id SERIAL PRIMARY KEY,
    path VARCHAR(10) NOT NULL UNIQUE,
    url VARCHAR(100) NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT now(),
    inactive BOOLEAN NOT NULL DEFAULT FALSE
);

-- +goose Down
DROP TABLE redirects;