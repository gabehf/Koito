package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/gabehf/koito/internal/db"
)

// A statement-local CTE scopes every reference to listens, including nested
// rankings and first-listen aggregates. main.listens is the physical table.
// Internal catalog jobs have no profile scope; HTTP middleware always sets one
// for public listening reads. Values here are typed integers, never URL strings.
func scopedRead(ctx context.Context, query string) string {
	id := db.ProfileID(ctx)
	if id <= 0 {
		return query
	}
	scope := fmt.Sprintf("listens AS (SELECT * FROM main.listens WHERE profile_id = %d)", id)
	q := strings.TrimSpace(query)
	upper := strings.ToUpper(q)
	if strings.HasPrefix(upper, "WITH RECURSIVE ") {
		return "WITH RECURSIVE " + scope + ", " + q[len("WITH RECURSIVE "):]
	}
	if strings.HasPrefix(upper, "WITH ") {
		return "WITH " + scope + ", " + q[len("WITH "):]
	}
	return "WITH " + scope + " " + q
}
func (s *Sqlite) queryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return s.db.QueryContext(ctx, scopedRead(ctx, query), args...)
}
func (s *Sqlite) queryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return s.db.QueryRowContext(ctx, scopedRead(ctx, query), args...)
}

func (s *Sqlite) requireRelated(ctx context.Context, kind string, id int32) error {
	if db.ProfileID(ctx) <= 0 {
		return nil
	}
	var condition string
	switch kind {
	case "track":
		condition = "l.track_id = ?"
	case "album":
		condition = "t.release_id = ?"
	case "artist":
		condition = "EXISTS (SELECT 1 FROM artist_tracks a WHERE a.track_id=l.track_id AND a.artist_id=?)"
	default:
		return fmt.Errorf("unknown catalog entity")
	}
	var found bool
	err := s.queryRowContext(ctx, `SELECT EXISTS(SELECT 1 FROM listens l JOIN tracks t ON t.id=l.track_id WHERE `+condition+`)`, id).Scan(&found)
	if err != nil {
		return err
	}
	if !found {
		return db.ErrNotFound
	}
	return nil
}
