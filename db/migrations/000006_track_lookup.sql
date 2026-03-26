-- +goose Up
CREATE TABLE track_lookup (
    lookup_key TEXT NOT NULL PRIMARY KEY,
    artist_id INT NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    album_id INT NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    track_id INT NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_track_lookup_track_id ON track_lookup(track_id);
CREATE INDEX idx_track_lookup_artist_id ON track_lookup(artist_id);
CREATE INDEX idx_track_lookup_album_id ON track_lookup(album_id);

-- +goose Down
DROP TABLE IF EXISTS track_lookup;
