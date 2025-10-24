-- name: InsertTrack :one
INSERT INTO tracks (musicbrainz_id, release_id, duration)
VALUES ($1, $2, $3)
RETURNING *;

-- name: AssociateArtistToTrack :exec
INSERT INTO artist_tracks (artist_id, track_id, is_primary)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetTrack :one
SELECT 
  t.*,
  get_artists_for_track(t.id) AS artists,
  r.image
FROM tracks_with_title t
JOIN releases r ON t.release_id = r.id
WHERE t.id = $1 LIMIT 1;

-- name: GetTrackByMbzID :one
SELECT * FROM tracks_with_title
WHERE musicbrainz_id = $1 LIMIT 1;

-- name: GetAllTracksFromArtist :many
SELECT t.*
FROM tracks_with_title t
JOIN artist_tracks at ON t.id = at.track_id
WHERE at.artist_id = $1;

-- name: GetTrackByTitleAndArtists :one
SELECT t.*
FROM tracks_with_title t
JOIN artist_tracks at ON at.track_id = t.id
WHERE t.title = $1
  AND at.artist_id = ANY($2::int[])
GROUP BY t.id, t.title, t.musicbrainz_id, t.duration, t.release_id
HAVING COUNT(DISTINCT at.artist_id) = cardinality($2::int[]);

-- name: GetTopTracksPaginated :many
SELECT
    t.id,
    t.title,
    t.musicbrainz_id,
    t.release_id,
    r.image,
    COUNT(*) AS listen_count,
    get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
JOIN releases r ON t.release_id = r.id
WHERE l.listened_at BETWEEN $1 AND $2
GROUP BY t.id, t.title, t.musicbrainz_id, t.release_id, r.image
ORDER BY listen_count DESC, t.id
LIMIT $3 OFFSET $4;

-- name: GetTopTracksByArtistPaginated :many
SELECT
    t.id,
    t.title,
    t.musicbrainz_id,
    t.release_id,
    r.image,
    COUNT(*) AS listen_count,
    get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
JOIN releases r ON t.release_id = r.id
JOIN artist_tracks at ON at.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND at.artist_id = $5
GROUP BY t.id, t.title, t.musicbrainz_id, t.release_id, r.image
ORDER BY listen_count DESC, t.id
LIMIT $3 OFFSET $4;

-- name: GetTopTracksInReleasePaginated :many
SELECT
    t.id,
    t.title,
    t.musicbrainz_id,
    t.release_id,
    r.image,
    COUNT(*) AS listen_count,
    get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
JOIN releases r ON t.release_id = r.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND t.release_id = $5
GROUP BY t.id, t.title, t.musicbrainz_id, t.release_id, r.image
ORDER BY listen_count DESC, t.id
LIMIT $3 OFFSET $4;

-- name: CountTopTracks :one
SELECT COUNT(DISTINCT l.track_id) AS total_count
FROM listens l
WHERE l.listened_at BETWEEN $1 AND $2;

-- name: CountTopTracksByArtist :one
SELECT COUNT(DISTINCT l.track_id) AS total_count
FROM listens l
JOIN artist_tracks at ON l.track_id = at.track_id
WHERE l.listened_at BETWEEN $1 AND $2
AND at.artist_id = $3;

-- name: CountTopTracksByRelease :one
SELECT COUNT(DISTINCT l.track_id) AS total_count
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
AND t.release_id = $3;

-- name: UpdateTrackMbzID :exec
UPDATE tracks SET musicbrainz_id = $2
WHERE id = $1;

-- name: UpdateTrackDuration :exec
UPDATE tracks SET duration = $2
WHERE id = $1;

-- name: UpdateReleaseForAll :exec
UPDATE tracks SET release_id = $2
WHERE release_id = $1;

-- name: UpdateTrackPrimaryArtist :exec
UPDATE artist_tracks SET is_primary = $3
WHERE artist_id = $1 AND track_id = $2;

-- name: DeleteTrack :exec
DELETE FROM tracks WHERE id = $1;