package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// this file is very long

func TestSubmitListen_CreateAllMbzIDs(t *testing.T) {
	truncateTestData(t)

	// artist gets created with musicbrainz id
	// release group gets created with mbz id
	// track gets created with mbz id
	// test listen time is opts time

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:       mbzArtistData,
		ReleaseGroups: mbzReleaseGroupData,
		Releases:      mbzReleaseData,
		Tracks:        mbzTrackData,
	}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseGroupMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:        "Tokyo Calling",
		RecordingMbzID:    trackMbzID,
		ReleaseTitle:      "AG! Calling",
		ReleaseMbzID:      releaseMbzID,
		ReleaseGroupMbzID: releaseGroupMbzID,
		Time:              time.Now(),
		UserID:            1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// Verify that listen time is correct
	p, err := store.GetListensPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 1})
	require.NoError(t, err)
	require.Len(t, p.Items, 1)
	l := p.Items[0]
	EqualTime(t, opts.Time.Truncate(time.Second), l.Time)
}

func TestSubmitListen_CreateAllMbzIDsNoReleaseGroupID(t *testing.T) {
	truncateTestData(t)

	// release group gets created with release id

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:       mbzArtistData,
		ReleaseGroups: mbzReleaseGroupData,
		Releases:      mbzReleaseData,
		Tracks:        mbzTrackData,
	}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:     "Tokyo Calling",
		RecordingMbzID: trackMbzID,
		ReleaseTitle:   "AG! Calling",
		ReleaseMbzID:   releaseMbzID,
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM releases_with_title
      WHERE title = $1
    )`, "AG! Calling")
	require.NoError(t, err)
	assert.True(t, exists, "expected release to be created")
}

func TestSubmitListen_CreateAllNoMbzIDs(t *testing.T) {
	truncateTestData(t)

	// artist gets created with artist names
	// release group gets created with artist and title
	// track gets created with title and artist

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{}
	opts := catalog.SubmitListenOpts{
		MbzCaller:    mbzc,
		ArtistNames:  []string{"ATARASHII GAKKO!"},
		Artist:       "ATARASHII GAKKO!",
		TrackTitle:   "Tokyo Calling",
		ReleaseTitle: "AG! Calling",
		Time:         time.Now(),
		UserID:       1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")
}

func TestSubmitListen_CreateAllNoMbzIDsNoArtistNamesNoReleaseTitle(t *testing.T) {
	truncateTestData(t)

	// artists get created with artist and track title
	// release group gets created with artist and track title

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{}
	opts := catalog.SubmitListenOpts{
		MbzCaller: mbzc,
		ArtistMbzIDs: []uuid.UUID{
			uuid.MustParse("00000000-0000-0000-0000-000000000000"),
		},
		Artist:     "Rat Tally",
		TrackTitle: "In My Car feat. Madeline Kenney",
		Time:       time.Now(),
		UserID:     1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM releases_with_title
      WHERE title = $1
    )`, opts.TrackTitle)
	require.NoError(t, err)
	assert.True(t, exists, "expected created release to have track title as title")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artists_with_name
      WHERE name = $1
    )`, "Rat Tally")
	require.NoError(t, err)
	assert.True(t, exists, "expected primary artist to be created")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artists_with_name
      WHERE name = $1
    )`, "Madeline Kenney")
	require.NoError(t, err)
	assert.True(t, exists, "expected featured artist to be created")

	// assert that Rat Tally is the primary artist
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artist_tracks
      WHERE artist_id = $1 AND is_primary = $2
    )`, 1, true)
	require.NoError(t, err)
	assert.True(t, exists, "expected primary artist to be marked as primary for track")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artist_releases
      WHERE artist_id = $1 AND is_primary = $2
    )`, 1, true)
	require.NoError(t, err)
	assert.True(t, exists, "expected primary artist to be marked as primary for release")
}

func TestSubmitListen_MatchAllMbzIDs(t *testing.T) {
	setupTestDataWithMbzIDs(t)

	// artist gets matched with musicbrainz id
	// release gets matched with mbz id
	// track gets matched with mbz id

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:  mbzArtistData,
		Releases: mbzReleaseData,
		Tracks:   mbzTrackData,
	}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:     "Tokyo Calling",
		RecordingMbzID: trackMbzID,
		ReleaseTitle:   "AG! Calling",
		ReleaseMbzID:   releaseMbzID,
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release group created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created")
}

func TestSubmitListen_MatchTrackFromMbzTitle(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Tracks: mbzTrackData,
	}
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:      mbzc,
		ArtistNames:    []string{"ATARASHII GAKKO!"},
		Artist:         "ATARASHII GAKKO!",
		TrackTitle:     "Tokyo Calling - Alt Title",
		RecordingMbzID: trackMbzID,
		ReleaseTitle:   "AG! Calling",
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release group created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created")
}

