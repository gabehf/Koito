package catalog_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db/psql"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
	_ "github.com/gabehf/koito/testing_init"
	"github.com/google/uuid"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	mbzArtistData = map[uuid.UUID]*mbz.MusicBrainzArtist{
		uuid.MustParse("00000000-0000-0000-0000-000000000001"): {
			Name:     "ATARASHII GAKKO!",
			SortName: "Atarashii Gakko",
			Aliases: []mbz.MusicBrainzArtistAlias{
				{
					Name:    "新しい学校のリーダーズ",
					Type:    "Artist name",
					Primary: true,
				},
			},
		},
	}
	mbzReleaseGroupData = map[uuid.UUID]*mbz.MusicBrainzReleaseGroup{
		uuid.MustParse("00000000-0000-0000-0000-000000000011"): {
			Title: "AG! Calling",
			Type:  "Album",
			ArtistCredit: []mbz.MusicBrainzArtistCredit{
				{
					Artist: mbz.MusicBrainzArtist{
						Name: "ATARASHII GAKKO!",
						Aliases: []mbz.MusicBrainzArtistAlias{
							{
								Name:    "新しい学校のリーダーズ",
								Type:    "Artist name",
								Primary: true,
							},
						},
					},
					Name: "ATARASHII GAKKO!",
				},
			},
			Releases: []mbz.MusicBrainzRelease{
				{
					Title: "AG! Calling",
					ID:    "00000000-0000-0000-0000-000000000101",
					ArtistCredit: []mbz.MusicBrainzArtistCredit{
						{
							Artist: mbz.MusicBrainzArtist{
								Name: "ATARASHII GAKKO!",
								Aliases: []mbz.MusicBrainzArtistAlias{
									{
										Name:    "ATARASHII GAKKO!",
										Type:    "Artist name",
										Primary: true,
									},
								},
							},
							Name: "ATARASHII GAKKO!",
						},
					},
					Status: "Official",
				},
				{
					Title: "AG! Calling - Alt Title",
					ID:    "00000000-0000-0000-0000-000000000102",
					ArtistCredit: []mbz.MusicBrainzArtistCredit{
						{
							Artist: mbz.MusicBrainzArtist{
								Name: "ATARASHII GAKKO!",
								Aliases: []mbz.MusicBrainzArtistAlias{
									{
										Name:    "ATARASHII GAKKO!",
										Type:    "Artist name",
										Primary: true,
									},
								},
							},
							Name: "ATARASHII GAKKO!",
						},
					},
					Status: "Official",
				},
			},
		},
	}
	mbzReleaseData = map[uuid.UUID]*mbz.MusicBrainzRelease{
		uuid.MustParse("00000000-0000-0000-0000-000000000101"): {
			Title: "AG! Calling",
			ID:    "00000000-0000-0000-0000-000000000101",
			ArtistCredit: []mbz.MusicBrainzArtistCredit{
				{
					Artist: mbz.MusicBrainzArtist{
						Name: "ATARASHII GAKKO!",
						Aliases: []mbz.MusicBrainzArtistAlias{
							{
								Name:    "新しい学校のリーダーズ",
								Type:    "Artist name",
								Primary: true,
							},
						},
					},
					Name: "ATARASHII GAKKO!",
				},
			},
			Status: "Official",
		},
		uuid.MustParse("00000000-0000-0000-0000-000000000202"): {
			Title: "EVANGELION FINALLY",
			ID:    "00000000-0000-0000-0000-000000000202",
			ArtistCredit: []mbz.MusicBrainzArtistCredit{
				{
					Artist: mbz.MusicBrainzArtist{
						Name: "Various Artists",
					},
					Name: "Various Artists",
				},
			},
			Status: "Official",
		},
	}
	mbzTrackData = map[uuid.UUID]*mbz.MusicBrainzTrack{
		uuid.MustParse("00000000-0000-0000-0000-000000001001"): {
			Title:    "Tokyo Calling",
			LengthMs: 191000,
		},
	}
)

