package catalog_test

import (
	"context"
	"testing"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/db/sqlite"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTrackWithListens(ctx context.Context, t *testing.T, store *sqlite.Sqlite, listenCount int) (primary, secondary *models.Artist, release *models.Album, track *models.Track) {
	t.Helper()

	mbzc := &mbz.MbzMockCaller{}
	for i := range listenCount {
		opts := catalog.SubmitListenOpts{
			MbzCaller:    mbzc,
			ArtistNames:  []string{"Primary Artist", "Secondary Artist"},
			Artist:       "Primary Artist feat. Secondary Artist",
			TrackTitle:   "Shared Track",
			ReleaseTitle: "Shared Release",
			Time:         time.Now().Add(-time.Duration(i) * time.Hour),
			UserID:       1,
		}
		require.NoError(t, catalog.SubmitListen(ctx, store, opts))
	}

	var err error
	primary, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "Primary Artist"})
	require.NoError(t, err)
	secondary, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "Secondary Artist"})
	require.NoError(t, err)
	release, err = store.GetAlbum(ctx, db.GetAlbumOpts{ArtistID: primary.ID, Title: "Shared Release"})
	require.NoError(t, err)
	track, err = store.GetTrack(ctx, db.GetTrackOpts{Title: "Shared Track", ReleaseID: release.ID, ArtistIDs: []int32{primary.ID}})
	require.NoError(t, err)

	return primary, secondary, release, track
}

func setArtistExcluded(ctx context.Context, t *testing.T, store *sqlite.Sqlite, id int32, excluded bool) {
	t.Helper()
	require.NoError(t, store.UpdateArtist(ctx, db.UpdateArtistOpts{
		ID:                        id,
		ExcludeFromTopListsUpdate: true,
		ExcludeFromTopListsValue:  excluded,
	}))
}

func setAlbumExcluded(ctx context.Context, t *testing.T, store *sqlite.Sqlite, id int32, excluded bool) {
	t.Helper()
	require.NoError(t, store.UpdateAlbum(ctx, db.UpdateAlbumOpts{
		ID:                        id,
		ExcludeFromTopListsUpdate: true,
		ExcludeFromTopListsValue:  excluded,
	}))
}

func allTimeOpts() db.GetItemsOpts {
	return db.GetItemsOpts{Timeframe: db.Timeframe{Period: db.PeriodAllTime}}
}

func TestGetTopArtistsPaginated_ExcludeFromTopLists(t *testing.T) {
	store := newTestDB()
	ctx := context.Background()
	listenCount := 3
	primary, secondary, release, _ := createTrackWithListens(ctx, t, store, listenCount)

	assertArtistListenCount := func(id int32, want int) {
		resp, err := store.GetTopArtistsPaginated(ctx, allTimeOpts())
		require.NoError(t, err)
		for _, it := range resp.Items {
			if it.Item.ID == id {
				assert.EqualValues(t, want, it.Item.ListenCount)
				return
			}
		}
		if want != 0 {
			t.Fatalf("artist %d not found in top artists", id)
		}
	}

	// baseline: both credited artists show all listens (no fan-out from being
	// jointly credited on the same track)
	assertArtistListenCount(primary.ID, listenCount)
	assertArtistListenCount(secondary.ID, listenCount)

	// excluding one artist only hides that artist
	setArtistExcluded(ctx, t, store, secondary.ID, true)
	assertArtistListenCount(primary.ID, listenCount)
	assertArtistListenCount(secondary.ID, 0) // absent from the list
	setArtistExcluded(ctx, t, store, secondary.ID, false)

	// this is always an "everything" view, so it must also respect the
	// release's exclusion flag: excluding the artists' only release hides both
	setAlbumExcluded(ctx, t, store, release.ID, true)
	assertArtistListenCount(primary.ID, 0)
	assertArtistListenCount(secondary.ID, 0)
}

