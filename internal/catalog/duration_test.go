package catalog_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBackfillDuration(t *testing.T) {
	setupTestDataWithMbzIDs(t)

	ctx := context.Background()
	mbzc := &mbz.MbzMockCaller{
		Artists:  mbzArtistData,
		Releases: mbzReleaseData,
		Tracks:   mbzTrackData,
	}

	var err error

	err = catalog.BackfillTrackDurationsFromMusicBrainz(context.Background(), store, &mbz.MbzErrorCaller{})
	assert.NoError(t, err)

	err = catalog.BackfillTrackDurationsFromMusicBrainz(ctx, store, mbzc)
	assert.NoError(t, err)

	count, err := store.Count(ctx, `
		SELECT COUNT(*) FROM tracks_with_title WHERE title = $1 AND duration > 0
		`, "Tokyo Calling")
	require.NoError(t, err)
	assert.Equal(t, 1, count, "track was not updated with duration")
}
