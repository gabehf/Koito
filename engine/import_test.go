package engine_test

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"testing"
	"time"

	"github.com/gabehf/koito/engine"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestImportMaloja(t *testing.T) {

	src := path.Join("..", "test_assets", "maloja_import_test.json")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "maloja_import_test.json")

	// not going to make the dest dir because engine should make it already

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	engine.RunImporter(logger.Get(), store, &mbz.MbzErrorCaller{})

	// maloja test import is 38 Magnify Tokyo streams
	a, err := store.GetArtist(context.Background(), db.GetArtistOpts{Name: "Magnify Tokyo"})
	require.NoError(t, err)
	t.Log(a)
	assert.Equal(t, "Magnify Tokyo", a.Name)
	assert.EqualValues(t, 38, a.ListenCount)

	truncateTestData(t)
}

func TestImportSpotify(t *testing.T) {

	src := path.Join("..", "test_assets", "Streaming_History_Audio_spotify_import_test.json")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "Streaming_History_Audio_spotify_import_test.json")

	// not going to make the dest dir because engine should make it already

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	engine.RunImporter(logger.Get(), store, &mbz.MbzErrorCaller{})

	a, err := store.GetArtist(context.Background(), db.GetArtistOpts{Name: "The Story So Far"})
	require.NoError(t, err)
	r, err := store.GetAlbum(context.Background(), db.GetAlbumOpts{ArtistID: a.ID, Title: "The Story So Far / Stick To Your Guns Split"})
	require.NoError(t, err)
	track, err := store.GetTrack(context.Background(), db.GetTrackOpts{Title: "Clairvoyant", ReleaseID: r.ID, ArtistIDs: []int32{a.ID}})
	require.NoError(t, err)
	t.Log(track)
	assert.Equal(t, "Clairvoyant", track.Title)
	// spotify includes duration data, but we only import when reason_end = trackdone
	// this is the only track with valid duration data
	assert.EqualValues(t, 181, track.Duration)

	truncateTestData(t)
}

func TestImportLastFM(t *testing.T) {

	src := path.Join("..", "test_assets", "recenttracks-shoko2-1749776100.json")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "recenttracks-shoko2-1749776100.json")

	// not going to make the dest dir because engine should make it already

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	mbzcMock := &mbz.MbzMockCaller{
		Artists: map[uuid.UUID]*mbz.MusicBrainzArtist{
			uuid.MustParse("4b00640f-3be6-43f8-9b34-ff81bd89320a"): &mbz.MusicBrainzArtist{
				Name: "OurR",
				Aliases: []mbz.MusicBrainzArtistAlias{
					{
						Name:    "OurR",
						Primary: true,
					},
				},
			},
		},
	}

	engine.RunImporter(logger.Get(), store, mbzcMock)

	album, err := store.GetAlbum(context.Background(), db.GetAlbumOpts{MusicBrainzID: uuid.MustParse("e9e78802-0bf8-4ca3-9655-1d943d2d2fa0")})
	require.NoError(t, err)
	assert.Equal(t, "ZOO!!", album.Title)
	artist, err := store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("4b00640f-3be6-43f8-9b34-ff81bd89320a")})
	require.NoError(t, err)
	assert.Equal(t, "OurR", artist.Name)
	artist, err = store.GetArtist(context.Background(), db.GetArtistOpts{Name: "Necry Talkie"})
	require.NoError(t, err)
	track, err := store.GetTrack(context.Background(), db.GetTrackOpts{Title: "放課後の記憶", ReleaseID: album.ID, ArtistIDs: []int32{artist.ID}})
	require.NoError(t, err)
	t.Log(track)
	listens, err := store.GetListensPaginated(context.Background(), db.GetItemsOpts{TrackID: int(track.ID), Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, listens.Items, 1)
	assert.WithinDuration(t, time.Unix(1749774900, 0), listens.Items[0].Time, 1*time.Second)

	truncateTestData(t)
}

