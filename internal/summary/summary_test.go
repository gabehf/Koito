package summary_test

import (
	"testing"

	"github.com/gabehf/koito/internal/cfg"
)

func TestMain(t *testing.M) {
	// dir, err := utils.GenerateRandomString(8)
	// if err != nil {
	// 	panic(err)
	// }
	cfg.Load(func(env string) string {
		switch env {
		case cfg.ENABLE_STRUCTURED_LOGGING_ENV:
			return "true"
		case cfg.LOG_LEVEL_ENV:
			return "debug"
		case cfg.DATABASE_URL_ENV:
			return "postgres://postgres:secret@localhost"
		case cfg.CONFIG_DIR_ENV:
			return "."
		case cfg.DISABLE_DEEZER_ENV, cfg.DISABLE_COVER_ART_ARCHIVE_ENV, cfg.DISABLE_MUSICBRAINZ_ENV, cfg.ENABLE_FULL_IMAGE_CACHE_ENV:
			return "true"
		default:
			return ""
		}
	}, "test")
	t.Run()
}

func TestGenerateSummary(t *testing.T) {
	// s := summary.Summary{
	// 	Title: "20XX Rewind",
	// TopArtistImage: path.Join("..", "..", "test_assets", "yuu.jpg"),
	// TopArtists: []struct {
	// 	Name            string
	// 	Plays           int
	// 	MinutesListened int
	// }{
	// 	{"CHUU", 738, 7321},
	// 	{"Paramore", 738, 7321},
	// 	{"ano", 738, 7321},
	// 	{"NELKE", 738, 7321},
	// 	{"ILLIT", 738, 7321},
	// },
	// TopAlbumImage: "",
	// TopAlbums: []struct {
	// 	Title           string
	// 	Plays           int
	// 	MinutesListened int
	// }{
	// 	{"Only cry in the rain", 738, 7321},
	// 	{"Paramore", 738, 7321},
	// 	{"ano", 738, 7321},
	// 	{"NELKE", 738, 7321},
	// 	{"ILLIT", 738, 7321},
	// },
	// TopTrackImage: "",
	// TopTracks: []struct {
	// 	Title           string
	// 	Plays           int
	// 	MinutesListened int
	// }{
	// 	{"虹の色よ鮮やかであれ (NELKE ver.)", 321, 12351},
	// 	{"Paramore", 738, 7321},
	// 	{"ano", 738, 7321},
	// 	{"NELKE", 738, 7321},
	// 	{"ILLIT", 738, 7321},
	// },
	// 	MinutesListened: 0,
	// 	Plays:           0,
	// 	AvgPlaysPerDay:  0,
	// 	UniqueTracks:    0,
	// 	UniqueAlbums:    0,
	// 	UniqueArtists:   0,
	// 	NewTracks:       0,
	// 	NewAlbums:       0,
	// 	NewArtists:      0,
	// }

	// assert.NoError(t, summary.GenerateImage(&s))
}
