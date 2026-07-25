-- +goose Up

ALTER TABLE artists ADD COLUMN exclude_from_top_lists INTEGER NOT NULL DEFAULT 0;
ALTER TABLE releases ADD COLUMN exclude_from_top_lists INTEGER NOT NULL DEFAULT 0;

-- +goose Down

ALTER TABLE artists DROP COLUMN exclude_from_top_lists;
ALTER TABLE releases DROP COLUMN exclude_from_top_lists;