func TestGetTopAlbumsPaginated_ExcludeFromTopLists(t *testing.T) {
	store := newTestDB()
	ctx := context.Background()
	listenCount := 3
	primary, secondary, release, _ := createTrackWithListens(ctx, t, store, listenCount)

	global := allTimeOpts()
	byArtist := allTimeOpts()
	byArtist.ArtistID = int(primary.ID)

	assertAlbumListenCount := func(opts db.GetItemsOpts, want int, msgAndArgs ...any) {
		resp, err := store.GetTopAlbumsPaginated(ctx, opts)
		require.NoError(t, err)
		var found *models.Album
		for _, it := range resp.Items {
			if it.Item.ID == release.ID {
				found = it.Item
			}
		}
		if want > 0 {
			require.NotNil(t, found, msgAndArgs...)
			assert.EqualValues(t, want, found.ListenCount, msgAndArgs...)
		} else {
			assert.Nil(t, found, msgAndArgs...)
		}
	}

	// baseline: visible with the correct count in both scopes (no fan-out from
	// the release's two artist_releases rows)
	assertAlbumListenCount(global, listenCount)
	assertAlbumListenCount(byArtist, listenCount)

	// excluding the non-primary artist never matters: only the release's
	// primary artist is consulted, and scoped views ignore exclusion entirely
	setArtistExcluded(ctx, t, store, secondary.ID, true)
	assertAlbumListenCount(global, listenCount, "secondary artist's flag should not affect the global list")
	setArtistExcluded(ctx, t, store, secondary.ID, false)

	// excluding the primary artist hides it globally but not when scoped to
	// that artist's own albums
	setArtistExcluded(ctx, t, store, primary.ID, true)
	assertAlbumListenCount(global, 0, "expected release hidden globally once its primary artist is excluded")
	assertAlbumListenCount(byArtist, listenCount, "expected release still visible when browsing the excluded artist's own albums")
	setArtistExcluded(ctx, t, store, primary.ID, false)

	// excluding the release itself hides it globally but not when scoped
	setAlbumExcluded(ctx, t, store, release.ID, true)
	assertAlbumListenCount(global, 0)
	assertAlbumListenCount(byArtist, listenCount, "expected release still visible when browsing its own artist's albums")
}

func TestGetTopTracksPaginated_ExcludeFromTopLists(t *testing.T) {
	store := newTestDB()
	ctx := context.Background()
	listenCount := 3
	primary, secondary, release, track := createTrackWithListens(ctx, t, store, listenCount)

	global := allTimeOpts()
	byArtist := allTimeOpts()
	byArtist.ArtistID = int(primary.ID)
	byAlbum := allTimeOpts()
	byAlbum.AlbumID = int(release.ID)

	assertTrackListenCount := func(opts db.GetItemsOpts, want int, store *sqlite.Sqlite, msgAndArgs ...any) {
		resp, err := store.GetTopTracksPaginated(ctx, opts)
		require.NoError(t, err)
		var found *models.Track
		for _, it := range resp.Items {
			if it.Item.ID == track.ID {
				found = it.Item
			}
		}
		if want > 0 {
			require.NotNil(t, found, msgAndArgs...)
			assert.EqualValues(t, want, found.ListenCount, msgAndArgs...)
		} else {
			assert.Nil(t, found, msgAndArgs...)
		}
	}

	// baseline: visible with the correct count in every scope (no fan-out from
	// the track's two artist_tracks rows)
	assertTrackListenCount(global, listenCount, store)
	assertTrackListenCount(byArtist, listenCount, store)
	assertTrackListenCount(byAlbum, listenCount, store)

	// excluding the non-primary artist never matters
	setArtistExcluded(ctx, t, store, secondary.ID, true)
	assertTrackListenCount(global, listenCount, store, "secondary artist's flag should not affect the global list")
	setArtistExcluded(ctx, t, store, secondary.ID, false)

	// excluding the primary artist hides it globally, but scoped views ignore
	// exclusion entirely
	setArtistExcluded(ctx, t, store, primary.ID, true)
	assertTrackListenCount(global, 0, store, "expected track hidden globally once the primary artist is excluded")
	assertTrackListenCount(byArtist, listenCount, store, "expected track still visible when browsing the excluded artist's own tracks")
	assertTrackListenCount(byAlbum, listenCount, store, "expected track still visible when browsing its own album")
	setArtistExcluded(ctx, t, store, primary.ID, false)

	// excluding the release: same story
	setAlbumExcluded(ctx, t, store, release.ID, true)
	assertTrackListenCount(global, 0, store, "expected track hidden globally once its release is excluded")
	assertTrackListenCount(byArtist, listenCount, store, "expected track still visible when browsing its artist's tracks")
	assertTrackListenCount(byAlbum, listenCount, store, "expected track still visible when browsing its own (excluded) album")
}
