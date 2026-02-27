package psql_test

import (
	"context"
	"fmt"
	"log"
	"os"
	"testing"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db/psql"
	_ "github.com/gabehf/koito/testing_init"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/require"
)

var store *psql.Psql

func getTestGetenv(resource *dockertest.Resource) func(string) string {
	return func(env string) string {
		switch env {
		case cfg.DATABASE_URL_ENV:
			return fmt.Sprintf("postgres://postgres:secret@localhost:%s", resource.GetPort("5432/tcp"))
		default:
			return ""
		}
	}
}

func TestMain(m *testing.M) {
	// Allow skipping Docker-based integration setup for local unit test runs.
	if os.Getenv("KOITO_SKIP_DOCKER") == "1" {
		log.Println("KOITO_SKIP_DOCKER=1 set; skipping Docker-based tests in psql_test")
		os.Exit(m.Run())
	}
	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not construct pool: %s", err)
	}

	// uses pool to try to connect to Docker
	err = pool.Client.Ping()
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.Run("postgres", "latest", []string{"POSTGRES_PASSWORD=secret"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	err = cfg.Load(getTestGetenv(resource), "test")
	if err != nil {
		log.Fatalf("Could not load cfg: %s", err)
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
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

	// as of go1.15 testing.M returns the exit code of m.Run(), so it is safe to use defer here
	defer func() {
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}()

	// insert a user into the db with id 1 to use for tests
	err = store.Exec(context.Background(), `INSERT INTO users (username, password) VALUES ('test', DECODE('abc123', 'hex'))`)
	if err != nil {
		log.Fatalf("Failed to insert test user: %v", err)
	}

	m.Run()
}

func testDataForTopItems(t *testing.T) {
	truncateTestData(t)

	// artist 1 has most listens older than 1 year
	// artist 2 has most listens older than 1 month
	// artist 3 has most listens older than 1 week
	// artist 4 has least listens

	err := store.Exec(context.Background(),
		`INSERT INTO artists (musicbrainz_id) 
			VALUES ('00000000-0000-0000-0000-000000000001'),
				   ('00000000-0000-0000-0000-000000000002'),
				   ('00000000-0000-0000-0000-000000000003'),
				   ('00000000-0000-0000-0000-000000000004')`)
	require.NoError(t, err)

	err = store.Exec(context.Background(),
		`INSERT INTO artist_aliases (artist_id, alias, source, is_primary) 
			VALUES (1, 'Artist One', 'Testing', true),
				   (2, 'Artist Two', 'Testing', true),
				   (3, 'Artist Three', 'Testing', true),
				   (4, 'Artist Four', 'Testing', true)`)
	require.NoError(t, err)

	// Insert release groups
	err = store.Exec(context.Background(),
		`INSERT INTO releases (musicbrainz_id) 
			VALUES ('00000000-0000-0000-0000-000000000011'),
				   ('00000000-0000-0000-0000-000000000022'),
				   ('00000000-0000-0000-0000-000000000033'),
				   ('00000000-0000-0000-0000-000000000044')`)
	require.NoError(t, err)

	err = store.Exec(context.Background(),
		`INSERT INTO release_aliases (release_id, alias, source, is_primary) 
			VALUES (1, 'Release One', 'Testing', true),
				   (2, 'Release Two', 'Testing', true),
				   (3, 'Release Three', 'Testing', true),
				   (4, 'Release Four', 'Testing', true)`)
	require.NoError(t, err)

	// Insert release groups
	err = store.Exec(context.Background(),
		`INSERT INTO artist_releases (release_id, artist_id) 
			VALUES (1, 1), (2, 2), (3, 3), (4, 4)`)
	require.NoError(t, err)

	// Insert tracks
	err = store.Exec(context.Background(),
		`INSERT INTO tracks (musicbrainz_id, release_id, duration) 
			VALUES ('11111111-1111-1111-1111-111111111111', 1, 100),
				   ('22222222-2222-2222-2222-222222222222', 2, 100),
				   ('33333333-3333-3333-3333-333333333333', 3, 100),
				   ('44444444-4444-4444-4444-444444444444', 4, 100)`)
	require.NoError(t, err)

	err = store.Exec(context.Background(),
		`INSERT INTO track_aliases (track_id, alias, source, is_primary) 
			VALUES (1, 'Track One', 'Testing', true),
				   (2, 'Track Two', 'Testing', true),
				   (3, 'Track Three', 'Testing', true),
				   (4, 'Track Four', 'Testing', true)`)
	require.NoError(t, err)

	// Associate tracks with artists
	err = store.Exec(context.Background(),
		`INSERT INTO artist_tracks (artist_id, track_id) 
			VALUES (1, 1), (2, 2), (3, 3), (4, 4)`)
	require.NoError(t, err)

	// Insert listens
	err = store.Exec(context.Background(),
		`INSERT INTO listens (user_id, track_id, listened_at) 
			VALUES (1, 1, NOW() - INTERVAL '2 years 1 day'),
				   (1, 1, NOW() - INTERVAL '2 years 2 days'),
				   (1, 1, NOW() - INTERVAL '2 years 3 days'),
				   (1, 1, NOW() - INTERVAL '2 years 4 days'),
				   (1, 2, NOW() - INTERVAL '2 months 1 day'),
				   (1, 2, NOW() - INTERVAL '2 months 2 days'),
				   (1, 2, NOW() - INTERVAL '2 months 3 days'),
				   (1, 3, NOW() - INTERVAL '2 weeks'),
				   (1, 3, NOW() - INTERVAL '2 weeks 1 day'),
				   (1, 4, NOW() - INTERVAL '2 days')`)
	require.NoError(t, err)
}

func testDataAbsoluteListenTimes(t *testing.T) {
	err := store.Exec(context.Background(),
		`TRUNCATE listens`)
	require.NoError(t, err)

	err = store.Exec(context.Background(),
		`INSERT INTO listens (user_id, track_id, listened_at) 
			VALUES (1, 1, '2023-06-22 19:11:25-07'),
				   (1, 1, '2023-06-22 19:12:25-07'),
				   (1, 1, '2023-06-22 19:13:25-07'),
				   (1, 1, '2023-06-22 19:14:25-07'),
				   (1, 2, '2024-06-22 19:15:25-07'),
				   (1, 2, '2024-06-22 19:16:25-07'),
				   (1, 2, '2024-06-22 19:17:25-07'),
				   (1, 3, '2024-10-02 19:18:25-07'),
				   (1, 3, '2024-10-02 19:19:25-07'),
				   (1, 4, '2025-05-16 19:20:25-07')`)
	require.NoError(t, err)
}