var store *psql.Psql

func getTestGetenv(resource *dockertest.Resource) func(string) string {
	dir, err := utils.GenerateRandomString(8)
	if err != nil {
		panic(err)
	}
	return func(env string) string {
		switch env {
		case cfg.ENABLE_STRUCTURED_LOGGING_ENV:
			return "true"
		case cfg.LOG_LEVEL_ENV:
			return "debug"
		case cfg.DATABASE_URL_ENV:
			return fmt.Sprintf("postgres://postgres:secret@localhost:%s", resource.GetPort("5432/tcp"))
		case cfg.CONFIG_DIR_ENV:
			return dir
		case cfg.DISABLE_DEEZER_ENV, cfg.DISABLE_COVER_ART_ARCHIVE_ENV, cfg.DISABLE_MUSICBRAINZ_ENV, cfg.ENABLE_FULL_IMAGE_CACHE_ENV:
			return "true"
		default:
			return ""
		}
	}
}

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

func setupTestDataWithMbzIDs(t *testing.T) {
	truncateTestData(t)

	err := store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000001')`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary)
			VALUES (1, 'ATARASHII GAKKO!', 'Testing', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO releases (musicbrainz_id)
			VALUES ('00000000-0000-0000-0000-000000000101')`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO release_aliases (release_id, alias, source, is_primary)
			VALUES (1, 'AG! Calling', 'Testing', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_releases (artist_id, release_id)
			VALUES (1, 1)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO tracks (release_id, musicbrainz_id)
			VALUES (1, '00000000-0000-0000-0000-000000001001')`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO track_aliases (track_id, alias, source, is_primary)
			VALUES (1, 'Tokyo Calling', 'Testing', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_tracks (artist_id, track_id)
			VALUES (1, 1)`)
	require.NoError(t, err)
}

func setupTestDataSansMbzIDs(t *testing.T) {
	truncateTestData(t)

	err := store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id)
			VALUES (NULL)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary)
			VALUES (1, 'ATARASHII GAKKO!', 'Testing', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO releases (musicbrainz_id)
			VALUES (NULL)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO release_aliases (release_id, alias, source, is_primary)
			VALUES (1, 'AG! Calling', 'Testing', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_releases (artist_id, release_id)
			VALUES (1, 1)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO tracks (release_id, musicbrainz_id)
			VALUES (1, NULL)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO track_aliases (track_id, alias, source, is_primary)
			VALUES (1, 'Tokyo Calling', 'Testing', true)`)
	require.NoError(t, err)
	err = store.Exec(context.Background(),
		`INSERT INTO artist_tracks (artist_id, track_id)
			VALUES (1, 1)`)
	require.NoError(t, err)
}

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	if err := pool.Client.Ping(); err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	err = cfg.Load(getTestGetenv(resource), "test")
	if err != nil {
		log.Fatalf("Could not load cfg: %s", err)
	}

	if err := pool.Retry(func() error {
		var err error
		store, err = psql.New()
		if err != nil {
			log.Println("Failed to connect to test database, retrying...")
			return err
		}
		return store.Ping(context.Background())
	}); err != nil {
		log.Fatalf("Could not connect to database: %s", err)
	}

	// insert a user into the db with id 1 to use for tests
	err = store.Exec(context.Background(), `INSERT INTO users (username, password) VALUES ('test', DECODE('abc123', 'hex'))`)
	if err != nil {
		log.Fatalf("Failed to insert test user: %v", err)
	}

	code := m.Run()

	// You can't defer this because os.Exit doesn't care for defer
	if err := pool.Purge(resource); err != nil {
		log.Fatalf("Could not purge resource: %s", err)
	}

	err = os.RemoveAll(cfg.ConfigDir())
	if err != nil {
		log.Fatalf("Could not remove temporary config dir: %v", err)
	}

	os.Exit(code)
}