func TestSubmitListen_VariousArtistsRelease(t *testing.T) {

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Releases: mbzReleaseData,
	}
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000202")
	opts := catalog.SubmitListenOpts{
		MbzCaller:    mbzc,
		ArtistNames:  []string{"ARIANNE"},
		Artist:       "ARIANNE",
		TrackTitle:   "KOMM, SUSSER TOD (M-10 Director's Edit version)",
		ReleaseTitle: "Evangelion Finally",
		ReleaseMbzID: releaseMbzID,
		Time:         time.Now(),
		UserID:       1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM releases WHERE various_artists = $1
	`, true)
	require.NoError(t, err)
	assert.EqualValues(t, 1, count)
}

func TestSubmitListen_MatchOneArtistMbzIDOneArtistName(t *testing.T) {
	setupTestDataWithMbzIDs(t)

	// artist gets matched with musicbrainz id
	// release gets matched with mbz id
	// track gets matched with mbz id

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:  mbzArtistData,
		Releases: mbzReleaseData,
		Tracks:   mbzTrackData,
	}
	// i really do want to use real tracks for tests but i dont wanna set up all the data for one test
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!", "Fake Artist"},
		Artist:      "ATARASHII GAKKO! feat. Fake Artist",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:     "Tokyo Calling",
		RecordingMbzID: trackMbzID,
		ReleaseTitle:   "AG! Calling",
		ReleaseMbzID:   releaseMbzID,
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release group created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "Fake Artist")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "expected featured artist to be created")
}

func TestSubmitListen_MatchAllMbzIDsNoReleaseGroupIDNoTrackID(t *testing.T) {
	setupTestDataWithMbzIDs(t)

	// release group gets matched with release id
	// track gets matched with title and artist

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:       mbzArtistData,
		ReleaseGroups: mbzReleaseGroupData,
		Releases:      mbzReleaseData,
		Tracks:        mbzTrackData,
	}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:   "Tokyo Calling",
		ReleaseTitle: "AG! Calling",
		ReleaseMbzID: releaseMbzID,
		Time:         time.Now(),
		UserID:       1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created")
}

func TestSubmitListen_MatchNoMbzIDs(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{}
	opts := catalog.SubmitListenOpts{
		MbzCaller:    mbzc,
		ArtistNames:  []string{"ATARASHII GAKKO!"},
		Artist:       "ATARASHII GAKKO!",
		TrackTitle:   "Tokyo Calling",
		ReleaseTitle: "AG! Calling",
		Time:         time.Now(),
		UserID:       1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1 AND musicbrainz_id IS NULL
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created or has been associated with fake musicbrainz id")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1 AND musicbrainz_id IS NULL
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release created or has been associated with fake musicbrainz id")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1 AND musicbrainz_id IS NULL
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created or has been associated with fake musicbrainz id")
}

func TestSubmitListen_UpdateTrackDuration(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{}
	opts := catalog.SubmitListenOpts{
		MbzCaller:    mbzc,
		ArtistNames:  []string{"ATARASHII GAKKO!"},
		Artist:       "ATARASHII GAKKO!",
		TrackTitle:   "Tokyo Calling",
		ReleaseTitle: "AG! Calling",
		Time:         time.Now(),
		Duration:     191,
		UserID:       1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1 AND duration = 191
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "expected duration to be updated")
}

func TestSubmitListen_UpdateTrackDurationWithMbz(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Tracks: mbzTrackData,
	}
	opts := catalog.SubmitListenOpts{
		MbzCaller:      mbzc,
		ArtistNames:    []string{"ATARASHII GAKKO!"},
		Artist:         "ATARASHII GAKKO!",
		TrackTitle:     "Tokyo Calling",
		RecordingMbzID: uuid.MustParse("00000000-0000-0000-0000-000000001001"),
		ReleaseTitle:   "AG! Calling",
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1 AND duration = 191
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "expected duration to be updated")
}

func TestSubmitListen_MatchFromTrackTitleNoMbzIDs(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists: mbzArtistData,
	}
	opts := catalog.SubmitListenOpts{
		MbzCaller: mbzc,
		ArtistMbzIDs: []uuid.UUID{
			uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		},
		Artist:       "ATARASHII GAKKO!",
		TrackTitle:   "Tokyo Calling",
		ReleaseTitle: "AG! Calling",
		Time:         time.Now(),
		UserID:       1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT * FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release created")
}

func TestSubmitListen_AssociateAllMbzIDs(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	// existing artist gets associated with mbz id (also updates aliases)
	// exisiting release gets associated with mbz id
	// existing track gets associated with mbz id (with new artist association)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:  mbzArtistData,
		Releases: mbzReleaseData,
		Tracks:   mbzTrackData,
	}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:     "Tokyo Calling",
		RecordingMbzID: trackMbzID,
		ReleaseTitle:   "AG! Calling",
		ReleaseMbzID:   releaseMbzID,
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created")

	// Verify that the mbz ids were saved
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM tracks
      WHERE musicbrainz_id = $1
    )`, trackMbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected track row with mbz id to exist")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artists
      WHERE musicbrainz_id = $1
    )`, artistMbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist row with mbz id to exist")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM releases
      WHERE musicbrainz_id = $1
    )`, releaseMbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected release row with mbz id to exist")
}

