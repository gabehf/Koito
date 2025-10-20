-- name: InsertRelease :one
INSERT INTO releases (musicbrainz_id, various_artists, image, image_source)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRelease :one
SELECT 
  *,
  get_artists_for_release(id) AS artists
FROM releases_with_title
WHERE id = $1 LIMIT 1;

-- name: GetReleaseByMbzID :one
SELECT * FROM releases_with_title
WHERE musicbrainz_id = $1 LIMIT 1;

-- name: GetReleaseByImageID :one
SELECT * FROM releases
WHERE image = $1 LIMIT 1;

-- name: GetReleaseByArtistAndTitle :one
SELECT r.*
FROM releases_with_title r
JOIN artist_releases ar ON r.id = ar.release_id
WHERE r.title = $1 AND ar.artist_id = $2
LIMIT 1;

-- name: GetReleaseByArtistAndTitles :one
SELECT r.*
FROM releases_with_title r
JOIN artist_releases ar ON r.id = ar.release_id
WHERE r.title = ANY ($1::TEXT[]) AND ar.artist_id = $2
LIMIT 1;

-- name: GetTopReleasesFromArtist :many
SELECT
  r.*,
  COUNT(*) AS listen_count,
  get_artists_for_release(r.id) AS artists
FROM listens l
JOIN tracks t ON l.track_id = t.id
JOIN releases_with_title r ON t.release_id = r.id
JOIN artist_releases ar ON r.id = ar.release_id
WHERE ar.artist_id = $5
  AND l.listened_at BETWEEN $1 AND $2
GROUP BY r.id, r.title, r.musicbrainz_id, r.various_artists, r.image, r.image_source
ORDER BY listen_count DESC, r.id
LIMIT $3 OFFSET $4;

-- name: GetTopReleasesPaginated :many
SELECT
  r.*,
  COUNT(*) AS listen_count,
  get_artists_for_release(r.id) AS artists
FROM listens l
JOIN tracks t ON l.track_id = t.id
JOIN releases_with_title r ON t.release_id = r.id
WHERE l.listened_at BETWEEN $1 AND $2
GROUP BY r.id, r.title, r.musicbrainz_id, r.various_artists, r.image, r.image_source
ORDER BY listen_count DESC, r.id
LIMIT $3 OFFSET $4;

-- name: CountTopReleases :one
SELECT COUNT(DISTINCT r.id) AS total_count
FROM listens l
JOIN tracks t ON l.track_id = t.id
JOIN releases r ON t.release_id = r.id
WHERE l.listened_at BETWEEN $1 AND $2;

-- name: CountReleasesFromArtist :one
SELECT COUNT(*)
FROM releases r 
JOIN artist_releases ar ON r.id = ar.release_id
WHERE ar.artist_id = $1;

-- name: AssociateArtistToRelease :exec
INSERT INTO artist_releases (artist_id, release_id, is_primary)
VALUES ($1, $2, $3)
ON CONFLICT DO NOTHING;

-- name: GetReleasesWithoutImages :many
SELECT
  r.*,
  get_artists_for_release(r.id) AS artists
FROM releases_with_title r 
WHERE r.image IS NULL 
  AND r.id > $2
ORDER BY r.id ASC
LIMIT $1;

-- name: UpdateReleaseMbzID :exec
UPDATE releases SET musicbrainz_id = $2
WHERE id = $1;

-- name: UpdateReleaseVariousArtists :exec
UPDATE releases SET various_artists = $2
WHERE id = $1;

-- name: UpdateReleasePrimaryArtist :exec
UPDATE artist_releases SET is_primary = $3
WHERE artist_id = $1 AND release_id = $2;

-- name: UpdateReleaseImage :exec
UPDATE releases SET image = $2, image_source = $3
WHERE id = $1;

-- name: DeleteRelease :exec
DELETE FROM releases WHERE id = $1;

-- name: DeleteReleasesFromArtist :exec 
DELETE FROM releases r
USING artist_releases ar
WHERE ar.release_id = r.id
  AND ar.artist_id = $1;