// From: https://brandur.org/fragments/go-equal-time
// EqualTime compares two times in a way that's safer and with better fail
// output than a call to `require.Equal` would produce.
//
// It takes care to:
//
//   - Strip off monotonic portions of timestamps so they aren't considered for
//     purposes of comparison.
//
//   - Truncate nanoseconds in a functionally equivalent way to how pgx would do
//     it so that times that have round-tripped from Postgres can still be
//     compared. Postgres only stores times to the microsecond level.
//
//   - Use formatted, human-friendly time outputs so that in case of a failure,
//     the discrepancy is easier to pick out.
func EqualTime(t testing.TB, t1, t2 time.Time) {
	// Note that leaving off the nanosecond portion will have the effect of
	// truncating it rather than rounding to the nearest microsecond, which
	// functionally matches pgx's behavior while persisting.
	const rfc3339Micro = "2006-01-02T15:04:05.999999Z07:00"

	require.Equal(t,
		t1.Format(rfc3339Micro),
		t2.Format(rfc3339Micro),
	)
}

func TestArtistStringParse(t *testing.T) {
	type input struct {
		Name  string
		Title string
	}
	cases := map[input][]string{
		// only one artist
		{"NELKE", ""}:                 {"NELKE"},
		{"The Brook & The Bluff", ""}: {"The Brook & The Bluff"},
		{"half·alive", ""}:            {"half·alive"},
		// Earth, Wind, & Fire
		{"Earth, Wind & Fire", "The Very Best of Earth, Wind & Fire"}: {"Earth, Wind & Fire"},
		// only artists in artist string
		{"Carly Rae Jepsen feat. Rufus Wainwright", ""}: {"Carly Rae Jepsen", "Rufus Wainwright"},
		{"Mimi (feat. HATSUNE MIKU & KAFU)", ""}:        {"Mimi", "HATSUNE MIKU", "KAFU"},
		{"Magnify Tokyo · Kanade Ishihara", ""}:         {"Magnify Tokyo", "Kanade Ishihara"},
		{"Daft Punk [feat. Paul Williams]", ""}:         {"Daft Punk", "Paul Williams"},
		// primary artist in artist string, features in title
		{"Tyler, The Creator", "CA (feat. Alice Smith, Leon Ware & Clem Creevy)"}: {"Tyler, The Creator", "Alice Smith", "Leon Ware", "Clem Creevy"},
		{"ONE OK ROCK", "C.U.R.I.O.S.I.T.Y. (feat. Paledusk and CHICO CARLITO)"}:  {"ONE OK ROCK", "Paledusk", "CHICO CARLITO"},
		{"Rat Tally", "In My Car feat. Madeline Kenney"}:                          {"Rat Tally", "Madeline Kenney"},
		// artists in both
		{"Daft Punk feat. Julian Casablancas", "Instant Crush (feat. Julian Casablancas)"}:   {"Daft Punk", "Julian Casablancas"},
		{"Paramore (feat. Joy Williams)", "Hate to See Your Heart Break feat. Joy Williams"}: {"Paramore", "Joy Williams"},
		{"MINSU", "오해 금지 (Feat. BIG Naughty)"}:                                               {"MINSU", "BIG Naughty"},
		{"MINSU", "오해 금지 [Feat. BIG Naughty]"}:                                               {"MINSU", "BIG Naughty"},
		{"MINSU", "오해 금지 Feat. BIG Naughty"}:                                                 {"MINSU", "BIG Naughty"},

		// custom separator
		{"MIMiNARI//楠木ともり", "眠れない"}: {"MIMiNARI", "楠木ともり"},
	}

	for in, out := range cases {
		artists := catalog.ParseArtists(in.Name, in.Title, []*regexp.Regexp{regexp.MustCompile(`\s*//\s*`), regexp.MustCompile(`\s+·\s+`)})
		assert.ElementsMatch(t, out, artists)
	}
}
