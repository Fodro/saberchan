-- +goose Up
ALTER TABLE board ADD COLUMN author varchar(255) DEFAULT NULL;

-- +goose Down
ALTER TABLE board DROP COLUMN author;

