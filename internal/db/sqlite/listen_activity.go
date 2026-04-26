package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
)

func (s *Sqlite) GetListenActivity(ctx context.Context, opts db.ListenActivityOpts) ([]db.ListenActivityItem, error) {
	if opts.Month != 0 && opts.Year == 0 {
		return nil, errors.New("GetListenActivity: year must be specified with month")
	}
	if opts.Range == 0 {
		opts.Range = db.DefaultRange
	}

	t1, t2 := db.ListenActivityOptsToTimes(opts)
	offset := tzOffset(opts.Timezone)

	// SQLite date strategy: add the UTC offset in seconds to the stored Unix epoch, then
	// format as a date. This correctly localises each listen to the user's calendar day.
	var rows *sql.Rows
	var err error

	switch {
	case opts.ArtistID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT date(listens.listened_at + ?, 'unixepoch') AS day, COUNT(*) AS listen_count
			FROM listens
			JOIN tracks t ON listens.track_id = t.id
			JOIN artist_tracks at2 ON t.id = at2.track_id
			WHERE listens.listened_at >= ? AND listens.listened_at < ?
			  AND at2.artist_id = ?
			GROUP BY day
			ORDER BY day`,
			offset, t1.Unix(), t2.Unix(), opts.ArtistID,
		)
	case opts.AlbumID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT date(listens.listened_at + ?, 'unixepoch') AS day, COUNT(*) AS listen_count
			FROM listens
			JOIN tracks t ON listens.track_id = t.id
			WHERE listens.listened_at >= ? AND listens.listened_at < ?
			  AND t.release_id = ?
			GROUP BY day
			ORDER BY day`,
			offset, t1.Unix(), t2.Unix(), opts.AlbumID,
		)
	case opts.TrackID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT date(listens.listened_at + ?, 'unixepoch') AS day, COUNT(*) AS listen_count
			FROM listens
			JOIN tracks t ON listens.track_id = t.id
			WHERE listens.listened_at >= ? AND listens.listened_at < ?
			  AND t.id = ?
			GROUP BY day
			ORDER BY day`,
			offset, t1.Unix(), t2.Unix(), opts.TrackID,
		)
	default:
		rows, err = s.db.QueryContext(ctx, `
			SELECT date(listened_at + ?, 'unixepoch') AS day, COUNT(*) AS listen_count
			FROM listens
			WHERE listened_at >= ? AND listened_at < ?
			GROUP BY day
			ORDER BY day`,
			offset, t1.Unix(), t2.Unix(),
		)
	}
	if err != nil {
		return nil, fmt.Errorf("GetListenActivity: %w", err)
	}
	defer rows.Close()

	var activity []db.ListenActivityItem
	for rows.Next() {
		var dayStr string
		var item db.ListenActivityItem
		if err := rows.Scan(&dayStr, &item.Listens); err != nil {
			return nil, err
		}
		t, err := time.Parse("2006-01-02", dayStr)
		if err != nil {
			return nil, fmt.Errorf("GetListenActivity: parse day %q: %w", dayStr, err)
		}
		item.Start = t
		activity = append(activity, item)
	}
	return activity, rows.Err()
}
