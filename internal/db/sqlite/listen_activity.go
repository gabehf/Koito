package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
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

	loc := opts.Timezone

	var rows *sql.Rows
	var err error

	switch {
	case opts.ArtistID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT (l.listened_at / 3600) * 3600 AS hour_bucket, COUNT(*) AS listen_count
			FROM listens l
			JOIN artist_tracks at2 ON l.track_id = at2.track_id
			WHERE l.listened_at >= ? AND l.listened_at < ? AND at2.artist_id = ?
			GROUP BY hour_bucket`,
			t1.Unix(), t2.Unix(), opts.ArtistID,
		)
	case opts.AlbumID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT (l.listened_at / 3600) * 3600 AS hour_bucket, COUNT(*) AS listen_count
			FROM listens l
			JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at >= ? AND l.listened_at < ? AND t.release_id = ?
			GROUP BY hour_bucket`,
			t1.Unix(), t2.Unix(), opts.AlbumID,
		)
	case opts.TrackID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT (listened_at / 3600) * 3600 AS hour_bucket, COUNT(*) AS listen_count
			FROM listens
			WHERE listened_at >= ? AND listened_at < ? AND track_id = ?
			GROUP BY hour_bucket`,
			t1.Unix(), t2.Unix(), opts.TrackID,
		)
	default:
		rows, err = s.db.QueryContext(ctx, `
			SELECT (listened_at / 3600) * 3600 AS hour_bucket, COUNT(*) AS listen_count
			FROM listens
			WHERE listened_at >= ? AND listened_at < ?
			GROUP BY hour_bucket`,
			t1.Unix(), t2.Unix(),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("GetListenActivity: %w", err)
	}
	defer rows.Close()

	countsByDay := make(map[time.Time]int)

	for rows.Next() {
		var hourUnix int64
		var count int
		if err := rows.Scan(&hourUnix, &count); err != nil {
			return nil, err
		}

		// Convert the UTC hour block to a Go time object localized to the user's timezone
		t := time.Unix(hourUnix, 0).In(loc)

		// Strip the hours/minutes/seconds to collapse it into a pure calendar day
		day := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, loc)

		countsByDay[day] += count
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Convert the map to the final slice and sort it chronologically
	activity := make([]db.ListenActivityItem, 0, len(countsByDay))
	for day, count := range countsByDay {
		activity = append(activity, db.ListenActivityItem{
			Start:   day,
			Listens: int64(count),
		})
	}

	sort.Slice(activity, func(i, j int) bool {
		return activity[i].Start.Before(activity[j].Start)
	})

	return activity, nil
}