func TestImportLastFM_MbzDisabled(t *testing.T) {

	src := path.Join("..", "test_assets", "recenttracks-shoko2-1749776100.json")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "recenttracks-shoko2-1749776100.json")

	// not going to make the dest dir because engine should make it already

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	engine.RunImporter(logger.Get(), store, &mbz.MbzErrorCaller{})

	album, err := store.GetAlbum(context.Background(), db.GetAlbumOpts{MusicBrainzID: uuid.MustParse("e9e78802-0bf8-4ca3-9655-1d943d2d2fa0")})
	require.NoError(t, err)
	assert.Equal(t, "ZOO!!", album.Title)
	artist, err := store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("4b00640f-3be6-43f8-9b34-ff81bd89320a")})
	require.NoError(t, err)
	assert.Equal(t, "OurR", artist.Name)
	artist, err = store.GetArtist(context.Background(), db.GetArtistOpts{Name: "Necry Talkie"})
	require.NoError(t, err)
	track, err := store.GetTrack(context.Background(), db.GetTrackOpts{Title: "放課後の記憶", ReleaseID: album.ID, ArtistIDs: []int32{artist.ID}})
	require.NoError(t, err)
	t.Log(track)
	listens, err := store.GetListensPaginated(context.Background(), db.GetItemsOpts{TrackID: int(track.ID), Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	require.Len(t, listens.Items, 1)
	assert.WithinDuration(t, time.Unix(1749774900, 0), listens.Items[0].Time, 1*time.Second)

	truncateTestData(t)
}

func TestImportListenBrainz(t *testing.T) {

	src := path.Join("..", "test_assets", "listenbrainz_shoko1_1749780844.zip")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "listenbrainz_shoko1_1749780844.zip")

	// not going to make the dest dir because engine should make it already

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	mbzcMock := &mbz.MbzMockCaller{
		Artists: map[uuid.UUID]*mbz.MusicBrainzArtist{
			uuid.MustParse("4b00640f-3be6-43f8-9b34-ff81bd89320a"): {
				Name: "OurR",
				Aliases: []mbz.MusicBrainzArtistAlias{
					{
						Name:    "OurR",
						Primary: true,
					},
				},
			},
			uuid.MustParse("09887aa7-226e-4ecc-9a0c-02d2ae5777e1"): {
				Name: "Carly Rae Jepsen",
				Aliases: []mbz.MusicBrainzArtistAlias{
					{
						Name:    "Carly Rae Jepsen",
						Primary: true,
					},
				},
			},
			uuid.MustParse("78e46ae5-9bfd-433b-be3f-19e993d67ecc"): &mbz.MusicBrainzArtist{
				Name: "Rufus Wainwright",
				Aliases: []mbz.MusicBrainzArtistAlias{
					{
						Name:    "OurR",
						Primary: true,
					},
				},
			},
		},
	}

	engine.RunImporter(logger.Get(), store, mbzcMock)

	album, err := store.GetAlbum(context.Background(), db.GetAlbumOpts{MusicBrainzID: uuid.MustParse("ce330d67-9c46-4a3b-9d62-08406370f234")})
	require.NoError(t, err)
	assert.Equal(t, "酸欠少女", album.Title)
	artist, err := store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("4b00640f-3be6-43f8-9b34-ff81bd89320a")})
	require.NoError(t, err)
	assert.Equal(t, "OurR", artist.Name)
	artist, err = store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("09887aa7-226e-4ecc-9a0c-02d2ae5777e1")})
	require.NoError(t, err)
	assert.Equal(t, "Carly Rae Jepsen", artist.Name)
	artist, err = store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("78e46ae5-9bfd-433b-be3f-19e993d67ecc")})
	require.NoError(t, err)
	assert.Equal(t, "Rufus Wainwright", artist.Name)
	track, err := store.GetTrack(context.Background(), db.GetTrackOpts{MusicBrainzID: uuid.MustParse("08e8f55b-f1a4-46b8-b2d1-fab4c592165c")})
	require.NoError(t, err)
	assert.Equal(t, "Desert", track.Title)
	listens, err := store.GetListensPaginated(context.Background(), db.GetItemsOpts{TrackID: int(track.ID), Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	assert.Len(t, listens.Items, 1)
	assert.WithinDuration(t, time.Unix(1749780612, 0), listens.Items[0].Time, 1*time.Second)

	truncateTestData(t)
}

func TestImportListenBrainz_MbzDisabled(t *testing.T) {

	src := path.Join("..", "test_assets", "listenbrainz_shoko1_1749780844.zip")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "listenbrainz_shoko1_1749780844.zip")

	// not going to make the dest dir because engine should make it already

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	engine.RunImporter(logger.Get(), store, &mbz.MbzErrorCaller{})

	album, err := store.GetAlbum(context.Background(), db.GetAlbumOpts{MusicBrainzID: uuid.MustParse("ce330d67-9c46-4a3b-9d62-08406370f234")})
	require.NoError(t, err)
	assert.Equal(t, "酸欠少女", album.Title)
	artist, err := store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("4b00640f-3be6-43f8-9b34-ff81bd89320a")})
	require.NoError(t, err)
	assert.Equal(t, "OurR", artist.Name)
	artist, err = store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("09887aa7-226e-4ecc-9a0c-02d2ae5777e1")})
	require.NoError(t, err)
	assert.Equal(t, "Carly Rae Jepsen", artist.Name)
	artist, err = store.GetArtist(context.Background(), db.GetArtistOpts{MusicBrainzID: uuid.MustParse("78e46ae5-9bfd-433b-be3f-19e993d67ecc")})
	require.NoError(t, err)
	assert.Equal(t, "Rufus Wainwright", artist.Name)
	track, err := store.GetTrack(context.Background(), db.GetTrackOpts{MusicBrainzID: uuid.MustParse("08e8f55b-f1a4-46b8-b2d1-fab4c592165c")})
	require.NoError(t, err)
	assert.Equal(t, "Desert", track.Title)
	listens, err := store.GetListensPaginated(context.Background(), db.GetItemsOpts{TrackID: int(track.ID), Timeframe: db.Timeframe{Period: db.PeriodAllTime}})
	require.NoError(t, err)
	assert.Len(t, listens.Items, 1)
	assert.WithinDuration(t, time.Unix(1749780612, 0), listens.Items[0].Time, 1*time.Second)

	truncateTestData(t)
}

