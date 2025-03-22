-- +goose Up
CREATE TABLE IF NOT EXISTS file (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	post_id UUID NOT NULL,
	key VARCHAR(255) UNIQUE NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
	deleted_at TIMESTAMP DEFAULT NULL
);
-- +goose Down
DROP TABLE IF EXISTS file;