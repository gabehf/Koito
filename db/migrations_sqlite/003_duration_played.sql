-- +goose Up
ALTER TABLE listens ADD COLUMN duration_played INTEGER;
-- +goose Down
ALTER TABLE listens DROP COLUMN duration_played;