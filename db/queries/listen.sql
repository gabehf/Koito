-- name: InsertListen :exec
INSERT INTO listens (track_id, listened_at, user_id, client)
VALUES ($1, $2, $3, $4)
ON CONFLICT DO NOTHING;

-- name: GetLastListensPaginated :many
SELECT 
  l.*,
  t.title AS track_title,
  t.release_id AS release_id,
  get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
ORDER BY l.listened_at DESC
LIMIT $3 OFFSET $4;

-- name: GetLastListensFromArtistPaginated :many
SELECT 
  l.*,
  t.title AS track_title,
  t.release_id AS release_id,
  get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
JOIN artist_tracks at ON t.id = at.track_id 
WHERE at.artist_id = $5
  AND l.listened_at BETWEEN $1 AND $2
ORDER BY l.listened_at DESC
LIMIT $3 OFFSET $4;

-- name: GetFirstListenFromArtist :one
SELECT 
  l.*
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
JOIN artist_tracks at ON t.id = at.track_id 
WHERE at.artist_id = $1
ORDER BY l.listened_at ASC
LIMIT 1;

-- name: GetLastListensFromReleasePaginated :many
SELECT 
  l.*,
  t.title AS track_title,
  t.release_id AS release_id,
  get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND t.release_id = $5
ORDER BY l.listened_at DESC
LIMIT $3 OFFSET $4;

-- name: GetFirstListenFromRelease :one
SELECT 
  l.*
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE t.release_id = $1
ORDER BY l.listened_at ASC
LIMIT 1;

-- name: GetLastListensFromTrackPaginated :many
SELECT 
  l.*,
  t.title AS track_title,
  t.release_id AS release_id,
  get_artists_for_track(t.id) AS artists
FROM listens l
JOIN tracks_with_title t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND t.id = $5
ORDER BY l.listened_at DESC
LIMIT $3 OFFSET $4;

-- name: GetFirstListenFromTrack :one
SELECT 
  l.*
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE t.id = $1
ORDER BY l.listened_at ASC
LIMIT 1;

-- name: CountListens :one
SELECT COUNT(*) AS total_count
FROM listens l
WHERE l.listened_at BETWEEN $1 AND $2;

-- name: CountListensFromTrack :one
SELECT COUNT(*) AS total_count
FROM listens l
WHERE l.listened_at BETWEEN $1 AND $2
  AND l.track_id = $3;

-- name: CountListensFromArtist :one
SELECT COUNT(*) AS total_count
FROM listens l
JOIN artist_tracks at ON l.track_id = at.track_id
WHERE l.listened_at BETWEEN $1 AND $2
  AND at.artist_id = $3;

-- name: CountListensFromRelease :one
SELECT COUNT(*) AS total_count
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND t.release_id = $3;

-- name: CountTimeListened :one
SELECT COALESCE(SUM(t.duration), 0)::BIGINT AS seconds_listened
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2;

-- name: CountTimeListenedToArtist :one
SELECT COALESCE(SUM(t.duration), 0)::BIGINT AS seconds_listened
FROM listens l
JOIN tracks t ON l.track_id = t.id
JOIN artist_tracks at ON t.id = at.track_id
WHERE l.listened_at BETWEEN $1 AND $2
  AND at.artist_id = $3;

-- name: CountTimeListenedToRelease :one
SELECT COALESCE(SUM(t.duration), 0)::BIGINT AS seconds_listened
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND t.release_id = $3;

-- name: CountTimeListenedToTrack :one
SELECT COALESCE(SUM(t.duration), 0)::BIGINT AS seconds_listened
FROM listens l
JOIN tracks t ON l.track_id = t.id
WHERE l.listened_at BETWEEN $1 AND $2
  AND t.id = $3;

-- name: ListenActivity :many
WITH buckets AS (
  SELECT generate_series($1::timestamptz, $2::timestamptz, $3::interval) AS bucket_start
),
bucketed_listens AS (
  SELECT
    b.bucket_start,
    COUNT(l.listened_at) AS listen_count
  FROM buckets b
  LEFT JOIN listens l
    ON l.listened_at >= b.bucket_start
    AND l.listened_at < b.bucket_start + $3::interval
  GROUP BY b.bucket_start
  ORDER BY b.bucket_start
)
SELECT * FROM bucketed_listens;