func TestSubmitListen_AssociateAllMbzIDsWithMbzUnreachable(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	// existing artist gets associated with mbz id (also updates aliases)
	// exisiting release gets associated with mbz id
	// existing track gets associated with mbz id (with new artist association)

	ctx := context.Background()
	mbzc := &mbz.MbzErrorCaller{}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:     "Tokyo Calling",
		RecordingMbzID: trackMbzID,
		ReleaseTitle:   "AG! Calling",
		ReleaseMbzID:   releaseMbzID,
		Time:           time.Now(),
		UserID:         1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM tracks_with_title WHERE title = $1
	`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate track created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM releases_with_title WHERE title = $1
	`, "AG! Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate release created")
	count, err = store.Count(ctx, `
	SELECT COUNT(*) FROM artists_with_name WHERE name = $1
	`, "ATARASHII GAKKO!")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "duplicate artist created")

	// Verify that the mbz ids were saved
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM tracks
      WHERE musicbrainz_id = $1
    )`, trackMbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected track row with mbz id to exist")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artists
      WHERE musicbrainz_id = $1
    )`, artistMbzID)
	require.NoError(t, err)
	// as artist names and mbz ids can be ids with unknown order
	assert.False(t, exists, "artists cannot be associated with mbz ids when mbz is unreachable")
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM releases
      WHERE musicbrainz_id = $1
    )`, releaseMbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected release row with mbz id to exist")
}

func TestSubmitListen_AssociateReleaseAliases(t *testing.T) {
	setupTestDataSansMbzIDs(t)

	// existing artist gets associated with mbz id (also updates aliases)
	// exisiting release group gets associated with mbz id
	// existing track gets associated with mbz id (with new artist association)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:       mbzArtistData,
		Releases:      mbzReleaseData,
		Tracks:        mbzTrackData,
		ReleaseGroups: mbzReleaseGroupData,
	}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseGroupMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:        "Tokyo Calling",
		RecordingMbzID:    trackMbzID,
		ReleaseTitle:      "AG! Calling",
		ReleaseMbzID:      releaseMbzID,
		ReleaseGroupMbzID: releaseGroupMbzID,
		Time:              time.Now(),
		UserID:            1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// verify that track, release group, and artist are existing ones and not duplicates
	count, err := store.Count(ctx, `
	SELECT COUNT(*) FROM release_aliases WHERE alias = $1
	`, "AG! Calling - Alt Title")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "expected release alias to exist")
}

func TestSubmitListen_MusicBrainzUnreachable(t *testing.T) {
	truncateTestData(t)

	// test don't fail when mbz unreachable

	ctx := context.Background()
	mbzc := &mbz.MbzErrorCaller{}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	releaseGroupMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!"},
		Artist:      "ATARASHII GAKKO!",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:        "Tokyo Calling",
		RecordingMbzID:    trackMbzID,
		ReleaseTitle:      "AG! Calling",
		ReleaseMbzID:      releaseMbzID,
		ReleaseGroupMbzID: releaseGroupMbzID,
		Time:              time.Now(),
		UserID:            1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")
}

func TestSubmitListen_MusicBrainzUnreachableMBIDMappings(t *testing.T) {
	truncateTestData(t)

	// correctly associate MBID when musicbrainz unreachable, but map provided

	ctx := context.Background()
	mbzc := &mbz.MbzErrorCaller{}
	artistMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	artist2MbzID := uuid.MustParse("00000000-0000-0000-0000-000000000002")
	releaseGroupMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000011")
	releaseMbzID := uuid.MustParse("00000000-0000-0000-0000-000000000101")
	trackMbzID := uuid.MustParse("00000000-0000-0000-0000-000000001001")
	artistMbzIdMap := []catalog.ArtistMbidMap{{Artist: "ATARASHII GAKKO!", Mbid: artistMbzID}, {Artist: "Featured Artist", Mbid: artist2MbzID}}
	opts := catalog.SubmitListenOpts{
		MbzCaller:   mbzc,
		ArtistNames: []string{"ATARASHII GAKKO!", "Featured Artist"},
		Artist:      "ATARASHII GAKKO! feat. Featured Artist",
		ArtistMbzIDs: []uuid.UUID{
			artistMbzID,
		},
		TrackTitle:         "Tokyo Calling",
		RecordingMbzID:     trackMbzID,
		ReleaseTitle:       "AG! Calling",
		ReleaseMbzID:       releaseMbzID,
		ReleaseGroupMbzID:  releaseGroupMbzID,
		ArtistMbidMappings: artistMbzIdMap,
		Time:               time.Now(),
		UserID:             1,
	}

	err := catalog.SubmitListen(ctx, store, opts)
	require.NoError(t, err)

	// Verify that the listen was saved
	exists, err := store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM listens
      WHERE track_id = $1
    )`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen row to exist")

	// Verify that the artist has the mbid saved
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artists
      WHERE musicbrainz_id = $1
    )`, artistMbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist to have correct musicbrainz id")

	// Verify that the artist has the mbid saved
	exists, err = store.RowExists(ctx, `
    SELECT EXISTS (
      SELECT 1 FROM artists
      WHERE musicbrainz_id = $1
    )`, artist2MbzID)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist to have correct musicbrainz id")
}
