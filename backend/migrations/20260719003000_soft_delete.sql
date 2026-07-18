-- +goose Up
ALTER TABLE board ADD COLUMN deleted_at TIMESTAMPTZ NULL;
ALTER TABLE board ADD COLUMN purged_at TIMESTAMPTZ NULL;

ALTER TABLE thread ADD COLUMN deleted_at TIMESTAMPTZ NULL;
ALTER TABLE thread ADD COLUMN purged_at TIMESTAMPTZ NULL;

ALTER TABLE post ADD COLUMN deleted_at TIMESTAMPTZ NULL;
ALTER TABLE post ADD COLUMN purged_at TIMESTAMPTZ NULL;

ALTER TABLE attachment ADD COLUMN key TEXT NOT NULL DEFAULT '';

CREATE INDEX IF NOT EXISTS idx_board_deleted_at ON board (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_thread_deleted_at ON thread (deleted_at) WHERE deleted_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_post_deleted_at ON post (deleted_at) WHERE deleted_at IS NOT NULL;

-- +goose Down
DROP INDEX IF EXISTS idx_post_deleted_at;
DROP INDEX IF EXISTS idx_thread_deleted_at;
DROP INDEX IF EXISTS idx_board_deleted_at;

ALTER TABLE attachment DROP COLUMN key;

ALTER TABLE post DROP COLUMN purged_at;
ALTER TABLE post DROP COLUMN deleted_at;

ALTER TABLE thread DROP COLUMN purged_at;
ALTER TABLE thread DROP COLUMN deleted_at;

ALTER TABLE board DROP COLUMN purged_at;
ALTER TABLE board DROP COLUMN deleted_at;
