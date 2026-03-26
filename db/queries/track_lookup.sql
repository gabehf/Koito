-- name: GetTrackLookup :one
SELECT artist_id, album_id, track_id
FROM track_lookup
WHERE lookup_key = $1;

-- name: InsertTrackLookup :exec
INSERT INTO track_lookup (lookup_key, artist_id, album_id, track_id)
VALUES ($1, $2, $3, $4)
ON CONFLICT (lookup_key) DO UPDATE SET
    artist_id = EXCLUDED.artist_id,
    album_id = EXCLUDED.album_id,
    track_id = EXCLUDED.track_id;

-- name: DeleteTrackLookupByArtist :exec
DELETE FROM track_lookup WHERE artist_id = $1;

-- name: DeleteTrackLookupByAlbum :exec
DELETE FROM track_lookup WHERE album_id = $1;

-- name: DeleteTrackLookupByTrack :exec
DELETE FROM track_lookup WHERE track_id = $1;
