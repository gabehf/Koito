-- name: GetMostReplayedTrackInYear :one
WITH ordered_listens AS (
    SELECT
        user_id,
        track_id,
        listened_at,
        ROW_NUMBER() OVER (PARTITION BY user_id ORDER BY listened_at) AS rn
    FROM listens
    WHERE EXTRACT(YEAR FROM listened_at) = @year::int
),
streaks AS (
    SELECT
        user_id,
        track_id,
        listened_at,
        rn,
        ROW_NUMBER() OVER (PARTITION BY user_id, track_id ORDER BY listened_at) AS track_rn
    FROM ordered_listens
),
grouped_streaks AS (
    SELECT
        user_id,
        track_id,
        rn - track_rn AS group_id,
        COUNT(*) AS streak_length
    FROM streaks
    GROUP BY user_id, track_id, rn - track_rn
),
ranked_streaks AS (
    SELECT *,
        RANK() OVER (PARTITION BY user_id ORDER BY streak_length DESC) AS r
    FROM grouped_streaks
)
SELECT
    t.*, 
    get_artists_for_track(t.id) as artists,
    streak_length
FROM ranked_streaks rs JOIN tracks_with_title t ON rs.track_id = t.id
WHERE user_id = @user_id::int AND r = 1;

-- name: TracksOnlyPlayedOnceInYear :many
SELECT
    t.id AS track_id,
    t.title,
    get_artists_for_track(t.id) as artists,
    COUNT(l.*) AS listen_count
FROM listens l
JOIN tracks_with_title t ON t.id = l.track_id
WHERE EXTRACT(YEAR FROM l.listened_at) = @year::int AND l.user_id = @user_id::int
GROUP BY t.id, t.title
HAVING COUNT(*) = 1
LIMIT $1;

-- name: ArtistsOnlyPlayedOnceInYear :many
SELECT
    a.id AS artist_id,
    a.name,
    COUNT(l.*) AS listen_count
FROM listens l
JOIN artist_tracks at ON at.track_id = l.track_id
JOIN artists_with_name a ON a.id = at.artist_id
WHERE EXTRACT(YEAR FROM l.listened_at) = @year::int AND l.user_id = @user_id::int
GROUP BY a.id, a.name
HAVING COUNT(*) = 1;

-- GetNewTrackWithMostListensInYear :one 
WITH first_plays_in_year AS (
    SELECT
        l.user_id,
        l.track_id,
        MIN(l.listened_at) AS first_listen
    FROM listens l
    WHERE EXTRACT(YEAR FROM l.listened_at) = @year::int
    AND NOT EXISTS (
        SELECT 1
        FROM listens l2
        WHERE l2.user_id = l.user_id
          AND l2.track_id = l.track_id
          AND l2.listened_at < @first_day_of_year::date
    )
    GROUP BY l.user_id, l.track_id
),
seven_day_window AS (
    SELECT
        f.user_id,
        f.track_id,
        f.first_listen,
        COUNT(l.*) AS plays_in_7_days
    FROM first_plays_in_year f
    JOIN listens l
        ON l.user_id = f.user_id
        AND l.track_id = f.track_id
        AND l.listened_at >= f.first_listen
        AND l.listened_at < f.first_listen + INTERVAL '7 days'
    GROUP BY f.user_id, f.track_id, f.first_listen
),
ranked AS (
    SELECT *,
        RANK() OVER (PARTITION BY user_id ORDER BY plays_in_7_days DESC) AS r
    FROM seven_day_window
)
SELECT
    s.user_id,
    s.track_id,
    t.title,
    get_artists_for_track(t.id) as artists,
    s.first_listen,
    s.plays_in_7_days
FROM ranked s
JOIN tracks_with_title t ON t.id = s.track_id
WHERE r = 1;

