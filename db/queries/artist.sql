-- name: InsertArtist :one
INSERT INTO artists (musicbrainz_id, image, image_source)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetArtist :one
SELECT 
  a.*,
  array_agg(aa.alias)::text[] AS aliases
FROM artists_with_name a
LEFT JOIN artist_aliases aa ON a.id = aa.artist_id
WHERE a.id = $1
GROUP BY a.id, a.musicbrainz_id, a.image, a.image_source, a.name;

-- name: GetTrackArtists :many
SELECT 
  a.*,
  at.is_primary as is_primary
FROM artists_with_name a
LEFT JOIN artist_tracks at ON a.id = at.artist_id
WHERE at.track_id = $1
GROUP BY a.id, a.musicbrainz_id, a.image, a.image_source, a.name, at.is_primary;

-- name: GetArtistByImage :one
SELECT * FROM artists WHERE image = $1 LIMIT 1;

-- name: GetReleaseArtists :many
SELECT 
  a.*,
  ar.is_primary as is_primary
FROM artists_with_name a
LEFT JOIN artist_releases ar ON a.id = ar.artist_id
WHERE ar.release_id = $1
GROUP BY a.id, a.musicbrainz_id, a.image, a.image_source, a.name, ar.is_primary;

-- name: GetArtistByName :one
WITH artist_with_aliases AS (
  SELECT 
    a.*,
    COALESCE(array_agg(aa.alias), '{}')::text[] AS aliases
  FROM artists_with_name a
  LEFT JOIN artist_aliases aa ON a.id = aa.artist_id
  WHERE a.id IN (
    SELECT aa2.artist_id FROM artist_aliases aa2 WHERE aa2.alias = $1
  )
  GROUP BY a.id, a.musicbrainz_id, a.image, a.image_source, a.name
)
SELECT * FROM artist_with_aliases;

-- name: GetArtistByMbzID :one
SELECT 
  a.*,
  array_agg(aa.alias)::text[] AS aliases
FROM artists_with_name a
LEFT JOIN artist_aliases aa ON a.id = aa.artist_id
WHERE a.musicbrainz_id = $1
GROUP BY a.id, a.musicbrainz_id, a.image, a.image_source, a.name;

-- name: GetTopArtistsPaginated :many
SELECT
    a.id,
    a.name,
    a.musicbrainz_id,
    a.image,
    COUNT(*) AS listen_count
FROM listens l
JOIN tracks t ON l.track_id = t.id
JOIN artist_tracks at ON at.track_id = t.id
JOIN artists_with_name a ON a.id = at.artist_id
WHERE l.listened_at BETWEEN $1 AND $2
GROUP BY a.id, a.name, a.musicbrainz_id, a.image, a.image_source, a.name
ORDER BY listen_count DESC, a.id
LIMIT $3 OFFSET $4;

-- name: CountTopArtists :one
SELECT COUNT(DISTINCT at.artist_id) AS total_count
FROM listens l
JOIN artist_tracks at ON l.track_id = at.track_id
WHERE l.listened_at BETWEEN $1 AND $2;

-- name: UpdateArtistMbzID :exec
UPDATE artists SET musicbrainz_id = $2
WHERE id = $1;

-- name: UpdateArtistImage :exec
UPDATE artists SET image = $2, image_source = $3
WHERE id = $1;

-- name: DeleteConflictingArtistTracks :exec
DELETE FROM artist_tracks at
WHERE at.artist_id = $1
  AND track_id IN (
    SELECT at.track_id FROM artist_tracks at WHERE at.artist_id = $2
  );

-- name: UpdateArtistTracks :exec
UPDATE artist_tracks
SET artist_id = $2
WHERE artist_id = $1;

-- name: DeleteConflictingArtistReleases :exec
DELETE FROM artist_releases ar
WHERE ar.artist_id = $1
  AND release_id IN (
    SELECT ar.release_id FROM artist_releases ar WHERE ar.artist_id = $2
  );

-- name: UpdateArtistReleases :exec
UPDATE artist_releases
SET artist_id = $2
WHERE artist_id = $1;

-- name: DeleteArtist :exec
DELETE FROM artists WHERE id = $1;