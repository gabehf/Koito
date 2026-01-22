-- +goose Up
DELETE FROM artist_releases ar
WHERE NOT EXISTS (
    SELECT 1
    FROM artist_tracks at
    JOIN tracks t ON at.track_id = t.id
    WHERE at.artist_id = ar.artist_id
      AND t.release_id = ar.release_id
);
