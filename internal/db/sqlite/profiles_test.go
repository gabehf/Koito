package sqlite

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gabehf/koito/internal/export"
	"github.com/gabehf/koito/internal/models"
	"os"
	"path/filepath"
	"testing"
	"time"

	migrations "github.com/gabehf/koito/db/migrations_sqlite"
	"github.com/gabehf/koito/internal/db"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
)

func migrationDB(t *testing.T, version int64) (*Sqlite, *goose.Provider, string) {
	t.Helper()
	filename := filepath.Join(t.TempDir(), "koito.db")
	conn, err := sql.Open("sqlite", filename+"?_pragma=foreign_keys(ON)&_pragma=journal_mode(WAL)")
	require.NoError(t, err)
	conn.SetMaxOpenConns(1)
	t.Cleanup(func() { conn.Close() })
	provider, err := goose.NewProvider(goose.DialectSQLite3, conn, migrations.Files)
	require.NoError(t, err)
	_, err = provider.UpTo(context.Background(), version)
	require.NoError(t, err)
	return &Sqlite{db: conn}, provider, filename
}

func execSQL(t *testing.T, s *Sqlite, query string, args ...any) {
	t.Helper()
	_, err := s.db.Exec(query, args...)
	require.NoError(t, err)
}

func scalar(t *testing.T, s *Sqlite, query string, args ...any) int {
	t.Helper()
	var value int
	require.NoError(t, s.db.QueryRow(query, args...).Scan(&value))
	return value
}

func seedMusic(t *testing.T, s *Sqlite) {
	t.Helper()
	execSQL(t, s, `INSERT INTO artists (id) VALUES (1)`)
	execSQL(t, s, `INSERT INTO artist_aliases VALUES (1,'Artist','User',1)`)
	execSQL(t, s, `INSERT INTO releases (id) VALUES (1)`)
	execSQL(t, s, `INSERT INTO artist_releases VALUES (1,1,1)`)
	execSQL(t, s, `INSERT INTO tracks (id,release_id,duration) VALUES (1,1,180)`)
}

func TestProfileMigrationPreservesLegacyDataAndBackup(t *testing.T) {
	ctx := context.Background()
	s, p, filename := migrationDB(t, 2)
	execSQL(t, s, `INSERT INTO users (id,username,password,role) VALUES (7,'everton',X'1234','admin')`)
	seedMusic(t, s)
	execSQL(t, s, `INSERT INTO api_keys VALUES (11,'preserved-token',7,123,'Navidrome')`)
	execSQL(t, s, `INSERT INTO listens VALUES (1,1700000000,7,'client')`)
	execSQL(t, s, `INSERT INTO sessions VALUES ('session',7,100,200,1)`)
	backup, err := backupBeforeProfiles(ctx, s.db, filename)
	require.NoError(t, err)
	require.NotEmpty(t, backup)
	info, err := os.Stat(backup)
	require.NoError(t, err)
	require.Equal(t, os.FileMode(0600), info.Mode().Perm())
	snapshot, err := sql.Open("sqlite", "file:"+backup+"?mode=ro")
	require.NoError(t, err)
	defer snapshot.Close()
	var count int
	require.NoError(t, snapshot.QueryRow(`SELECT COUNT(*) FROM listens`).Scan(&count))
	require.Equal(t, 1, count) // Committed data was still in the source WAL.
	require.NoError(t, snapshot.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE name='listening_profiles'`).Scan(&count))
	require.Zero(t, count)

	_, err = p.UpTo(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles WHERE id=7 AND slug='default' AND display_name='everton' AND created_at>0`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM api_keys WHERE id=11 AND key='preserved-token' AND user_id=7 AND profile_id=7 AND created_at=123 AND label='Navidrome'`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE profile_id=7 AND track_id=1 AND listened_at=1700000000 AND user_id=7 AND client='client'`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM users WHERE id=7 AND username='everton' AND password=X'1234' AND role='admin'`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM sessions WHERE id='session' AND user_id=7 AND persistent=1`))
	assertForeignKeys(t, s)
	backup, err = backupBeforeProfiles(ctx, s.db, filename)
	require.NoError(t, err)
	require.Empty(t, backup)
	_, err = p.UpTo(ctx, 3)
	require.NoError(t, err)
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles`))
	_, err = p.Down(ctx)
	require.Error(t, err, "downgrade must not silently discard cross-profile listens")
	require.Contains(t, err.Error(), "restore the pre-profiles backup")
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listens`))
}