-- GetTopThreeNewArtistsInYear :many
WITH first_artist_plays_in_year AS (
    SELECT
        l.user_id,
        at.artist_id,
        MIN(l.listened_at) AS first_listen
    FROM listens l
    JOIN artist_tracks at ON at.track_id = l.track_id
    WHERE EXTRACT(YEAR FROM l.listened_at) = @year::int
      AND NOT EXISTS (
          SELECT 1
          FROM listens l2
          JOIN artist_tracks at2 ON at2.track_id = l2.track_id
          WHERE l2.user_id = l.user_id
            AND at2.artist_id = at.artist_id
            AND l2.listened_at < @first_day_of_year::date
      )
    GROUP BY l.user_id, at.artist_id
),
artist_plays_in_year AS (
    SELECT
        f.user_id,
        f.artist_id,
        f.first_listen,
        COUNT(l.*) AS total_plays_in_year
    FROM first_artist_plays_in_year f
    JOIN listens l ON l.user_id = f.user_id
    JOIN artist_tracks at ON at.track_id = l.track_id
    WHERE at.artist_id = f.artist_id
      AND EXTRACT(YEAR FROM l.listened_at) = @year::int
    GROUP BY f.user_id, f.artist_id, f.first_listen
),
ranked AS (
    SELECT *,
        RANK() OVER (PARTITION BY user_id ORDER BY total_plays_in_year DESC) AS r
    FROM artist_plays_in_year
)
SELECT
    a.user_id,
    a.artist_id,
    awn.name AS artist_name,
    a.first_listen,
    a.total_plays_in_year
FROM ranked a
JOIN artists_with_name awn ON awn.id = a.artist_id
WHERE r <= 3;

-- name: GetArtistWithLongestGapInYear :one
WITH first_listens AS (
    SELECT
        l.user_id,
        at.artist_id,
        MIN(l.listened_at::date) AS first_listen_of_year
    FROM listens l
    JOIN artist_tracks at ON at.track_id = l.track_id
    WHERE EXTRACT(YEAR FROM l.listened_at) = @year::int
    GROUP BY l.user_id, at.artist_id
),
last_listens AS (
    SELECT
        l.user_id,
        at.artist_id,
        MAX(l.listened_at::date) AS last_listen
    FROM listens l
    JOIN artist_tracks at ON at.track_id = l.track_id
    WHERE l.listened_at < @first_day_of_year::date
    GROUP BY l.user_id, at.artist_id
),
comebacks AS (
    SELECT
        f.user_id,
        f.artist_id,
        f.first_listen_of_year,
        p.last_listen,
        (f.first_listen_of_year - p.last_listen) AS gap_days
    FROM first_listens f
    JOIN last_listens p
      ON f.user_id = p.user_id AND f.artist_id = p.artist_id
),
ranked AS (
    SELECT *,
        RANK() OVER (PARTITION BY user_id ORDER BY gap_days DESC) AS r
    FROM comebacks
)
SELECT
    c.user_id,
    c.artist_id,
    awn.name AS artist_name,
    c.last_listen,
    c.first_listen_of_year,
    c.gap_days
FROM ranked c
JOIN artists_with_name awn ON awn.id = c.artist_id
WHERE r = 1;

-- name: GetFirstListenInYear :one
SELECT 
    l.*, 
    t.*, 
    get_artists_for_track(t.id) as artists 
FROM listens l 
LEFT JOIN tracks_with_title t ON l.track_id = t.id 
WHERE EXTRACT(YEAR FROM l.listened_at) = 2025 
ORDER BY l.listened_at ASC 
LIMIT 1;

-- name: GetTracksPlayedAtLeastOncePerMonthInYear :many
WITH monthly_plays AS (
    SELECT
        l.track_id,
        EXTRACT(MONTH FROM l.listened_at) AS month
    FROM listens l
    WHERE EXTRACT(YEAR FROM l.listened_at) = @user_id::int
    GROUP BY l.track_id, EXTRACT(MONTH FROM l.listened_at)
),
monthly_counts AS (
    SELECT
        track_id,
        COUNT(DISTINCT month) AS months_played
    FROM monthly_plays
    GROUP BY track_id
)
SELECT
    t.id AS track_id,
    t.title
FROM monthly_counts mc
JOIN tracks_with_title t ON t.id = mc.track_id
WHERE mc.months_played = 12;

-- name: GetWeekWithMostListensInYear :one
SELECT
    DATE_TRUNC('week', listened_at + INTERVAL '1 day') - INTERVAL '1 day' AS week_start,
    COUNT(*) AS listen_count
FROM listens 
WHERE EXTRACT(YEAR FROM listened_at) = @year::int 
    AND user_id = @user_id::int
GROUP BY week_start
ORDER BY listen_count DESC
LIMIT 1;

