package migrate

import (
	"context"
	"database/sql"
	"fmt"
	"path"
	"time"

	migrations_sqlite "github.com/gabehf/koito/db/migrations_sqlite"
	"github.com/gabehf/koito/internal/cfg"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
	_ "modernc.org/sqlite"
)

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Migrate streams all data from the Postgres database at cfg.DatabaseUrl() into
// the SQLite database at cfg.ConfigDir()/koito.db. It runs the SQLite schema
// migrations first, then aborts if the target already contains data.
func Migrate(ctx context.Context, l *zerolog.Logger) error {
	pgDB, err := sql.Open("pgx", cfg.DatabaseUrl())
	if err != nil {
		return fmt.Errorf("migrate: open postgres: %w", err)
	}
	defer pgDB.Close()
	if err := pgDB.PingContext(ctx); err != nil {
		return fmt.Errorf("migrate: ping postgres: %w", err)
	}

	sqliteDSN := path.Join(cfg.ConfigDir(), "koito.db") +
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	sqliteDB, err := sql.Open("sqlite", sqliteDSN)
	if err != nil {
		return fmt.Errorf("migrate: open sqlite: %w", err)
	}
	defer sqliteDB.Close()

	goose.SetBaseFS(migrations_sqlite.Files)
	goose.SetDialect("sqlite3")
	if err := goose.Up(sqliteDB, "."); err != nil {
		return fmt.Errorf("migrate: sqlite schema migration: %w", err)
	}

	var existing int
	if err := sqliteDB.QueryRowContext(ctx, "SELECT COUNT(*) FROM users").Scan(&existing); err != nil {
		return fmt.Errorf("migrate: check existing data: %w", err)
	}
	if existing > 0 {
		return fmt.Errorf("migrate: target SQLite database already contains data; aborting to prevent data loss")
	}

	if _, err := sqliteDB.ExecContext(ctx, "PRAGMA foreign_keys = OFF"); err != nil {
		return fmt.Errorf("migrate: disable foreign keys: %w", err)
	}

	tx, err := sqliteDB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("migrate: begin transaction: %w", err)
	}
	defer func() {
		if tx != nil {
			tx.Rollback()
		}
	}()

	if err := migrateUsers(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateApiKeys(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateSessions(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateArtists(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateArtistAliases(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateReleases(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateReleaseAliases(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateArtistReleases(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateTracks(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateTrackAliases(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateArtistTracks(ctx, pgDB, tx, l); err != nil {
		return err
	}
	if err := migrateListens(ctx, pgDB, tx, l); err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("migrate: commit: %w", err)
	}
	tx = nil

	if _, err := sqliteDB.ExecContext(ctx, "PRAGMA foreign_keys = ON"); err != nil {
		return fmt.Errorf("migrate: re-enable foreign keys: %w", err)
	}

	return nil
}

func migrateUsers(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating users")
	rows, err := pg.QueryContext(ctx, `SELECT id, username, role::text, password FROM users ORDER BY id`)
	if err != nil {
		return fmt.Errorf("migrate users: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO users (id, username, role, password) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate users: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var id int64
		var username, role string
		var password []byte
		if err := rows.Scan(&id, &username, &role, &password); err != nil {
			return fmt.Errorf("migrate users: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, id, username, role, password); err != nil {
			return fmt.Errorf("migrate users: insert id=%d: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate users: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d users", n)
	return nil
}

func migrateApiKeys(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating api_keys")
	rows, err := pg.QueryContext(ctx, `SELECT id, key, user_id, created_at, label FROM api_keys ORDER BY id`)
	if err != nil {
		return fmt.Errorf("migrate api_keys: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO api_keys (id, key, user_id, created_at, label) VALUES (?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate api_keys: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var id, userID int64
		var key, label string
		var createdAt time.Time
		if err := rows.Scan(&id, &key, &userID, &createdAt, &label); err != nil {
			return fmt.Errorf("migrate api_keys: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, id, key, userID, createdAt.Unix(), label); err != nil {
			return fmt.Errorf("migrate api_keys: insert id=%d: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate api_keys: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d api_keys", n)
	return nil
}

func migrateSessions(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating sessions")
	rows, err := pg.QueryContext(ctx, `SELECT id::text, user_id, created_at, expires_at, persistent FROM sessions ORDER BY created_at`)
	if err != nil {
		return fmt.Errorf("migrate sessions: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO sessions (id, user_id, created_at, expires_at, persistent) VALUES (?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate sessions: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var id string
		var userID int64
		var createdAt, expiresAt time.Time
		var persistent bool
		if err := rows.Scan(&id, &userID, &createdAt, &expiresAt, &persistent); err != nil {
			return fmt.Errorf("migrate sessions: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, id, userID, createdAt.Unix(), expiresAt.Unix(), boolToInt(persistent)); err != nil {
			return fmt.Errorf("migrate sessions: insert id=%s: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate sessions: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d sessions", n)
	return nil
}

func migrateArtists(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating artists")
	rows, err := pg.QueryContext(ctx, `SELECT id, musicbrainz_id::text, image::text, image_source FROM artists ORDER BY id`)
	if err != nil {
		return fmt.Errorf("migrate artists: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO artists (id, musicbrainz_id, image, image_source) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate artists: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var id int64
		var mbzID, image, imageSrc sql.NullString
		if err := rows.Scan(&id, &mbzID, &image, &imageSrc); err != nil {
			return fmt.Errorf("migrate artists: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, id, mbzID, image, imageSrc); err != nil {
			return fmt.Errorf("migrate artists: insert id=%d: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate artists: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d artists", n)
	return nil
}

func migrateArtistAliases(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating artist_aliases")
	rows, err := pg.QueryContext(ctx, `SELECT artist_id, alias, source, is_primary FROM artist_aliases`)
	if err != nil {
		return fmt.Errorf("migrate artist_aliases: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO artist_aliases (artist_id, alias, source, is_primary) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate artist_aliases: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var artistID int64
		var alias, source string
		var isPrimary bool
		if err := rows.Scan(&artistID, &alias, &source, &isPrimary); err != nil {
			return fmt.Errorf("migrate artist_aliases: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, artistID, alias, source, boolToInt(isPrimary)); err != nil {
			return fmt.Errorf("migrate artist_aliases: insert artist_id=%d alias=%q: %w", artistID, alias, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate artist_aliases: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d artist_aliases", n)
	return nil
}

func migrateReleases(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating releases")
	rows, err := pg.QueryContext(ctx, `SELECT id, musicbrainz_id::text, image::text, various_artists, image_source FROM releases ORDER BY id`)
	if err != nil {
		return fmt.Errorf("migrate releases: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO releases (id, musicbrainz_id, image, various_artists, image_source) VALUES (?,?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate releases: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var id int64
		var mbzID, image, imageSrc sql.NullString
		var variousArtists bool
		if err := rows.Scan(&id, &mbzID, &image, &variousArtists, &imageSrc); err != nil {
			return fmt.Errorf("migrate releases: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, id, mbzID, image, boolToInt(variousArtists), imageSrc); err != nil {
			return fmt.Errorf("migrate releases: insert id=%d: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate releases: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d releases", n)
	return nil
}

func migrateReleaseAliases(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating release_aliases")
	rows, err := pg.QueryContext(ctx, `SELECT release_id, alias, source, is_primary FROM release_aliases`)
	if err != nil {
		return fmt.Errorf("migrate release_aliases: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO release_aliases (release_id, alias, source, is_primary) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate release_aliases: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var releaseID int64
		var alias, source string
		var isPrimary bool
		if err := rows.Scan(&releaseID, &alias, &source, &isPrimary); err != nil {
			return fmt.Errorf("migrate release_aliases: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, releaseID, alias, source, boolToInt(isPrimary)); err != nil {
			return fmt.Errorf("migrate release_aliases: insert release_id=%d alias=%q: %w", releaseID, alias, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate release_aliases: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d release_aliases", n)
	return nil
}

func migrateArtistReleases(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating artist_releases")
	rows, err := pg.QueryContext(ctx, `SELECT artist_id, release_id, is_primary FROM artist_releases`)
	if err != nil {
		return fmt.Errorf("migrate artist_releases: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO artist_releases (artist_id, release_id, is_primary) VALUES (?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate artist_releases: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var artistID, releaseID int64
		var isPrimary bool
		if err := rows.Scan(&artistID, &releaseID, &isPrimary); err != nil {
			return fmt.Errorf("migrate artist_releases: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, artistID, releaseID, boolToInt(isPrimary)); err != nil {
			return fmt.Errorf("migrate artist_releases: insert artist_id=%d release_id=%d: %w", artistID, releaseID, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate artist_releases: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d artist_releases", n)
	return nil
}

func migrateTracks(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating tracks")
	rows, err := pg.QueryContext(ctx, `SELECT id, musicbrainz_id::text, release_id, duration FROM tracks ORDER BY id`)
	if err != nil {
		return fmt.Errorf("migrate tracks: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO tracks (id, musicbrainz_id, release_id, duration) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate tracks: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var id, releaseID int64
		var duration int32
		var mbzID sql.NullString
		if err := rows.Scan(&id, &mbzID, &releaseID, &duration); err != nil {
			return fmt.Errorf("migrate tracks: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, id, mbzID, releaseID, duration); err != nil {
			return fmt.Errorf("migrate tracks: insert id=%d: %w", id, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate tracks: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d tracks", n)
	return nil
}

func migrateTrackAliases(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating track_aliases")
	rows, err := pg.QueryContext(ctx, `SELECT track_id, alias, source, is_primary FROM track_aliases`)
	if err != nil {
		return fmt.Errorf("migrate track_aliases: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO track_aliases (track_id, alias, source, is_primary) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate track_aliases: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var trackID int64
		var alias, source string
		var isPrimary bool
		if err := rows.Scan(&trackID, &alias, &source, &isPrimary); err != nil {
			return fmt.Errorf("migrate track_aliases: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, trackID, alias, source, boolToInt(isPrimary)); err != nil {
			return fmt.Errorf("migrate track_aliases: insert track_id=%d alias=%q: %w", trackID, alias, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate track_aliases: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d track_aliases", n)
	return nil
}

func migrateArtistTracks(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating artist_tracks")
	rows, err := pg.QueryContext(ctx, `SELECT artist_id, track_id, is_primary FROM artist_tracks`)
	if err != nil {
		return fmt.Errorf("migrate artist_tracks: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT INTO artist_tracks (artist_id, track_id, is_primary) VALUES (?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate artist_tracks: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var artistID, trackID int64
		var isPrimary bool
		if err := rows.Scan(&artistID, &trackID, &isPrimary); err != nil {
			return fmt.Errorf("migrate artist_tracks: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, artistID, trackID, boolToInt(isPrimary)); err != nil {
			return fmt.Errorf("migrate artist_tracks: insert artist_id=%d track_id=%d: %w", artistID, trackID, err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate artist_tracks: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d artist_tracks", n)
	return nil
}

func migrateListens(ctx context.Context, pg *sql.DB, tx *sql.Tx, l *zerolog.Logger) error {
	l.Info().Msg("Migrate: migrating listens")
	rows, err := pg.QueryContext(ctx, `SELECT track_id, listened_at, user_id, COALESCE(client, '') FROM listens ORDER BY listened_at`)
	if err != nil {
		return fmt.Errorf("migrate listens: query: %w", err)
	}
	defer rows.Close()

	stmt, err := tx.PrepareContext(ctx, `INSERT OR IGNORE INTO listens (track_id, listened_at, user_id, client) VALUES (?,?,?,?)`)
	if err != nil {
		return fmt.Errorf("migrate listens: prepare: %w", err)
	}
	defer stmt.Close()

	var n int
	for rows.Next() {
		var trackID, userID int64
		var listenedAt time.Time
		var client string
		if err := rows.Scan(&trackID, &listenedAt, &userID, &client); err != nil {
			return fmt.Errorf("migrate listens: scan: %w", err)
		}
		if _, err := stmt.ExecContext(ctx, trackID, listenedAt.Unix(), userID, client); err != nil {
			return fmt.Errorf("migrate listens: insert track_id=%d listened_at=%d: %w", trackID, listenedAt.Unix(), err)
		}
		n++
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("migrate listens: rows: %w", err)
	}
	l.Info().Msgf("Migrate: migrated %d listens", n)
	return nil
}