func TestProfileMigrationMultipleUsers(t *testing.T) {
	s, p, _ := migrationDB(t, 2)
	seedMusic(t, s)
	for i, name := range []string{"admin", "a_b", "a.b", "user-1"} {
		id := i + 1
		execSQL(t, s, `INSERT INTO users (id,username,password) VALUES (?,?,X'1234')`, id, name)
		execSQL(t, s, `INSERT INTO listens VALUES (1,?,?,'legacy')`, 1700000000+id, id)
		execSQL(t, s, `INSERT INTO api_keys (key,user_id,created_at) VALUES (?,?,123)`, name, id)
	}
	_, err := p.UpTo(context.Background(), 3)
	require.NoError(t, err)
	require.Equal(t, 4, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles`))
	require.Equal(t, 4, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles WHERE slug='user-' || id`))
	require.Equal(t, 4, scalar(t, s, `SELECT COUNT(*) FROM listens l JOIN legacy_user_profiles m ON m.user_id=l.user_id AND m.profile_id=l.profile_id`))
	require.Equal(t, 4, scalar(t, s, `SELECT COUNT(*) FROM api_keys k JOIN legacy_user_profiles m ON m.user_id=k.user_id AND m.profile_id=k.profile_id`))
	assertForeignKeys(t, s)
}

func TestProfilesFreshDatabaseAndWriteCompatibility(t *testing.T) {
	ctx := context.Background()
	s, _, filename := migrationDB(t, 3)
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles`))
	execSQL(t, s, `INSERT INTO users (username,password) VALUES ('admin',X'1234')`)
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles WHERE slug='default'`))
	execSQL(t, s, `INSERT INTO users (username,password) VALUES ('second',X'1234')`)
	seedMusic(t, s)
	ts := time.Unix(1700000000, 0)
	for _, userID := range []int32{1, 2} {
		opts := db.SaveListenOpts{TrackID: 1, UserID: userID, Time: ts}
		require.NoError(t, s.SaveListen(ctx, opts))
		require.NoError(t, s.SaveListen(ctx, opts)) // Deduplicate within the profile.
	}
	require.Equal(t, 2, scalar(t, s, `SELECT COUNT(*) FROM listens`))
	require.Error(t, s.SaveListen(ctx, db.SaveListenOpts{TrackID: 1, UserID: 999, Time: ts}))
	require.Error(t, s.SaveListen(ctx, db.SaveListenOpts{TrackID: 1, Time: ts}))
	require.Error(t, s.SaveListen(ctx, db.SaveListenOpts{TrackID: 1, UserID: 1, ProfileID: 2, Time: ts}))
	require.Error(t, s.SaveListen(ctx, db.SaveListenOpts{TrackID: 999, ProfileID: 1, Time: ts}))
	key, err := s.SaveApiKey(ctx, db.SaveApiKeyOpts{Key: "first", UserID: 1, Label: "phone"})
	require.NoError(t, err)
	require.EqualValues(t, 1, key.ProfileID)
	keys, err := s.GetApiKeysByUserID(ctx, 1)
	require.NoError(t, err)
	require.Len(t, keys, 1)
	require.EqualValues(t, 1, keys[0].ProfileID)
	_, err = s.SaveApiKey(ctx, db.SaveApiKeyOpts{Key: "second", ProfileID: 1})
	require.NoError(t, err)
	_, err = s.SaveApiKey(ctx, db.SaveApiKeyOpts{Key: "other-profile", UserID: 2})
	require.NoError(t, err)
	require.Equal(t, 2, scalar(t, s, `SELECT COUNT(*) FROM api_keys WHERE profile_id=1`))
	require.Error(t, s.SaveListen(ctx, db.SaveListenOpts{TrackID: 1, ProfileID: 999, Time: ts}))

	// Public identities need no login account or password.
	execSQL(t, s, `INSERT INTO listening_profiles (id,slug,display_name,created_at) VALUES (10,'independent','Independent',123)`)
	require.NoError(t, s.SaveListen(ctx, db.SaveListenOpts{TrackID: 1, ProfileID: 10, Time: ts}))
	_, err = s.SaveApiKey(ctx, db.SaveApiKeyOpts{Key: "independent", ProfileID: 10})
	require.NoError(t, err)
	execSQL(t, s, `DELETE FROM api_keys WHERE key='independent'`)
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE profile_id=10`))
	execSQL(t, s, `DELETE FROM users WHERE id=1`)
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE profile_id=1`))
	require.Equal(t, 2, scalar(t, s, `SELECT COUNT(*) FROM api_keys WHERE profile_id=1`))
	execSQL(t, s, `DELETE FROM listening_profiles WHERE id=2`)
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE profile_id=2`))
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM api_keys WHERE profile_id=2`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM tracks`))
	require.Equal(t, 2, scalar(t, s, `SELECT COUNT(*) FROM listens`))
	assertForeignKeys(t, s)
	backup, err := backupBeforeProfiles(ctx, s.db, filename)
	require.NoError(t, err)
	require.Empty(t, backup)
}

func TestProfileSchemaRejectsInvalidSlugsAndUnscopedWrites(t *testing.T) {
	s, _, _ := migrationDB(t, 3)
	for _, slug := range []string{"", "admin", "apis", "image", "public", "rewind", "A", "a_b", "a.b", "-name", "name-"} {
		_, err := s.db.Exec(`INSERT INTO listening_profiles (slug,display_name,created_at) VALUES (?,'Name',123)`, slug)
		require.Error(t, err, "invalid or reserved slug: %q", slug)
	}
	execSQL(t, s, `INSERT INTO listening_profiles (slug,display_name,created_at) VALUES ('everton-berger','Éverton Berger',123)`)
	_, err := s.db.Exec(`INSERT INTO listening_profiles (slug,display_name,created_at) VALUES ('everton-berger','Duplicate',123)`)
	require.Error(t, err)
	_, err = s.db.Exec(`INSERT INTO api_keys (key,created_at) VALUES ('unscoped',123)`)
	require.Error(t, err)
	seedMusic(t, s)
	_, err = s.db.Exec(`INSERT INTO listens (track_id,listened_at) VALUES (1,123)`)
	require.Error(t, err)
}

func TestBackupFailureLeavesVersionTwoUntouched(t *testing.T) {
	s, p, filename := migrationDB(t, 2)
	_, err := backupBeforeProfiles(context.Background(), s.db, filepath.Join(filename, "missing.db"))
	require.Error(t, err)
	version, err := p.GetDBVersion(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 2, version)
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM sqlite_master WHERE name='listening_profiles'`))
}

