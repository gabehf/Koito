package psql_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testDataForListens(t *testing.T) {
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

	// Insert release aliases
	err = store.Exec(context.Background(),
		`INSERT INTO release_aliases (release_id, alias, source, is_primary)
			VALUES (1, 'Release One', 'Testing', true),
				   (2, 'Release Two', 'Testing', true)`)
	require.NoError(t, err)

	// Insert tracks
	err = store.Exec(context.Background(),
		`INSERT INTO tracks (musicbrainz_id, release_id)
			VALUES ('11111111-1111-1111-1111-111111111111', 1),
				   ('22222222-2222-2222-2222-222222222222', 2)`)
	require.NoError(t, err)

	// Insert track aliases
	err = store.Exec(context.Background(),
		`INSERT INTO track_aliases (track_id, alias, source, is_primary)
			VALUES (1, 'Track One', 'Testing', true),
				   (2, 'Track Two', 'Testing', true)`)
	require.NoError(t, err)

	// Insert artist track associations
	err = store.Exec(context.Background(),
		`INSERT INTO artist_tracks (track_id, artist_id)
			VALUES (1, 1),
				   (2, 2)`)
	require.NoError(t, err)
}

func TestGetListens(t *testing.T) {
	testDataForTopItems(t)
	ctx := context.Background()

	// Test valid
	resp, err := store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 10)
	assert.Equal(t, int64(10), resp.TotalCount)
	require.Len(t, resp.Items[0].Track.Artists, 1)
	require.Len(t, resp.Items[1].Track.Artists, 1)
	// ensure tracks are in the right order (time, desc)
	assert.Equal(t, "Artist Four", resp.Items[0].Track.Artists[0].Name)
	assert.Equal(t, "Artist Three", resp.Items[1].Track.Artists[0].Name)

	// Test pagination
	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 2, Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	require.Len(t, resp.Items[0].Track.Artists, 1)
	assert.Equal(t, true, resp.HasNextPage)
	assert.EqualValues(t, 2, resp.CurrentPage)
	assert.EqualValues(t, 1, resp.ItemsPerPage)
	assert.EqualValues(t, 10, resp.TotalCount)
	assert.Equal(t, "Artist Three", resp.Items[0].Track.Artists[0].Name)

	// Test page out of range
	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Limit: 10, Page: 10, Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.False(t, resp.HasNextPage)

	// Test invalid inputs
	_, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Limit: -1, Page: 0})
	assert.Error(t, err)

	_, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: -1})
	assert.Error(t, err)

	// Test specify period
	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodDay}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)
	// should default to PeriodDay
	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodWeek}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodMonth}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodYear}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 6)
	assert.Equal(t, int64(6), resp.TotalCount)

	// Test filter by artists, releases, and tracks
	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, ArtistID: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 4)
	assert.Equal(t, int64(4), resp.TotalCount)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, AlbumID: 2})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, TrackID: 3})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, int64(2), resp.TotalCount)
	// when both artistID and albumID are specified, artist id is ignored
	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, AlbumID: 2, ArtistID: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)

	// Test specify dates

	testDataAbsoluteListenTimes(t)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Year: 2023}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 4)
	assert.Equal(t, int64(4), resp.TotalCount)

	resp, err = store.GetListensPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Month: 6, Year: 2024}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)

}

func TestSaveListen(t *testing.T) {
	testDataForListens(t)
	ctx := context.Background()

	// Test SaveListen with valid inputs
	err := store.SaveListen(ctx, db.SaveListenOpts{
		TrackID: 1,
		Time:    time.Now(),
		UserID:  1,
	})
	require.NoError(t, err)

	// Verify the listen was saved
	exists, err := store.RowExists(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM listens
			WHERE track_id = $1
		)`, 1)
	require.NoError(t, err)
	assert.True(t, exists, "expected listen to exist")

	// Test SaveListen with missing TrackID
	err = store.SaveListen(ctx, db.SaveListenOpts{
		TrackID: 0,
		Time:    time.Now(),
	})
	assert.Error(t, err)
}

func TestDeleteListen(t *testing.T) {
	testDataForListens(t)
	ctx := context.Background()

	err := store.Exec(ctx, `
		INSERT INTO listens (user_id, track_id, listened_at)
		VALUES (1, 1, to_timestamp(1749464138.0))`)
	require.NoError(t, err)

	err = store.DeleteListen(ctx, 1, time.Unix(1749464138, 0))
	require.NoError(t, err)

	exists, err := store.RowExists(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM listens
			WHERE track_id = $1
		)`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected listen to be deleted")
}
