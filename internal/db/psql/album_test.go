package psql_test

import (
	"context"
	"testing"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func truncateTestData(t *testing.T) {
	err := store.Exec(context.Background(),
		`TRUNCATE
		artists,
		artist_aliases,
		tracks,
		artist_tracks,
		releases,
		artist_releases,
		release_aliases,
		listens
		RESTART IDENTITY CASCADE`)
	require.NoError(t, err)
}

func testDataForRelease(t *testing.T) {
	truncateTestData(t)
	err := store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000001')`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary)
			VALUES (1, 'ATARASHII GAKKO!', 'MusicBrainz', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000002')`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary)
			VALUES (2, 'Masayuki Suzuki', 'MusicBrainz', true)`)
	require.NoError(t, err)
}

func TestGetAlbum(t *testing.T) {
	testDataForTopItems(t)
	ctx := context.Background()

	// Test GetAlbum by ID
	result, err := store.GetAlbum(ctx, db.GetAlbumOpts{ID: 1})
	require.NoError(t, err)
	assert.EqualValues(t, 1, result.ID)
	assert.Equal(t, "Release One", result.Title)
	assert.EqualValues(t, 4, result.ListenCount)
	assert.EqualValues(t, 400, result.TimeListened)

	// Test GetAlbum with insufficient information
	_, err = store.GetAlbum(ctx, db.GetAlbumOpts{})
	assert.Error(t, err)

	truncateTestData(t)
}

func TestSaveAlbum(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	// Save release group with artist IDs
	artistIDs := []int32{1, 2}
	rg, err := store.SaveAlbum(ctx, db.SaveAlbumOpts{
		Title:     "New Release Group",
		ArtistIDs: artistIDs,
	})
	require.NoError(t, err)

	// Verify release group was saved
	assert.Equal(t, "New Release Group", rg.Title)

	// Verify release was created for release group
	exists, err := store.RowExists(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM releases_with_title
			WHERE title = $1 AND id = $2
		)`, "New Release Group", rg.ID)
	require.NoError(t, err)
	assert.True(t, exists, "expected release to exist")

	// Verify artist associations were created for release group
	for _, aid := range artistIDs {
		exists, err := store.RowExists(ctx, `
			SELECT EXISTS (
				SELECT 1 FROM artist_releases
				WHERE artist_id = $1 AND release_id = $2
			)`, aid, rg.ID)
		require.NoError(t, err)
		assert.True(t, exists, "expected artist association to exist")
	}

	truncateTestData(t)
}

func TestUpdateAlbum(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	rg, err := store.SaveAlbum(ctx, db.SaveAlbumOpts{
		Title:     "Old Title",
		ArtistIDs: []int32{1},
	})
	require.NoError(t, err)

	newMbzID := uuid.New()
	imgid := uuid.New()
	err = store.UpdateAlbum(ctx, db.UpdateAlbumOpts{
		ID:                   rg.ID,
		MusicBrainzID:        newMbzID,
		Image:                imgid,
		ImageSrc:             catalog.ImageSourceUserUpload,
		VariousArtistsUpdate: true,
		VariousArtistsValue:  true,
	})
	require.NoError(t, err)

	result, err := store.GetAlbum(ctx, db.GetAlbumOpts{ID: rg.ID})
	require.NoError(t, err)
	assert.Equal(t, newMbzID, *result.MbzID)
	assert.Contains(t, result.Image.Small, imgid.String())
	assert.True(t, result.VariousArtists)

	truncateTestData(t)
}
func TestAddArtistsToAlbum(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	// Insert test album
	rg, err := store.SaveAlbum(ctx, db.SaveAlbumOpts{
		Title:     "Test Album",
		ArtistIDs: []int32{1},
	})
	require.NoError(t, err)

	// Add additional artists to the album
	err = store.AddArtistsToAlbum(ctx, db.AddArtistsToAlbumOpts{
		AlbumID:   rg.ID,
		ArtistIDs: []int32{2},
	})
	require.NoError(t, err)

	// Verify artist associations were created
	exists, err := store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM artist_releases
            WHERE artist_id = $1 AND release_id = $2
        )`, 2, rg.ID)
	require.NoError(t, err)
	assert.True(t, exists, "expected artist association to exist")

	truncateTestData(t)
}
func TestSaveAlbumAliases(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	// Insert test album
	rg, err := store.SaveAlbum(ctx, db.SaveAlbumOpts{
		Title:     "Test Album",
		ArtistIDs: []int32{1},
	})
	require.NoError(t, err)

	// Save aliases for the album
	aliases := []string{"Alias 1", "Alias 2"}
	err = store.SaveAlbumAliases(ctx, rg.ID, aliases, "TestSource")
	require.NoError(t, err)

	// Verify aliases were saved
	for _, alias := range aliases {
		exists, err := store.RowExists(ctx, `
            SELECT EXISTS (
                SELECT 1 FROM release_aliases
                WHERE release_id = $1 AND alias = $2
            )`, rg.ID, alias)
		require.NoError(t, err)
		assert.True(t, exists, "expected alias to exist")
	}

	err = store.SetPrimaryAlbumAlias(ctx, 1, "Alias 1")
	require.NoError(t, err)
	album, err := store.GetAlbum(ctx, db.GetAlbumOpts{ID: rg.ID})
	require.NoError(t, err)
	assert.Equal(t, "Alias 1", album.Title)

	err = store.SetPrimaryAlbumAlias(ctx, 1, "Fake Alias")
	require.Error(t, err)

	store.SetPrimaryAlbumAlias(ctx, 1, "Album One")

	truncateTestData(t)
}
func TestDeleteAlbum(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	testDataForTopItems(t)

	// Delete the album
	err := store.DeleteAlbum(ctx, 1)
	require.NoError(t, err)

	// Verify album was deleted
	exists, err := store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM releases
            WHERE id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected album to be deleted")

	// Verify album's track was deleted
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM tracks
            WHERE id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected album's tracks to be deleted")

	// Verify album's listens was deleted
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM listens
            WHERE track_id = $1
        )`, 1)
	require.NoError(t, err)
	assert.False(t, exists, "expected album's listens to be deleted")

	truncateTestData(t)
}
func TestDeleteAlbumAlias(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	// Insert test album
	rg, err := store.SaveAlbum(ctx, db.SaveAlbumOpts{
		Title:     "Test Album",
		ArtistIDs: []int32{1},
	})
	require.NoError(t, err)

	// Save aliases for the album
	aliases := []string{"Alias 1", "Alias 2"}
	err = store.SaveAlbumAliases(ctx, rg.ID, aliases, "TestSource")
	require.NoError(t, err)

	// Delete one alias
	err = store.DeleteAlbumAlias(ctx, rg.ID, "Alias 1")
	require.NoError(t, err)

	// Verify alias was deleted
	exists, err := store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM release_aliases
            WHERE release_id = $1 AND alias = $2
        )`, rg.ID, "Alias 1")
	require.NoError(t, err)
	assert.False(t, exists, "expected alias to be deleted")

	// Verify other alias still exists
	exists, err = store.RowExists(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM release_aliases
            WHERE release_id = $1 AND alias = $2
        )`, rg.ID, "Alias 2")
	require.NoError(t, err)
	assert.True(t, exists, "expected alias to still exist")

	// Ensure primary alias cannot be deleted
	err = store.DeleteAlbumAlias(ctx, rg.ID, "Test Album")
	require.NoError(t, err) // shouldn't error when nothing is deleted
	rg, err = store.GetAlbum(ctx, db.GetAlbumOpts{ID: rg.ID})
	require.NoError(t, err)
	assert.Equal(t, "Test Album", rg.Title)

	truncateTestData(t)
}
func TestGetAllAlbumAliases(t *testing.T) {
	testDataForRelease(t)
	ctx := context.Background()

	// Insert test album
	rg, err := store.SaveAlbum(ctx, db.SaveAlbumOpts{
		Title:     "Test Album",
		ArtistIDs: []int32{1},
	})
	require.NoError(t, err)

	// Save aliases for the album
	aliases := []string{"Alias 1", "Alias 2"}
	err = store.SaveAlbumAliases(ctx, rg.ID, aliases, "TestSource")
	require.NoError(t, err)

	// Retrieve all aliases
	result, err := store.GetAllAlbumAliases(ctx, rg.ID)
	require.NoError(t, err)
	assert.Len(t, result, len(aliases)+1) // new + canonical

	for _, alias := range aliases {
		found := false
		for _, res := range result {
			if res.Alias == alias {
				found = true
				break
			}
		}
		assert.True(t, found, "expected alias to be retrieved")
	}

	truncateTestData(t)
}
