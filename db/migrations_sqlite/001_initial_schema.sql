-- +goose Up

CREATE TABLE IF NOT EXISTS artists (
    id             INTEGER PRIMARY KEY,
    musicbrainz_id TEXT,
    image          TEXT,
    image_source   TEXT
);

CREATE TABLE IF NOT EXISTS artist_aliases (
    artist_id  INTEGER NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    alias      TEXT NOT NULL,
    source     TEXT NOT NULL,
    is_primary INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (artist_id, alias)
);

CREATE TABLE IF NOT EXISTS releases (
    id              INTEGER PRIMARY KEY,
    musicbrainz_id  TEXT,
    image           TEXT,
    image_source    TEXT,
    various_artists INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS artist_releases (
    artist_id  INTEGER NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    is_primary INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (artist_id, release_id)
);

CREATE TABLE IF NOT EXISTS release_aliases (
    release_id INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    alias      TEXT NOT NULL,
    source     TEXT NOT NULL,
    is_primary INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (release_id, alias)
);

CREATE TABLE IF NOT EXISTS tracks (
    id             INTEGER PRIMARY KEY,
    musicbrainz_id TEXT,
    release_id     INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    duration       INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS artist_tracks (
    artist_id  INTEGER NOT NULL REFERENCES artists(id) ON DELETE CASCADE,
    track_id   INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    is_primary INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (artist_id, track_id)
);

CREATE TABLE IF NOT EXISTS track_aliases (
    track_id   INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    alias      TEXT NOT NULL,
    source     TEXT NOT NULL,
    is_primary INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (track_id, alias)
);

CREATE TABLE IF NOT EXISTS users (
    id       INTEGER PRIMARY KEY,
    username TEXT NOT NULL UNIQUE,
    role     TEXT NOT NULL DEFAULT 'user' CHECK (role IN ('admin', 'user')),
    password BLOB NOT NULL
);

CREATE TABLE IF NOT EXISTS api_keys (
    id         INTEGER PRIMARY KEY,
    key        TEXT NOT NULL UNIQUE,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    label      TEXT NOT NULL DEFAULT ''
);

CREATE TABLE IF NOT EXISTS sessions (
    id         TEXT PRIMARY KEY,
    user_id    INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at INTEGER NOT NULL,
    expires_at INTEGER NOT NULL,
    persistent INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS listens (
    track_id    INTEGER NOT NULL REFERENCES tracks(id) ON DELETE CASCADE,
    listened_at INTEGER NOT NULL,
    user_id     INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    client      TEXT NOT NULL DEFAULT '',
    PRIMARY KEY (track_id, listened_at)
);

CREATE VIEW IF NOT EXISTS artists_with_name AS
SELECT a.*, aa.alias AS name
FROM artists a
JOIN artist_aliases aa ON a.id = aa.artist_id
WHERE aa.is_primary = 1;

CREATE VIEW IF NOT EXISTS releases_with_title AS
SELECT r.*, ra.alias AS title
FROM releases r
JOIN release_aliases ra ON r.id = ra.release_id
WHERE ra.is_primary = 1;

CREATE VIEW IF NOT EXISTS tracks_with_title AS
SELECT t.*, ta.alias AS title
FROM tracks t
JOIN track_aliases ta ON t.id = ta.track_id
WHERE ta.is_primary = 1;

-- +goose StatementBegin
CREATE TRIGGER IF NOT EXISTS trg_delete_orphan_releases
AFTER DELETE ON artist_releases
BEGIN
    DELETE FROM releases
    WHERE id = OLD.release_id
      AND NOT EXISTS (
          SELECT 1 FROM artist_releases WHERE release_id = OLD.release_id
      );
END;
-- +goose StatementEnd

CREATE INDEX IF NOT EXISTS idx_listens_listened_at       ON listens(listened_at);
CREATE INDEX IF NOT EXISTS idx_listens_track_id          ON listens(track_id);
CREATE INDEX IF NOT EXISTS idx_listens_user_id           ON listens(user_id);
CREATE INDEX IF NOT EXISTS idx_artist_tracks_track_id    ON artist_tracks(track_id);
CREATE INDEX IF NOT EXISTS idx_artist_tracks_artist_id   ON artist_tracks(artist_id);
CREATE INDEX IF NOT EXISTS idx_artist_releases_release_id ON artist_releases(release_id);
CREATE INDEX IF NOT EXISTS idx_tracks_release_id         ON tracks(release_id);
CREATE INDEX IF NOT EXISTS idx_artist_aliases_alias      ON artist_aliases(alias);
CREATE INDEX IF NOT EXISTS idx_release_aliases_alias     ON release_aliases(alias);
CREATE INDEX IF NOT EXISTS idx_track_aliases_alias       ON track_aliases(alias);
CREATE INDEX IF NOT EXISTS idx_sessions_expires_at       ON sessions(expires_at);

-- +goose Down

DROP TRIGGER IF EXISTS trg_delete_orphan_releases;
DROP VIEW IF EXISTS tracks_with_title;
DROP VIEW IF EXISTS releases_with_title;
DROP VIEW IF EXISTS artists_with_name;
DROP TABLE IF EXISTS listens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS api_keys;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS track_aliases;
DROP TABLE IF EXISTS artist_tracks;
DROP TABLE IF EXISTS tracks;
DROP TABLE IF EXISTS release_aliases;
DROP TABLE IF EXISTS artist_releases;
DROP TABLE IF EXISTS releases;
DROP TABLE IF EXISTS artist_aliases;
DROP TABLE IF EXISTS artists;
