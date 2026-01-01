package psql_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCountListens(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountListens
	period := db.PeriodWeek
	count, err := store.CountListens(ctx, db.Timeframe{Period: period})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "expected listens count to match inserted data")

	truncateTestData(t)
}

func TestCountTracks(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountTracks
	period := db.PeriodMonth
	count, err := store.CountTracks(ctx, db.Timeframe{Period: period})
	require.NoError(t, err)
	assert.Equal(t, int64(2), count, "expected tracks count to match inserted data")

	truncateTestData(t)
}

func TestCountNewTracks(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	testDataAbsoluteListenTimes(t)

	// Test CountTracks
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t1u := t1.Unix()
	t2, _ := time.Parse(time.DateOnly, "2025-12-31")
	t2u := t2.Unix()
	count, err := store.CountNewTracks(ctx, db.Timeframe{FromUnix: t1u, ToUnix: t2u})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "expected tracks count to match inserted data")

	truncateTestData(t)
}

func TestCountAlbums(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountAlbums
	period := db.PeriodYear
	count, err := store.CountAlbums(ctx, db.Timeframe{Period: period})
	require.NoError(t, err)
	assert.Equal(t, int64(3), count, "expected albums count to match inserted data")

	truncateTestData(t)
}

func TestCountNewAlbums(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	testDataAbsoluteListenTimes(t)

	// Test CountTracks
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t1u := t1.Unix()
	t2, _ := time.Parse(time.DateOnly, "2025-12-31")
	t2u := t2.Unix()
	count, err := store.CountNewAlbums(ctx, db.Timeframe{FromUnix: t1u, ToUnix: t2u})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "expected albums count to match inserted data")

	truncateTestData(t)
}

func TestCountArtists(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountArtists
	period := db.PeriodAllTime
	count, err := store.CountArtists(ctx, db.Timeframe{Period: period})
	require.NoError(t, err)
	assert.Equal(t, int64(4), count, "expected artists count to match inserted data")

	truncateTestData(t)
}

func TestCountNewArtists(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	testDataAbsoluteListenTimes(t)

	// Test CountTracks
	t1, _ := time.Parse(time.DateOnly, "2025-01-01")
	t1u := t1.Unix()
	t2, _ := time.Parse(time.DateOnly, "2025-12-31")
	t2u := t2.Unix()
	count, err := store.CountNewArtists(ctx, db.Timeframe{FromUnix: t1u, ToUnix: t2u})
	require.NoError(t, err)
	assert.Equal(t, int64(1), count, "expected artists count to match inserted data")

	truncateTestData(t)
}

func TestCountTimeListened(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)

	// Test CountTimeListened
	period := db.PeriodMonth
	count, err := store.CountTimeListened(ctx, db.Timeframe{Period: period})
	require.NoError(t, err)
	// 3 listens in past month, each 100 seconds
	assert.Equal(t, int64(300), count, "expected total time listened to match inserted data")

	truncateTestData(t)
}

func TestCountTimeListenedToArtist(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	period := db.PeriodAllTime
	count, err := store.CountTimeListenedToItem(ctx, db.TimeListenedOpts{Timeframe: db.Timeframe{Period: period}, ArtistID: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 400, count)
	truncateTestData(t)
}

func TestCountTimeListenedToAlbum(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	period := db.PeriodAllTime
	count, err := store.CountTimeListenedToItem(ctx, db.TimeListenedOpts{Timeframe: db.Timeframe{Period: period}, AlbumID: 2})
	require.NoError(t, err)
	assert.EqualValues(t, 300, count)
	truncateTestData(t)
}

func TestCountTimeListenedToTrack(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	period := db.PeriodAllTime
	count, err := store.CountTimeListenedToItem(ctx, db.TimeListenedOpts{Timeframe: db.Timeframe{Period: period}, TrackID: 3})
	require.NoError(t, err)
	assert.EqualValues(t, 200, count)
	truncateTestData(t)
}

func TestListensToArtist(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	period := db.PeriodAllTime
	count, err := store.CountListensToItem(ctx, db.TimeListenedOpts{Timeframe: db.Timeframe{Period: period}, ArtistID: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 4, count)
	truncateTestData(t)
}

func TestListensToAlbum(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	period := db.PeriodAllTime
	count, err := store.CountListensToItem(ctx, db.TimeListenedOpts{Timeframe: db.Timeframe{Period: period}, AlbumID: 2})
	require.NoError(t, err)
	assert.EqualValues(t, 3, count)
	truncateTestData(t)
}

func TestListensToTrack(t *testing.T) {
	ctx := context.Background()
	testDataForTopItems(t)
	period := db.PeriodAllTime
	count, err := store.CountListensToItem(ctx, db.TimeListenedOpts{Timeframe: db.Timeframe{Period: period}, TrackID: 3})
	require.NoError(t, err)
	assert.EqualValues(t, 2, count)
	truncateTestData(t)
}
