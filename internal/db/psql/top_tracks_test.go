package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTopTracksPaginated(t *testing.T) {
	testDataForTopItems(t)
	ctx := context.Background()

	// Test valid
	resp, err := store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 4)
	assert.Equal(t, int64(4), resp.TotalCount)
	assert.Equal(t, "Track One", resp.Items[0].Title)
	assert.Equal(t, "Track Two", resp.Items[1].Title)
	assert.Equal(t, "Track Three", resp.Items[2].Title)
	assert.Equal(t, "Track Four", resp.Items[3].Title)
	// ensure artists are included
	require.Len(t, resp.Items[0].Artists, 1)
	assert.Equal(t, "Artist One", resp.Items[0].Artists[0].Name)

	// Test pagination
	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 2, Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "Track Two", resp.Items[0].Title)

	// Test page out of range
	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 10, Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.False(t, resp.HasNextPage)

	// Test invalid inputs
	_, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Limit: -1, Page: 0})
	assert.Error(t, err)

	_, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: -1})
	assert.Error(t, err)

	// Test specify period
	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodDay}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)
	// should default to PeriodDay
	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)

	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodWeek}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Track Four", resp.Items[0].Title)

	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodMonth}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, int64(2), resp.TotalCount)
	assert.Equal(t, "Track Three", resp.Items[0].Title)
	assert.Equal(t, "Track Four", resp.Items[1].Title)

	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodYear}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, "Track Two", resp.Items[0].Title)
	assert.Equal(t, "Track Three", resp.Items[1].Title)
	assert.Equal(t, "Track Four", resp.Items[2].Title)

	// Test filter by artists and releases
	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, ArtistID: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Track One", resp.Items[0].Title)

	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, AlbumID: 2})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Track Two", resp.Items[0].Title)
	// when both artistID and albumID are specified, artist id is ignored
	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}, AlbumID: 2, ArtistID: 1})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Track Two", resp.Items[0].Title)

	// Test specify dates

	testDataAbsoluteListenTimes(t)

	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Year: 2023}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Track One", resp.Items[0].Title)

	resp, err = store.GetTopTracksPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Month: 6, Year: 2024}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Track Two", resp.Items[0].Title)
}
