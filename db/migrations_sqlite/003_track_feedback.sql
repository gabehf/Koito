-- +goose Up

-- Per-user love/hate feedback on tracks, as submitted through the ListenBrainz
-- feedback/recording-feedback endpoint. score is 1 (love) or -1 (hate); a score
-- of 0 removes the row.
CREATE TABLE IF NOT EXISTS track_feedback (
    track_id   INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    score      INTEGER NOT NULL CHECK (score IN (-1, 1)),
    created_at INTEGER NOT NULL,
    PRIMARY KEY (track_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_track_feedback_user_created ON track_feedback(user_id, created_at);

-- +goose Down

DROP INDEX IF EXISTS idx_track_feedback_user_created;
DROP TABLE IF EXISTS track_feedback;
