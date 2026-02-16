package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDataForTracks(t *testing.T) {
	truncateTestData(t)

	// Insert artists
	err := store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000001'),
				   ('00000000-0000-0000-0000-000000000002')`)
	require.NoError(t, err)

	// Insert artist aliases
	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary)
			VALUES (1, 'Artist One', 'Testing', true),
				   (2, 'Artist Two', 'Testing', true)`)
	require.NoError(t, err)

	// Insert release groups
	err = store.Exec(context.Background(),
		`INSERT INTO releases (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000011'),
				   ('00000000-0000-0000-0000-000000000022')`)
	require.NoError(t, err)

	// Associate releases with artists
	err = store.Exec(context.Background(),
		`INSERT INTO artist_releases (artist_id, release_id, is_primary)
     VALUES (1, 1, true),
            (2, 2, true)`)
	require.NoError(t, err)

	// Insert release aliases
	err = store.Exec(context.Background(),
		`INSERT INTO release_aliases (release_id, alias, source, is_primary)
			VALUES (1, 'Release Group One', 'Testing', true),
				   (2, 'Release Group Two', 'Testing', true)`)
	require.NoError(t, err)

	// Insert tracks
	err = store.Exec(context.Background(),
		`INSERT INTO tracks (musicbrainz_id, release_id, duration)
			VALUES ('11111111-1111-1111-1111-111111111111', 1, 100),
				   ('22222222-2222-2222-2222-222222222222', 2, 100)`)
	require.NoError(t, err)

	// Insert track aliases
	err = store.Exec(context.Background(),
		`INSERT INTO track_aliases (track_id, alias, source, is_primary)
			VALUES (1, 'Track One', 'Testing', true),
				   (2, 'Track Two', 'Testing', true)`)
	require.NoError(t, err)

	// Associate tracks with artists
	err = store.Exec(context.Background(),
		`INSERT INTO artist_tracks (artist_id, track_id, is_primary)
			VALUES (1, 1, true), (2, 2, true)`)
	require.NoError(t, err)

	// Insert listens
	err = store.Exec(context.Background(),
		`INSERT INTO listens (user_id, track_id, listened_at)
			VALUES (1, 1, NOW()), (1, 2, NOW())`)
	require.NoError(t, err)
}

func TestGetTrack(t *testing.T) {
	testDataForTracks(t)
	ctx := context.Background()

	// Test GetTrack by ID
	track, err := store.GetTrack(ctx, db.GetTrackOpts{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, int32(1), track.ID)
	assert.Equal(t, "Track One", track.Title)
	assert.Equal(t, uuid.MustParse("11111111-1111-1111-1111-111111111111"), *track.MbzID)
	assert.EqualValues(t, 100, track.TimeListened)

	// Test GetTrack by MusicBrainzID
	track, err = store.GetTrack(ctx, db.GetTrackOpts{MusicBrainzID: uuid.MustParse("22222222-2222-2222-2222-222222222222")})
	require.NoError(t, err)
	assert.Equal(t, int32(2), track.ID)
	assert.Equal(t, "Track Two", track.Title)
	assert.EqualValues(t, 100, track.TimeListened)

	// Test GetTrack by Title, Release and ArtistIDs
	track, err = store.GetTrack(ctx, db.GetTrackOpts{
		Title:     "Track One",
		ReleaseID: 1,
		ArtistIDs: []int32{1},
	})
	require.NoError(t, err)
	assert.Equal(t, int32(1), track.ID)
	assert.Equal(t, "Track One", track.Title)
	assert.EqualValues(t, 100, track.TimeListened)

	// Test GetTrack with insufficient information
	_, err = store.GetTrack(ctx, db.GetTrackOpts{Title: "Track One"})
	assert.Error(t, err)
}
func TestSaveTrack(t *testing.T) {
	testDataForTracks(t)
	ctx := context.Background()

	// Test SaveTrack with valid inputs
	track, err := store.SaveTrack(ctx, db.SaveTrackOpts{
		Title:          "New Track",
		ArtistIDs:      []int32{1},
		RecordingMbzID: uuid.MustParse("33333333-3333-3333-3333-333333333333"),
		AlbumID:        1,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Track", track.Title)
	assert.Equal(t, uuid.MustParse("33333333-3333-3333-3333-333333333333"), *track.MbzID)

	// Verify artist associations exist
	exists, err := store.RowExists(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM artist_tracks
			WHERE artist_id = $1 AND track_id = $2
		)`, 1, track.ID)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist association to exist")

	// Verify alias exists
	exists, err = store.RowExists(ctx, `
	SELECT EXISTS (
		SELECT 1 FROM track_aliases
		WHERE track_id = $1 AND is_primary = true
	)`, track.ID)
	require.NoError(t, err)
	assert.True(t, exists, "expected primary alias to exist")

	// Test SaveTrack with missing ArtistIDs
	_, err = store.SaveTrack(ctx, db.SaveTrackOpts{
		Title:          "Invalid Track",
		ArtistIDs:      []int32{},
		RecordingMbzID: uuid.MustParse("44444444-4444-4444-4444-444444444444"),
	})
	assert.Error(t, err)

	// Test SaveTrack with invalid ArtistIDs
	_, err = store.SaveTrack(ctx, db.SaveTrackOpts{
		Title:          "Invalid Track",
		ArtistIDs:      []int32{0},
		RecordingMbzID: uuid.MustParse("55555555-5555-5555-5555-555555555555"),
	})
	assert.Error(t, err)
}

func TestUpdateTrack(t *testing.T) {
	testDataForTracks(t)
	ctx := context.Background()

	newMbzID := uuid.MustParse("66666666-6666-6666-6666-666666666666")
	newDuration := 100
	err := store.UpdateTrack(ctx, db.UpdateTrackOpts{
		ID:            1,
		MusicBrainzID: newMbzID,
		Duration:      int32(newDuration),
	})
	require.NoError(t, err)

	// Verify the update
	track, err := store.GetTrack(ctx, db.GetTrackOpts{ID: 1})
	require.NoError(t, err)
	require.Equal(t, newMbzID, *track.MbzID)
	require.EqualValues(t, newDuration, track.Duration)

	// Test UpdateTrack with missing ID
	err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
		ID:            0,
		MusicBrainzID: newMbzID,
		Duration:      int32(newDuration),
	})
	assert.Error(t, err)

	// Test UpdateTrack with nil MusicBrainz ID
	err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
		ID:            1,
		MusicBrainzID: uuid.Nil,
		Duration:      int32(newDuration),
	})
	assert.NoError(t, err) // No update should occur

	// Test adding artist to track
	err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
		ID:         1,
		AddArtists: []int32{2},
	})
	assert.NoError(t, err)

	exists, err := store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artist_tracks
            WHERE artist_id = $1 AND track_id = $2
        )`, 2, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist to be associated with track")

	// Test removing artist from track
	err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
		ID:            1,
		RemoveArtists: []int32{2},
	})
	assert.NoError(t, err)

	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artist_tracks
            WHERE artist_id = $1 AND track_id = $2
        )`, 2, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected artist to be unassociated with track")

	// Test that the primary artist cannot be removed
	err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
		ID:            1,
		RemoveArtists: []int32{int32(1)},
	})
	assert.NoError(t, err)
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artist_tracks
            WHERE artist_id = $1 AND track_id = $2
        )`, 1, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected primary artist to not be removed from track")
}

