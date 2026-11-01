package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
)

// backupBeforeProfiles takes a consistent SQLite snapshot, including committed WAL
// pages. Unlike copying koito.db, VACUUM INTO also works with an active WAL.
// A backup failure prevents migration. Fresh or already-migrated DBs are skipped.
func backupBeforeProfiles(ctx context.Context, conn *sql.DB, databasePath string) (string, error) {
	var existing bool
	if err := conn.QueryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM sqlite_master WHERE type = 'table' AND name = 'users')`).Scan(&existing); err != nil {
		return "", err
	}
	if !existing {
		return "", nil
	}
	var applied bool
	err := conn.QueryRowContext(ctx, `SELECT is_applied FROM goose_db_version WHERE version_id = 3 ORDER BY id DESC LIMIT 1`).Scan(&applied)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	if applied {
		return "", nil
	}
	f, err := os.CreateTemp(filepath.Dir(databasePath), filepath.Base(databasePath)+".pre-profiles-*.bak")
	if err != nil {
		return "", err
	}
	filename := f.Name()
	if err := f.Close(); err != nil {
		os.Remove(filename)
		return "", err
	}
	if _, err := conn.ExecContext(ctx, `VACUUM INTO ?`, filename); err != nil {
		os.Remove(filename)
		return "", fmt.Errorf("snapshot: %w", err)
	}
	return filename, nil
}
