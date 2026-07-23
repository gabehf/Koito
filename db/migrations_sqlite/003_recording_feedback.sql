-- +goose Up

CREATE TABLE IF NOT EXISTS recording_feedback (
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    track_id   INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    score      INTEGER NOT NULL CHECK (score IN (1, -1)),
    created_at INTEGER NOT NULL,
    updated_at INTEGER NOT NULL,
    PRIMARY KEY (user_id, track_id)
);

CREATE INDEX IF NOT EXISTS idx_recording_feedback_track_id ON recording_feedback(track_id);

-- +goose Down

DROP INDEX IF EXISTS idx_recording_feedback_track_id;
DROP TABLE IF EXISTS recording_feedback;
