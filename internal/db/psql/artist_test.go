package psql_test

import (
	"context"
	"slices"
	"testing"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetArtist(t *testing.T) {
	testDataForTopItems(t)
	ctx := context.Background()
	mbzId := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	// Test GetArtist by ID
	result, err := store.GetArtist(ctx, db.GetArtistOpts{ID: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.ID)
	assert.Equal(t, "Artist One", result.Name)
	assert.EqualValues(t, 4, result.ListenCount)
	assert.EqualValues(t, 400, result.TimeListened)

	// Test GetArtist by Name
	result, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "Artist One"})
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.ID)
	assert.Equal(t, "Artist One", result.Name)
	assert.EqualValues(t, 4, result.ListenCount)
	assert.EqualValues(t, 400, result.TimeListened)

	// Test GetArtist by MusicBrainzID
	result, err = store.GetArtist(ctx, db.GetArtistOpts{MusicBrainzID: mbzId})
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.ID)
	assert.Equal(t, "Artist One", result.Name)
	assert.EqualValues(t, 4, result.ListenCount)
	assert.EqualValues(t, 400, result.TimeListened)

	// Test GetArtist with insufficient information
	_, err = store.GetArtist(ctx, db.GetArtistOpts{})
	assert.Error(t, err)

	truncateTestData(t)
}

func TestSaveAliases(t *testing.T) {
	ctx := context.Background()

	// Insert test artist
	artist, err := store.SaveArtist(ctx, db.SaveArtistOpts{
		Name: "Alias Artist",
	})
	require.NoError(t, err)

	// Save aliases
	aliases := []string{"Alias1", "Alias2"}
	err = store.SaveArtistAliases(ctx, artist.ID, aliases, "MusicBrainz")
	require.NoError(t, err)

	// Verify aliases were saved
	for _, alias := range aliases {
		exists, err := store.RowExists(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM artist_aliases
				WHERE artist_id = $1 AND alias = $2
			)`, artist.ID, alias)
		require.NoError(t, err)
		assert.True(t, exists, "expected alias to exist")
	}

	err = store.SetPrimaryArtistAlias(ctx, 1, "Alias1")
	require.NoError(t, err)
	artist, err = store.GetArtist(ctx, db.GetArtistOpts{ID: artist.ID})
	require.NoError(t, err)
	assert.Equal(t, "Alias1", artist.Name)

	err = store.SetPrimaryArtistAlias(ctx, 1, "Fake Alias")
	require.Error(t, err)

	truncateTestData(t)
}

func TestSaveArtist(t *testing.T) {
	ctx := context.Background()

	// Save artist with aliases
	aliases := []string{"Alias1", "Alias2"}
	artist, err := store.SaveArtist(ctx, db.SaveArtistOpts{
		Name:    "New Artist",
		Aliases: aliases,
	})
	require.NoError(t, err)

	// Verify artist was saved
	assert.Equal(t, "New Artist", artist.Name)

	// Verify aliases were saved
	for _, alias := range slices.Concat(aliases, []string{"New Artist"}) {
		exists, err := store.RowExists(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM artist_aliases
				WHERE artist_id = $1 AND alias = $2
			)`, artist.ID, alias)
		require.NoError(t, err)
		assert.True(t, exists, "expected alias '%s' to exist", alias)
	}

	truncateTestData(t)
}

func TestUpdateArtist(t *testing.T) {
	ctx := context.Background()

	// Insert test artist
	artist, err := store.SaveArtist(ctx, db.SaveArtistOpts{
		Name: "Old Name",
	})
	require.NoError(t, err)

	imgid := uuid.New()
	err = store.UpdateArtist(ctx, db.UpdateArtistOpts{
		ID:       artist.ID,
		Image:    imgid,
		ImageSrc: catalog.ImageSourceUserUpload,
	})
	require.NoError(t, err)

	result, err := store.GetArtist(ctx, db.GetArtistOpts{ID: artist.ID})
	require.NoError(t, err)
	assert.Contains(t, result.Image.Small, imgid.String())

	truncateTestData(t)
}
func TestGetAllArtistAliases(t *testing.T) {
	ctx := context.Background()

	// Insert test artist
	artist, err := store.SaveArtist(ctx, db.SaveArtistOpts{
		Name:    "Alias Artist",
		Aliases: []string{"Alias1", "Alias2"},
	})
	require.NoError(t, err)

	// Retrieve all aliases
	result, err := store.GetAllArtistAliases(ctx, artist.ID)
	require.NoError(t, err)
	assert.Len(t, result, 3) // Includes canonical alias

	// Verify aliases were retrieved
	expectedAliases := []string{"Alias Artist", "Alias1", "Alias2"}
	for _, alias := range expectedAliases {
		found := false
		for _, res := range result {
			if res.Alias == alias {
				found = true
				break
			}
		}
		assert.True(t, found, "expected alias '%s' to be retrieved", alias)
	}

	truncateTestData(t)
}
func TestDeleteArtistAlias(t *testing.T) {
	ctx := context.Background()

	// Insert test artist
	artist, err := store.SaveArtist(ctx, db.SaveArtistOpts{
		Name:    "Alias Artist",
		Aliases: []string{"Alias1", "Alias2"},
	})
	require.NoError(t, err)

	// Delete one alias
	err = store.DeleteArtistAlias(ctx, artist.ID, "Alias1")
	require.NoError(t, err)

	// Verify alias was deleted
	exists, err := store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artist_aliases
            WHERE artist_id = $1 AND alias = $2
        )`, artist.ID, "Alias1")
	require.NoError(t, err)
	assert.False(t, exists, "expected alias to be deleted")

	// Verify other alias still exists
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artist_aliases
            WHERE artist_id = $1 AND alias = $2
        )`, artist.ID, "Alias2")
	require.NoError(t, err)
	assert.True(t, exists, "expected alias to still exist")

	// Ensure primary alias cannot be deleted
	err = store.DeleteArtistAlias(ctx, artist.ID, "Alias Artist")
	require.NoError(t, err) // shouldn't error when nothing is deleted
	artist, err = store.GetArtist(ctx, db.GetArtistOpts{ID: 1})
	require.NoError(t, err)
	assert.Equal(t, "Alias Artist", artist.Name)

	truncateTestData(t)
}
func TestDeleteArtist(t *testing.T) {
	ctx := context.Background()

	// set up a lot of test data, 4 artists, 4 albums, 4 tracks, 10 listens
	testDataForTopItems(t)

	// Delete the artist
	err := store.DeleteArtist(ctx, 1)
	require.NoError(t, err)

	// Verify artist was deleted
	exists, err := store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artists
            WHERE id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected artist to be deleted")

	// Verify artist's release was deleted
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM releases
            WHERE id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected artist's release to be deleted")

	// Verify artist's track was deleted
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM tracks
            WHERE id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected artist's tracks to be deleted")

	// Verify artist's listens was deleted
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM listens
            WHERE track_id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected artist's listens to be deleted")

	truncateTestData(t)
}