func TestProfileMigrationRollsBackInvalidLegacyData(t *testing.T) {
	s, p, _ := migrationDB(t, 2)
	execSQL(t, s, `PRAGMA foreign_keys=OFF`)
	execSQL(t, s, `INSERT INTO api_keys (key,user_id,created_at) VALUES ('orphan',999,123)`)
	execSQL(t, s, `PRAGMA foreign_keys=ON`)
	_, err := p.UpTo(context.Background(), 3)
	require.Error(t, err)
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM api_keys WHERE key='orphan'`))
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM sqlite_master WHERE name='listening_profiles'`))
	version, err := p.GetDBVersion(context.Background())
	require.NoError(t, err)
	require.EqualValues(t, 2, version)
}

func assertForeignKeys(t *testing.T, s *Sqlite) {
	t.Helper()
	rows, err := s.db.Query(`PRAGMA foreign_key_check`)
	require.NoError(t, err)
	defer rows.Close()
	require.False(t, rows.Next(), "database must have no dangling references")
	require.NoError(t, rows.Err())
}

// Opt-in acceptance test. The source is opened read-only; only its snapshot is migrated.
func TestProfileMigrationExampleCopy(t *testing.T) {
	filename := os.Getenv("KOITO_MIGRATION_TEST_DB")
	if filename == "" {
		t.Skip("set KOITO_MIGRATION_TEST_DB to test a read-only snapshot of an example DB")
	}
	source, err := sql.Open("sqlite", "file:"+filename+"?mode=ro")
	require.NoError(t, err)
	defer source.Close()
	copyPath := filepath.Join(t.TempDir(), "example.db")
	_, err = source.Exec(`VACUUM INTO ?`, copyPath)
	require.NoError(t, err)
	copyDB, err := sql.Open("sqlite", copyPath+"?_pragma=foreign_keys(ON)")
	require.NoError(t, err)
	defer copyDB.Close()
	copyDB.SetMaxOpenConns(1)
	s := &Sqlite{db: copyDB}
	queries := []string{
		`SELECT id,username,role,password FROM users ORDER BY id`,
		`SELECT id,key,user_id,created_at,label FROM api_keys ORDER BY id`,
		`SELECT track_id,listened_at,user_id,client FROM listens ORDER BY track_id,listened_at`,
		`SELECT * FROM sessions ORDER BY id`,
		`SELECT * FROM artists ORDER BY id`,
		`SELECT * FROM releases ORDER BY id`,
		`SELECT * FROM tracks ORDER BY id`,
	}
	digests := make([][32]byte, len(queries))
	for i, query := range queries {
		digests[i] = queryDigest(t, copyDB, query)
	}
	versionBefore := scalar(t, s, `SELECT MAX(version_id) FROM goose_db_version WHERE is_applied=1`)
	backup, err := backupBeforeProfiles(context.Background(), copyDB, copyPath)
	require.NoError(t, err)
	if versionBefore < 3 {
		require.NotEmpty(t, backup)
	} else {
		require.Empty(t, backup)
	}
	p, err := goose.NewProvider(goose.DialectSQLite3, copyDB, migrations.Files)
	require.NoError(t, err)
	_, err = p.Up(context.Background())
	require.NoError(t, err)
	for i, query := range queries {
		require.Equal(t, digests[i], queryDigest(t, copyDB, query), "legacy data changed: %s", query)
	}
	assertForeignKeys(t, s)
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM listens l LEFT JOIN listening_profiles p ON p.id=l.profile_id WHERE p.id IS NULL`))
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM api_keys k LEFT JOIN listening_profiles p ON p.id=k.profile_id WHERE p.id IS NULL`))
	versionAfter, err := p.GetDBVersion(context.Background())
	require.NoError(t, err)
	t.Logf("Snapshot migrated to v%d: %d profiles, %d listens, %d keys; legacy data digests unchanged", versionAfter, scalar(t, s, `SELECT COUNT(*) FROM listening_profiles`), scalar(t, s, `SELECT COUNT(*) FROM listens`), scalar(t, s, `SELECT COUNT(*) FROM api_keys`))
}