-- name: ListenActivityForArtist :many
WITH buckets AS (
  SELECT generate_series($1::timestamptz, $2::timestamptz, $3::interval) AS bucket_start
),
filtered_listens AS (
  SELECT l.*
  FROM listens l
  JOIN artist_tracks t ON l.track_id = t.track_id
  WHERE t.artist_id = $4
),
bucketed_listens AS (
  SELECT
    b.bucket_start,
    COUNT(l.listened_at) AS listen_count
  FROM buckets b
  LEFT JOIN filtered_listens l
    ON l.listened_at >= b.bucket_start
    AND l.listened_at < b.bucket_start + $3::interval
  GROUP BY b.bucket_start
  ORDER BY b.bucket_start
)
SELECT * FROM bucketed_listens;

-- name: ListenActivityForRelease :many
WITH buckets AS (
  SELECT generate_series($1::timestamptz, $2::timestamptz, $3::interval) AS bucket_start
),
filtered_listens AS (
  SELECT l.*
  FROM listens l
  JOIN tracks t ON l.track_id = t.id
  WHERE t.release_id = $4
),
bucketed_listens AS (
  SELECT
    b.bucket_start,
    COUNT(l.listened_at) AS listen_count
  FROM buckets b
  LEFT JOIN filtered_listens l
    ON l.listened_at >= b.bucket_start
    AND l.listened_at < b.bucket_start + $3::interval
  GROUP BY b.bucket_start
  ORDER BY b.bucket_start
)
SELECT * FROM bucketed_listens;

-- name: ListenActivityForTrack :many
WITH buckets AS (
  SELECT generate_series($1::timestamptz, $2::timestamptz, $3::interval) AS bucket_start
),
filtered_listens AS (
  SELECT l.*
  FROM listens l
  JOIN tracks t ON l.track_id = t.id
  WHERE t.id = $4
),
bucketed_listens AS (
  SELECT
    b.bucket_start,
    COUNT(l.listened_at) AS listen_count
  FROM buckets b
  LEFT JOIN filtered_listens l
    ON l.listened_at >= b.bucket_start
    AND l.listened_at < b.bucket_start + $3::interval
  GROUP BY b.bucket_start
  ORDER BY b.bucket_start
)
SELECT * FROM bucketed_listens;

-- name: UpdateTrackIdForListens :exec
UPDATE listens SET track_id = $2
WHERE track_id = $1;

-- name: DeleteListen :exec
DELETE FROM listens WHERE track_id = $1 AND listened_at = $2;

-- name: GetListensExportPage :many
SELECT
    l.listened_at,
    l.user_id,
    l.client,

    -- Track info
    t.id AS track_id,
    t.musicbrainz_id AS track_mbid,
    t.duration AS track_duration,
    (
        SELECT json_agg(json_build_object(
            'alias', ta.alias,
            'source', ta.source,
            'is_primary', ta.is_primary
        ))
        FROM track_aliases ta
        WHERE ta.track_id = t.id
    ) AS track_aliases,

    -- Release info
    r.id AS release_id,
    r.musicbrainz_id AS release_mbid,
    r.image AS release_image,
    r.image_source AS release_image_source,
    r.various_artists,
    (
        SELECT json_agg(json_build_object(
            'alias', ra.alias,
            'source', ra.source,
            'is_primary', ra.is_primary
        ))
        FROM release_aliases ra
        WHERE ra.release_id = r.id
    ) AS release_aliases,

    -- Artists
    (
        SELECT json_agg(json_build_object(
            'id', a.id,
            'musicbrainz_id', a.musicbrainz_id,
            'image', a.image,
            'image_source', a.image_source,
            'aliases', (
                SELECT json_agg(json_build_object(
                    'alias', aa.alias,
                    'source', aa.source,
                    'is_primary', aa.is_primary
                ))
                FROM artist_aliases aa
                WHERE aa.artist_id = a.id
            )
        ))
        FROM artist_tracks at
        JOIN artists a ON a.id = at.artist_id
        WHERE at.track_id = t.id
    ) AS artists

FROM listens l
JOIN tracks t ON l.track_id = t.id
JOIN releases r ON t.release_id = r.id

WHERE l.user_id = @user_id::int
  AND (l.listened_at, l.track_id) > (@listened_at::timestamptz, @track_id::int)
ORDER BY l.listened_at, l.track_id
LIMIT $1;