func TestTrackAliases(t *testing.T) {
	testDataForTracks(t)
	ctx := context.Background()

	err := store.SaveTrackAliases(ctx, 1, []string{"Alias One", "Alias Two"}, "Testing")
	require.NoError(t, err)
	aliases, err := store.GetAllTrackAliases(ctx, 1)
	require.NoError(t, err)
	assert.Len(t, aliases, 3)

	err = store.SetPrimaryTrackAlias(ctx, 1, "Alias One")
	require.NoError(t, err)
	track, err := store.GetTrack(ctx, db.GetTrackOpts{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Alias One", track.Title)

	err = store.SetPrimaryTrackAlias(ctx, 1, "Fake Alias")
	require.Error(t, err)

	// Ensure primary alias cannot be deleted
	err = store.DeleteTrackAlias(ctx, track.ID, "Alias One")
	require.NoError(t, err) // shouldn't error when nothing is deleted
	track, err = store.GetTrack(ctx, db.GetTrackOpts{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Alias One", track.Title)

	store.SetPrimaryTrackAlias(ctx, 1, "Track One")
}

func TestDeleteTrack(t *testing.T) {
	testDataForTracks(t)
	ctx := context.Background()

	err := store.DeleteTrack(ctx, 2)
	require.NoError(t, err)

	_, err = store.Count(ctx, `SELECT * FROM tracks WHERE id = 2`)
	require.ErrorIs(t, err, pgx.ErrNoRows) // no rows error
}

func TestReleaseAssociations(t *testing.T) {
	testDataForTracks(t)
	ctx := context.Background()

	track, err := store.SaveTrack(ctx, db.SaveTrackOpts{
		Title:     "Track Three",
		AlbumID:   2,
		ArtistIDs: []int32{2, 1}, // Artist Two feat. Artist One
		Duration:  100,
	})
	require.NoError(t, err)
	count, err := store.Count(ctx, `SELECT COUNT(*) FROM artist_releases WHERE release_id = 2`)
	require.NoError(t, err)
	require.Equal(t, 2, count, "expected release to be associated with artist from inserted track")

	err = store.DeleteTrack(ctx, track.ID)
	require.NoError(t, err)

	count, err = store.Count(ctx, `SELECT COUNT(*) FROM artist_releases WHERE release_id = 2`)
	require.NoError(t, err)
	require.Equal(t, 1, count, "expected artist no longer on release to be disassociated from release")

}