func queryDigest(t *testing.T, conn *sql.DB, query string) [32]byte {
	t.Helper()
	rows, err := conn.Query(query)
	require.NoError(t, err)
	defer rows.Close()
	columns, err := rows.Columns()
	require.NoError(t, err)
	hash := sha256.New()
	for rows.Next() {
		values := make([]any, len(columns))
		pointers := make([]any, len(columns))
		for i := range values {
			pointers[i] = &values[i]
		}
		require.NoError(t, rows.Scan(pointers...))
		fmt.Fprintf(hash, "%#v\n", values)
	}
	require.NoError(t, rows.Err())
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func TestProfileKeysCannotMoveAndRelayDoesNotReassign(t *testing.T) {
	s, _, _ := migrationDB(t, 3)
	ctx := context.Background()
	execSQL(t, s, `INSERT INTO users(username,password,role) VALUES ('admin',X'1234','admin')`)
	require.NoError(t, s.InitializeLegacyRelay(ctx, "https://relay.test", "legacy-token"))
	url, token, err := s.GetProfileRelay(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "https://relay.test", url)
	require.Equal(t, "legacy-token", token)
	require.NoError(t, s.UpdateListeningProfile(ctx, 1, "Renamed", "renamed"))
	require.NoError(t, s.SetProfileRelay(ctx, 1, "", ""))
	require.NoError(t, s.InitializeLegacyRelay(ctx, "https://relay.test", "legacy-token"))
	url, _, err = s.GetProfileRelay(ctx, 1)
	require.NoError(t, err)
	require.Empty(t, url)
	execSQL(t, s, `INSERT INTO listening_profiles(id,slug,display_name,created_at) VALUES (2,'other','Other',123)`)
	_, err = s.SaveApiKey(ctx, db.SaveApiKeyOpts{Key: "fixed-profile", ProfileID: 1})
	require.NoError(t, err)
	_, err = s.db.Exec(`UPDATE api_keys SET profile_id=2 WHERE key='fixed-profile'`)
	require.Error(t, err)
	execSQL(t, s, `DELETE FROM listening_profiles WHERE id=1`)
	require.NoError(t, s.InitializeLegacyRelay(ctx, "https://relay.test", "legacy-token"))
	url, _, err = s.GetProfileRelay(ctx, 2)
	require.NoError(t, err)
	require.Empty(t, url)
	assertForeignKeys(t, s)
}

func TestMergeTracksPreservesProfilesAtSameTimestamp(t *testing.T) {
	s, _, _ := migrationDB(t, 3)
	ctx := context.Background()
	execSQL(t, s, `INSERT INTO users(username,password) VALUES ('one',X'1234'),('two',X'1234')`)
	seedMusic(t, s)
	execSQL(t, s, `INSERT INTO tracks(id,release_id,duration) VALUES (2,1,180)`)
	execSQL(t, s, `INSERT INTO artist_tracks(artist_id,track_id) VALUES (1,1),(1,2)`)
	ts := time.Unix(1700000000, 0)
	for _, pair := range [][2]int32{{1, 1}, {1, 2}, {2, 2}} {
		require.NoError(t, s.SaveListen(ctx, db.SaveListenOpts{ProfileID: pair[0], TrackID: pair[1], Time: ts}))
	}
	require.NoError(t, s.MergeTracks(ctx, 1, 2))
	require.Equal(t, 2, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE track_id=2`))
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE track_id=1`))
	require.Error(t, s.MergeTracks(ctx, 2, 2))
	require.Equal(t, 2, scalar(t, s, `SELECT COUNT(*) FROM listens`))
	assertForeignKeys(t, s)
}

