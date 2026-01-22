-- name: CleanOrphanedEntries :exec
DO $$
BEGIN
  DELETE FROM tracks WHERE id NOT IN (SELECT l.track_id FROM listens l);
  DELETE FROM releases WHERE id NOT IN (SELECT t.release_id FROM tracks t);
  DELETE FROM artists WHERE id NOT IN (SELECT at.artist_id FROM artist_tracks at);
  DELETE FROM artist_releases ar
  WHERE NOT EXISTS (
      SELECT 1
      FROM artist_tracks at
      JOIN tracks t ON at.track_id = t.id
      WHERE at.artist_id = ar.artist_id
        AND t.release_id = ar.release_id
  );
END $$;
