package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetTopArtistsPaginated(t *testing.T) {
	testDataForTopItems(t)
	ctx := context.Background()

	// Test valid
	resp, err := store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 4)
	assert.Equal(t, int64(4), resp.TotalCount)
	assert.Equal(t, "Artist One", resp.Items[0].Name)
	assert.Equal(t, "Artist Two", resp.Items[1].Name)
	assert.Equal(t, "Artist Three", resp.Items[2].Name)
	assert.Equal(t, "Artist Four", resp.Items[3].Name)

	// Test pagination
	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 2, Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, "Artist Two", resp.Items[0].Name)

	// Test page out of range
	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: 10, Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	assert.Empty(t, resp.Items)
	assert.False(t, resp.HasNextPage)

	// Test invalid inputs
	_, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Limit: -1, Page: 0})
	assert.Error(t, err)

	_, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Limit: 1, Page: -1})
	assert.Error(t, err)

	// Test specify period
	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodDay}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)
	// should default to PeriodDay
	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{})
	require.NoError(t, err)
	require.Len(t, resp.Items, 0) // empty
	assert.Equal(t, int64(0), resp.TotalCount)

	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodWeek}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Artist Four", resp.Items[0].Name)

	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodMonth}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 2)
	assert.Equal(t, int64(2), resp.TotalCount)
	assert.Equal(t, "Artist Three", resp.Items[0].Name)
	assert.Equal(t, "Artist Four", resp.Items[1].Name)

	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodYear}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 3)
	assert.Equal(t, int64(3), resp.TotalCount)
	assert.Equal(t, "Artist Two", resp.Items[0].Name)
	assert.Equal(t, "Artist Three", resp.Items[1].Name)
	assert.Equal(t, "Artist Four", resp.Items[2].Name)

	// Test specify dates

	testDataAbsoluteListenTimes(t)

	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Year: 2023}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Artist One", resp.Items[0].Name)

	resp, err = store.GetTopArtistsPaginated(ctx, db.GetItemsOpts{Timeframe: db.Timeframe{Month: 6, Year: 2024}})
	require.NoError(t, err)
	require.Len(t, resp.Items, 1)
	assert.Equal(t, int64(1), resp.TotalCount)
	assert.Equal(t, "Artist Two", resp.Items[0].Name)
}
