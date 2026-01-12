-- name: GetGroupedListensFromArtist :many
WITH artist_listens AS (
    SELECT
        l.listened_at
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    JOIN artist_tracks at ON at.track_id = t.id
    WHERE at.artist_id = $1
),
bounds AS (
    SELECT
        MIN(listened_at) AS start_time,
        MAX(listened_at) AS end_time
    FROM artist_listens
),
bucketed AS (
    SELECT
        LEAST(
            sqlc.arg(bucket_count) - 1,
            FLOOR(
                (
                    EXTRACT(EPOCH FROM (al.listened_at - b.start_time))
                    /
                    NULLIF(EXTRACT(EPOCH FROM (b.end_time - b.start_time)), 0)
                ) * sqlc.arg(bucket_count)
            )::int
        ) AS bucket_idx,
        b.start_time,
        b.end_time
    FROM artist_listens al
    CROSS JOIN bounds b
),
aggregated AS (
    SELECT
        start_time
            + (
                bucket_idx * (end_time - start_time)
                / sqlc.arg(bucket_count)
              ) AS bucket_start,
        start_time
            + (
                (bucket_idx + 1) * (end_time - start_time)
                / sqlc.arg(bucket_count)
              ) AS bucket_end,
        COUNT(*) AS listen_count
    FROM bucketed
    GROUP BY bucket_idx, start_time, end_time
)
SELECT
    bucket_start::timestamptz,
    bucket_end::timestamptz,
    listen_count
FROM aggregated
ORDER BY bucket_start;

-- name: GetGroupedListensFromRelease :many
WITH artist_listens AS (
    SELECT
        l.listened_at
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    WHERE t.release_id = $1
),
bounds AS (
    SELECT
        MIN(listened_at) AS start_time,
        MAX(listened_at) AS end_time
    FROM artist_listens
),
bucketed AS (
    SELECT
        LEAST(
            sqlc.arg(bucket_count) - 1,
            FLOOR(
                (
                    EXTRACT(EPOCH FROM (al.listened_at - b.start_time))
                    /
                    NULLIF(EXTRACT(EPOCH FROM (b.end_time - b.start_time)), 0)
                ) * sqlc.arg(bucket_count)
            )::int
        ) AS bucket_idx,
        b.start_time,
        b.end_time
    FROM artist_listens al
    CROSS JOIN bounds b
),
aggregated AS (
    SELECT
        start_time
            + (
                bucket_idx * (end_time - start_time)
                / sqlc.arg(bucket_count)
              ) AS bucket_start,
        start_time
            + (
                (bucket_idx + 1) * (end_time - start_time)
                / sqlc.arg(bucket_count)
              ) AS bucket_end,
        COUNT(*) AS listen_count
    FROM bucketed
    GROUP BY bucket_idx, start_time, end_time
)
SELECT
    bucket_start::timestamptz,
    bucket_end::timestamptz,
    listen_count
FROM aggregated
ORDER BY bucket_start;

-- name: GetGroupedListensFromTrack :many
WITH artist_listens AS (
    SELECT
        l.listened_at
    FROM listens l
    JOIN tracks t ON t.id = l.track_id
    WHERE t.id = $1
),
bounds AS (
    SELECT
        MIN(listened_at) AS start_time,
        MAX(listened_at) AS end_time
    FROM artist_listens
),
bucketed AS (
    SELECT
        LEAST(
            sqlc.arg(bucket_count) - 1,
            FLOOR(
                (
                    EXTRACT(EPOCH FROM (al.listened_at - b.start_time))
                    /
                    NULLIF(EXTRACT(EPOCH FROM (b.end_time - b.start_time)), 0)
                ) * sqlc.arg(bucket_count)
            )::int
        ) AS bucket_idx,
        b.start_time,
        b.end_time
    FROM artist_listens al
    CROSS JOIN bounds b
),
aggregated AS (
    SELECT
        start_time
            + (
                bucket_idx * (end_time - start_time)
                / sqlc.arg(bucket_count)
              ) AS bucket_start,
        start_time
            + (
                (bucket_idx + 1) * (end_time - start_time)
                / sqlc.arg(bucket_count)
              ) AS bucket_end,
        COUNT(*) AS listen_count
    FROM bucketed
    GROUP BY bucket_idx, start_time, end_time
)
SELECT
    bucket_start::timestamptz,
    bucket_end::timestamptz,
    listen_count
FROM aggregated
ORDER BY bucket_start;
