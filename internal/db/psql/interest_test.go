package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// an llm wrote this because i didn't feel like it. it looks like it works, although
// it could stand to be more thorough
func TestGetInterest(t *testing.T) {
	truncateTestData(t)

	ctx := context.Background()

	// --- Setup Data ---

	// Insert Artists
	err := store.Exec(ctx, `
		INSERT INTO artists (musicbrainz_id)
		VALUES ('00000000-0000-0000-0000-000000000001'),
		       ('00000000-0000-0000-0000-000000000002')`)
	require.NoError(t, err)

	// Insert Releases (Albums)
	err = store.Exec(ctx, `
		INSERT INTO releases (musicbrainz_id)
		VALUES ('00000000-0000-0000-0000-000000000011')`)
	require.NoError(t, err)

	// Insert Tracks (Both on Release 1)
	err = store.Exec(ctx, `
		INSERT INTO tracks (musicbrainz_id, release_id)
		VALUES ('11111111-1111-1111-1111-111111111111', 1),
		       ('22222222-2222-2222-2222-222222222222', 1)`)
	require.NoError(t, err)

	// Link Artists to Tracks
	// Artist 1 -> Track 1
	// Artist 2 -> Track 2
	err = store.Exec(ctx, `
		INSERT INTO artist_tracks (artist_id, track_id)
		VALUES (1, 1), (2, 2)`)
	require.NoError(t, err)

	// Insert Listens
	// Track 1 (Artist 1, Release 1): 3 Listens
	// Track 2 (Artist 2, Release 1): 2 Listens
	err = store.Exec(ctx, `
		INSERT INTO listens (user_id, track_id, listened_at) VALUES
		(1, 1, NOW() - INTERVAL '1 hour'),
		(1, 1, NOW() - INTERVAL '2 hours'),
		(1, 1, NOW() - INTERVAL '3 hours'),
		(1, 2, NOW() - INTERVAL '1 hour'),
		(1, 2, NOW() - INTERVAL '2 hours')
	`)
	require.NoError(t, err)

	// --- Test Validation ---

	t.Run("Validation", func(t *testing.T) {
		// Error: Missing Buckets
		_, err := store.GetInterest(ctx, db.GetInterestOpts{ArtistID: 1})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "bucket count must be provided")

		// Error: Missing ID
		_, err = store.GetInterest(ctx, db.GetInterestOpts{Buckets: 10})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be provided")
	})

	// --- Test Data Retrieval ---
	// Note: We use Buckets: 1 to ensure all listens are aggregated into a single result
	// for easier assertion, avoiding complex date/time math in the test.

	t.Run("Artist Interest", func(t *testing.T) {
		// Artist 1 should have 3 listens (from Track 1)
		buckets, err := store.GetInterest(ctx, db.GetInterestOpts{
			ArtistID: 1,
			Buckets:  1,
		})
		require.NoError(t, err)
		require.Len(t, buckets, 1)
		assert.EqualValues(t, 3, buckets[0].ListenCount, "Artist 1 should have 3 listens")
	})

	t.Run("Album Interest", func(t *testing.T) {
		// Album 1 contains Track 1 (3 listens) and Track 2 (2 listens) = 5 Total
		buckets, err := store.GetInterest(ctx, db.GetInterestOpts{
			AlbumID: 1,
			Buckets: 1,
		})
		require.NoError(t, err)
		require.Len(t, buckets, 1)
		assert.EqualValues(t, 5, buckets[0].ListenCount, "Album 1 should have 5 listens total")
	})

	t.Run("Track Interest", func(t *testing.T) {
		// Track 2 should have 2 listens
		buckets, err := store.GetInterest(ctx, db.GetInterestOpts{
			TrackID: 2,
			Buckets: 1,
		})
		require.NoError(t, err)
		require.Len(t, buckets, 1)
		assert.EqualValues(t, 2, buckets[0].ListenCount, "Track 2 should have 2 listens")
	})
}
