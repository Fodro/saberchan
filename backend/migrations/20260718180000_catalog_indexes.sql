-- +goose Up
CREATE INDEX IF NOT EXISTS idx_thread_board_updated_at ON thread (board_id, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_post_thread_number ON post (thread_id, number);
CREATE INDEX IF NOT EXISTS idx_attachment_post_id ON attachment (post_id);

-- +goose Down
DROP INDEX IF EXISTS idx_attachment_post_id;
DROP INDEX IF EXISTS idx_post_thread_number;
DROP INDEX IF EXISTS idx_thread_board_updated_at;