-- name: GetPercentageOfTotalListensFromTopTracksInYear :one
WITH user_listens AS (
    SELECT
        l.track_id,
        COUNT(*) AS listen_count
    FROM listens l
    WHERE l.user_id = @user_id::int
      AND EXTRACT(YEAR FROM l.listened_at) = @year::int
    GROUP BY l.track_id
),
top_tracks AS (
    SELECT
        track_id,
        listen_count
    FROM user_listens
    ORDER BY listen_count DESC
    LIMIT $1
),
totals AS (
    SELECT
        (SELECT SUM(listen_count) FROM top_tracks) AS top_tracks_total,
        (SELECT SUM(listen_count) FROM user_listens) AS overall_total
)
SELECT
    top_tracks_total,
    overall_total,
    ROUND((top_tracks_total::decimal / overall_total) * 100, 2) AS percent_of_total
FROM totals;

-- name: GetPercentageOfTotalListensFromTopArtistsInYear :one
WITH user_artist_listens AS (
    SELECT
        at.artist_id,
        COUNT(*) AS listen_count
    FROM listens l
    JOIN artist_tracks at ON at.track_id = l.track_id
    WHERE l.user_id = @user_id::int
      AND EXTRACT(YEAR FROM l.listened_at) = @year::int
    GROUP BY at.artist_id
),
top_artists AS (
    SELECT
        artist_id,
        listen_count
    FROM user_artist_listens
    ORDER BY listen_count DESC
    LIMIT $1
),
totals AS (
    SELECT
        (SELECT SUM(listen_count) FROM top_artists) AS top_artist_total,
        (SELECT SUM(listen_count) FROM user_artist_listens) AS overall_total
)
SELECT
    top_artist_total,
    overall_total,
    ROUND((top_artist_total::decimal / overall_total) * 100, 2) AS percent_of_total
FROM totals;

-- name: GetArtistsWithOnlyOnePlayInYear :many
WITH first_artist_plays_in_year AS (
    SELECT
        l.user_id,
        at.artist_id,
        MIN(l.listened_at) AS first_listen
    FROM listens l
    JOIN artist_tracks at ON at.track_id = l.track_id
    WHERE EXTRACT(YEAR FROM l.listened_at) = 2024
      AND NOT EXISTS (
          SELECT 1
          FROM listens l2
          JOIN artist_tracks at2 ON at2.track_id = l2.track_id
          WHERE l2.user_id = l.user_id
            AND at2.artist_id = at.artist_id
            AND l2.listened_at < DATE '2024-01-01'
      )
    GROUP BY l.user_id, at.artist_id
) 
SELECT
    f.user_id,
    f.artist_id,
    f.first_listen, a.name,
    COUNT(l.*) AS total_plays_in_year
FROM first_artist_plays_in_year f
JOIN listens l ON l.user_id = f.user_id
JOIN artist_tracks at ON at.track_id = l.track_id JOIN artists_with_name a ON at.artist_id = a.id
WHERE at.artist_id = f.artist_id
  AND EXTRACT(YEAR FROM l.listened_at) = 2024
GROUP BY f.user_id, f.artist_id, f.first_listen, a.name HAVING COUNT(*) = 1;

-- name: GetArtistCountInYear :one
SELECT
    COUNT(DISTINCT at.artist_id) AS artist_count
FROM listens l
JOIN artist_tracks at ON at.track_id = l.track_id
WHERE l.user_id = @user_id::int
  AND EXTRACT(YEAR FROM l.listened_at) = @year::int;

-- name: GetListenPercentageInTimeWindowInYear :one
WITH user_listens_in_year AS (
    SELECT
        listened_at
    FROM listens
    WHERE user_id = @user_id::int
      AND EXTRACT(YEAR FROM listened_at) = @year::int
),
windowed AS (
    SELECT
        COUNT(*) AS in_window
    FROM user_listens_in_year
    WHERE EXTRACT(HOUR FROM listened_at) >= @hour_window_start::int
      AND EXTRACT(HOUR FROM listened_at) < @hour_window_end::int
),
total AS (
    SELECT COUNT(*) AS total_listens
    FROM user_listens_in_year
)
SELECT
    w.in_window,
    t.total_listens,
    ROUND((w.in_window::decimal / t.total_listens) * 100, 2) AS percent_of_total
FROM windowed w, total t;