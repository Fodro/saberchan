-- +goose Up
CREATE TABLE IF NOT EXISTS board (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	alias VARCHAR(255) UNIQUE NOT NULL,
	name VARCHAR(255) UNIQUE NOT NULL,
	description TEXT NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
	locked BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS thread (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	board_id UUID NOT NULL REFERENCES board(id),
	title VARCHAR(255) NOT NULL,
	locked BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
	updated_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

CREATE TABLE IF NOT EXISTS post (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	thread_id UUID NOT NULL REFERENCES thread(id),
	number SERIAL NOT NULL,
	text TEXT NOT NULL,
	sage BOOLEAN NOT NULL DEFAULT FALSE,
	op_marker BOOLEAN NOT NULL DEFAULT FALSE,
	created_at TIMESTAMP NOT NULL DEFAULT current_timestamp,
	browser_fingerprint VARCHAR(255) NOT NULL,
	ip VARCHAR(255) NOT NULL,
	has_attachment BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS attachment (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	post_id UUID NOT NULL REFERENCES post(id),
	link TEXT NOT NULL,
	name VARCHAR(255) NOT NULL,
	type VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS config (
	id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
	nickname VARCHAR(255) NOT NULL,
	bump_limit NUMERIC NOT NULL,
	current BOOLEAN NOT NULL DEFAULT FALSE,
	site_title VARCHAR(255) NOT NULL,
	created_at TIMESTAMP NOT NULL DEFAULT current_timestamp
);

-- +goose Down
DROP TABLE IF EXISTS attachment;
DROP TABLE IF EXISTS post;
DROP TABLE IF EXISTS thread;
DROP TABLE IF EXISTS board;
DROP TABLE IF EXISTS config;