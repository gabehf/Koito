-- +goose Up
PRAGMA foreign_keys = OFF;

DROP TRIGGER IF EXISTS trg_delete_orphan_releases;

-- tracks: add UNIQUE to musicbrainz_id
DROP VIEW IF EXISTS tracks_with_title;
CREATE TABLE tracks_new (
    id             INTEGER PRIMARY KEY,
    musicbrainz_id TEXT UNIQUE,
    release_id     INTEGER NOT NULL REFERENCES releases(id) ON DELETE CASCADE,
    duration       INTEGER NOT NULL DEFAULT 0
);
INSERT INTO tracks_new SELECT * FROM tracks;
DROP TABLE tracks;
ALTER TABLE tracks_new RENAME TO tracks;
CREATE INDEX IF NOT EXISTS idx_tracks_release_id ON tracks(release_id);
CREATE VIEW tracks_with_title AS
SELECT t.*, ta.alias AS title
FROM tracks t
JOIN track_aliases ta ON t.id = ta.track_id
WHERE ta.is_primary = 1;

-- releases: add UNIQUE to musicbrainz_id
DROP VIEW IF EXISTS releases_with_title;
CREATE TABLE releases_new (
    id              INTEGER PRIMARY KEY,
    musicbrainz_id  TEXT UNIQUE,
    image           TEXT,
    image_source    TEXT,
    various_artists INTEGER NOT NULL DEFAULT 0
);
INSERT INTO releases_new SELECT * FROM releases;
DROP TABLE releases;
ALTER TABLE releases_new RENAME TO releases;
CREATE VIEW releases_with_title AS
SELECT r.*, ra.alias AS title
FROM releases r
JOIN release_aliases ra ON r.id = ra.release_id
WHERE ra.is_primary = 1;

-- artists: add UNIQUE to musicbrainz_id
DROP VIEW IF EXISTS artists_with_name;
CREATE TABLE artists_new (
    id             INTEGER PRIMARY KEY,
    musicbrainz_id TEXT UNIQUE,
    image          TEXT,
    image_source   TEXT
);
INSERT INTO artists_new SELECT * FROM artists;
DROP TABLE artists;
ALTER TABLE artists_new RENAME TO artists;
CREATE VIEW artists_with_name AS
SELECT a.*, aa.alias AS name
FROM artists a
JOIN artist_aliases aa ON a.id = aa.artist_id
WHERE aa.is_primary = 1;

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

PRAGMA foreign_keys = ON;