func TestImportKoito(t *testing.T) {

	src := path.Join("..", "test_assets", "koito_export_test.json")
	destDir := filepath.Join(cfg.ConfigDir(), "import")
	dest := filepath.Join(destDir, "koito_export_test.json")

	ctx := context.Background()

	// 4 every wave to ever rise, 3 i can't feel you, 5 giri giri, 1 nijinoiroyo
	giriReleaseMBID := uuid.MustParse("ac1f8da0-21d7-426e-83b0-befff06f0871")
	suzukiMBID := uuid.MustParse("30f851bb-dba3-4e9b-811c-5f27f595c86a")
	nijinoTrackMBID := uuid.MustParse("a4f26836-3894-46c1-acac-227808308687")
	lp3MBID := uuid.MustParse("d0ec30bd-7cdc-417c-979d-5a0631b8a161")

	input, err := os.ReadFile(src)
	require.NoError(t, err)

	require.NoError(t, os.WriteFile(dest, input, os.ModePerm))

	engine.RunImporter(logger.Get(), store, &mbz.MbzErrorCaller{})

	// ensure all artists are saved
	_, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "American Football"})
	assert.NoError(t, err)
	_, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "Rachel Goswell"})
	assert.NoError(t, err)
	_, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "Elizabeth Powell"})
	assert.NoError(t, err)

	// ensure artist aliases are saved
	artist, err := store.GetArtist(ctx, db.GetArtistOpts{MusicBrainzID: suzukiMBID})
	require.NoError(t, err)
	assert.Equal(t, "鈴木雅之", artist.Name)
	assert.Contains(t, artist.Aliases, "Masayuki Suzuki")
	_, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "すぅ"})
	require.NoError(t, err)

	// ensure albums are saved
	album, err := store.GetAlbum(ctx, db.GetAlbumOpts{MusicBrainzID: giriReleaseMBID})
	require.NoError(t, err)
	assert.Equal(t, "GIRI GIRI", album.Title)
	// ensure album aliases are saved
	artist, err = store.GetArtist(ctx, db.GetArtistOpts{Name: "NELKE"})
	require.NoError(t, err)
	album, err = store.GetAlbum(ctx, db.GetAlbumOpts{Title: "虹の色よ鮮やかであれ (NELKE ver.)", ArtistID: artist.ID})
	require.NoError(t, err)
	aliases, err := store.GetAllAlbumAliases(ctx, album.ID)
	require.NoError(t, err)
	assert.Contains(t, utils.FlattenAliases(aliases), "Nijinoiroyo Azayakadeare (NELKE ver.)")
	// ensure album associations are saved
	album, err = store.GetAlbum(ctx, db.GetAlbumOpts{MusicBrainzID: lp3MBID})
	require.NoError(t, err)
	assert.Contains(t, utils.FlattenSimpleArtistNames(album.Artists), "Elizabeth Powell")
	assert.Contains(t, utils.FlattenSimpleArtistNames(album.Artists), "Rachel Goswell")
	assert.Contains(t, utils.FlattenSimpleArtistNames(album.Artists), "American Football")

	// ensure all tracks are saved
	track, err := store.GetTrack(ctx, db.GetTrackOpts{MusicBrainzID: nijinoTrackMBID})
	require.NoError(t, err)
	assert.Equal(t, "虹の色よ鮮やかであれ (NELKE ver.)", track.Title)
	aliases, err = store.GetAllTrackAliases(ctx, track.ID)
	require.NoError(t, err)
	assert.Contains(t, utils.FlattenAliases(aliases), "Nijinoiroyo Azayakadeare (NELKE ver.)")
	// ensure track duration is saved
	assert.EqualValues(t, 218, track.Duration)

	artist, err = store.GetArtist(ctx, db.GetArtistOpts{MusicBrainzID: suzukiMBID})
	require.NoError(t, err)
	album, err = store.GetAlbum(ctx, db.GetAlbumOpts{ArtistID: artist.ID, Title: "GIRI GIRI"})
	require.NoError(t, err)
	_, err = store.GetTrack(ctx, db.GetTrackOpts{Title: "GIRI GIRI", ReleaseID: album.ID, ArtistIDs: []int32{artist.ID}})
	require.NoError(t, err)

	count, err := store.CountTracks(ctx, db.Timeframe{Period: db.PeriodAllTime})
	require.NoError(t, err)
	assert.EqualValues(t, 4, count)
	count, err = store.CountAlbums(ctx, db.Timeframe{Period: db.PeriodAllTime})
	require.NoError(t, err)
	assert.EqualValues(t, 3, count)
	count, err = store.CountArtists(ctx, db.Timeframe{Period: db.PeriodAllTime})
	require.NoError(t, err)
	assert.EqualValues(t, 6, count)

	truncateTestData(t)
}
