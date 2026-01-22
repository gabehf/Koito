-- name: GetGroupedListensFromArtist :many
WITH bounds AS (
    SELECT
        MIN(l.listened_at) AS start_time,
        NOW() AS end_time
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    JOIN artist_tracks at ON at.track_id = t.id
    WHERE at.artist_id = $1
),
stats AS (
    SELECT
        start_time,
        end_time,
        EXTRACT(EPOCH FROM (end_time - start_time)) AS total_seconds,
        ((end_time - start_time) / sqlc.arg(bucket_count)::int) AS bucket_interval
    FROM bounds
),
bucket_series AS (
    SELECT generate_series(0, sqlc.arg(bucket_count)::int - 1) AS idx
),
listen_indices AS (
    SELECT
        LEAST(
            sqlc.arg(bucket_count)::int - 1,
            FLOOR(
                (EXTRACT(EPOCH FROM (l.listened_at - s.start_time)) / NULLIF(s.total_seconds, 0))
                * sqlc.arg(bucket_count)::int
            )::int
        ) AS bucket_idx
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    JOIN artist_tracks at ON at.track_id = t.id
    CROSS JOIN stats s
    WHERE at.artist_id = $1
      AND s.start_time IS NOT NULL
)
SELECT
    (s.start_time + (s.bucket_interval * bs.idx))::timestamptz AS bucket_start,
    (s.start_time + (s.bucket_interval * (bs.idx + 1)))::timestamptz AS bucket_end,
    COUNT(li.bucket_idx) AS listen_count
FROM bucket_series bs
CROSS JOIN stats s
LEFT JOIN listen_indices li ON bs.idx = li.bucket_idx
WHERE s.start_time IS NOT NULL
GROUP BY bs.idx, s.start_time, s.bucket_interval
ORDER BY bs.idx;

-- name: GetGroupedListensFromRelease :many
WITH bounds AS (
    SELECT
        MIN(l.listened_at) AS start_time,
        NOW() AS end_time
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    WHERE t.release_id = $1
),
stats AS (
    SELECT
        start_time,
        end_time,
        EXTRACT(EPOCH FROM (end_time - start_time)) AS total_seconds,
        ((end_time - start_time) / sqlc.arg(bucket_count)::int) AS bucket_interval
    FROM bounds
),
bucket_series AS (
    SELECT generate_series(0, sqlc.arg(bucket_count)::int - 1) AS idx
),
listen_indices AS (
    SELECT
        LEAST(
            sqlc.arg(bucket_count)::int - 1,
            FLOOR(
                (EXTRACT(EPOCH FROM (l.listened_at - s.start_time)) / NULLIF(s.total_seconds, 0))
                * sqlc.arg(bucket_count)::int
            )::int
        ) AS bucket_idx
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    CROSS JOIN stats s
    WHERE t.release_id = $1
      AND s.start_time IS NOT NULL
)
SELECT
    (s.start_time + (s.bucket_interval * bs.idx))::timestamptz AS bucket_start,
    (s.start_time + (s.bucket_interval * (bs.idx + 1)))::timestamptz AS bucket_end,
    COUNT(li.bucket_idx) AS listen_count
FROM bucket_series bs
CROSS JOIN stats s
LEFT JOIN listen_indices li ON bs.idx = li.bucket_idx
WHERE s.start_time IS NOT NULL
GROUP BY bs.idx, s.start_time, s.bucket_interval
ORDER BY bs.idx;

-- name: GetGroupedListensFromTrack :many
WITH bounds AS (
    SELECT
        MIN(l.listened_at) AS start_time,
        NOW() AS end_time
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    WHERE t.id = $1
),
stats AS (
    SELECT
        start_time,
        end_time,
        EXTRACT(EPOCH FROM (end_time - start_time)) AS total_seconds,
        ((end_time - start_time) / sqlc.arg(bucket_count)::int) AS bucket_interval
    FROM bounds
),
bucket_series AS (
    SELECT generate_series(0, sqlc.arg(bucket_count)::int - 1) AS idx
),
listen_indices AS (
    SELECT
        LEAST(
            sqlc.arg(bucket_count)::int - 1,
            FLOOR(
                (EXTRACT(EPOCH FROM (l.listened_at - s.start_time)) / NULLIF(s.total_seconds, 0))
                * sqlc.arg(bucket_count)::int
            )::int
        ) AS bucket_idx
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    CROSS JOIN stats s
    WHERE t.id = $1
      AND s.start_time IS NOT NULL
)
SELECT
    (s.start_time + (s.bucket_interval * bs.idx))::timestamptz AS bucket_start,
    (s.start_time + (s.bucket_interval * (bs.idx + 1)))::timestamptz AS bucket_end,
    COUNT(li.bucket_idx) AS listen_count
FROM bucket_series bs
CROSS JOIN stats s
LEFT JOIN listen_indices li ON bs.idx = li.bucket_idx
WHERE s.start_time IS NOT NULL
GROUP BY bs.idx, s.start_time, s.bucket_interval
ORDER BY bs.idx;