func TestProfileExportPaginationDoesNotSkipSameTimestamp(t *testing.T) {
	s, _, _ := migrationDB(t, 3)
	execSQL(t, s, `INSERT INTO users(username,password) VALUES ('listener',X'1234')`)
	seedMusic(t, s)
	execSQL(t, s, `INSERT INTO tracks(id,release_id,duration) VALUES (2,1,180)`)
	execSQL(t, s, `WITH RECURSIVE seq(n) AS (SELECT 1 UNION ALL SELECT n+1 FROM seq WHERE n<1000)
       INSERT INTO listens(profile_id,track_id,listened_at)
       SELECT 1,CASE WHEN n=1 THEN 2 ELSE 1 END,1700000000+n FROM seq`)
	execSQL(t, s, `INSERT INTO listens(profile_id,track_id,listened_at) VALUES (1,2,1700001000)`)
	var out bytes.Buffer
	require.NoError(t, export.ExportData(context.Background(), &models.ListeningProfile{ID: 1, DisplayName: "Listener"}, s, &out))
	var result export.KoitoExport
	require.NoError(t, json.Unmarshal(out.Bytes(), &result))
	require.Len(t, result.Listens, 1001)
	require.True(t, time.Unix(1700001000, 0).Equal(result.Listens[1000].ListenedAt))
}

func TestConsolidatedProfileMigrationRollsBackLateFailure(t *testing.T) {
	s, p, _ := migrationDB(t, 2)
	ctx := context.Background()
	execSQL(t, s, `INSERT INTO users(id,username,password,role) VALUES(1,'legacy-admin',X'1234','admin')`)
	seedMusic(t, s)
	execSQL(t, s, `INSERT INTO api_keys(key,user_id,created_at) VALUES('legacy-token',1,123)`)
	execSQL(t, s, `INSERT INTO listens(track_id,listened_at,user_id) VALUES(1,1700000000,1)`)
	// A failure in the relay section must roll back the earlier profile rebuilds.
	execSQL(t, s, `CREATE TABLE profile_relays(conflict TEXT)`)
	_, err := p.Up(ctx)
	require.Error(t, err)
	version, err := p.GetDBVersion(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 2, version)
	require.Zero(t, scalar(t, s, `SELECT COUNT(*) FROM sqlite_master WHERE name='listening_profiles'`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM api_keys WHERE key='legacy-token' AND user_id=1`))
	require.Equal(t, 1, scalar(t, s, `SELECT COUNT(*) FROM listens WHERE user_id=1 AND listened_at=1700000000`))
	// Retrying after removing the conflict applies the entire upgrade only once.
	execSQL(t, s, `DROP TABLE profile_relays`)
	_, err = p.Up(ctx)
	require.NoError(t, err)
	version, err = p.GetDBVersion(ctx)
	require.NoError(t, err)
	require.EqualValues(t, 3, version)
	require.NoError(t, s.SetProfileRelay(ctx, 1, "https://relay.test", "relay-token"))
	url, token, err := s.GetProfileRelay(ctx, 1)
	require.NoError(t, err)
	require.Equal(t, "https://relay.test", url)
	require.Equal(t, "relay-token", token)
	_, err = s.db.Exec(`UPDATE api_keys SET profile_id=2 WHERE key='legacy-token'`)
	require.ErrorContains(t, err, "API keys cannot be transferred")
	assertForeignKeys(t, s)
}
