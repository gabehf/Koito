package sqlite

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabehf/koito/internal/db"
)

func (s *Sqlite) CountListens(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var count int64
	err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM listens WHERE listened_at BETWEEN ? AND ?`,
		t1.Unix(), t2.Unix()).Scan(&count)
	return count, err
}

func (s *Sqlite) CountListensToItem(ctx context.Context, opts db.TimeListenedOpts) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)
	var count int64
	var err error
	switch {
	case opts.ArtistID > 0:
		err = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens l
			JOIN artist_tracks at2 ON l.track_id = at2.track_id
			WHERE l.listened_at BETWEEN ? AND ? AND at2.artist_id = ?`,
			t1.Unix(), t2.Unix(), opts.ArtistID).Scan(&count)
	case opts.AlbumID > 0:
		err = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens l
			JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?`,
			t1.Unix(), t2.Unix(), opts.AlbumID).Scan(&count)
	case opts.TrackID > 0:
		err = s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens
			WHERE listened_at BETWEEN ? AND ? AND track_id = ?`,
			t1.Unix(), t2.Unix(), opts.TrackID).Scan(&count)
	default:
		return 0, errors.New("CountListensToItem: an id must be provided")
	}
	if err != nil {
		return 0, fmt.Errorf("CountListensToItem: %w", err)
	}
	return count, nil
}

func (s *Sqlite) CountTimeListened(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(timeframe)
	var seconds int64
	err := s.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(t.duration), 0)
		FROM listens l JOIN tracks t ON l.track_id = t.id
		WHERE l.listened_at BETWEEN ? AND ?`,
		t1.Unix(), t2.Unix()).Scan(&seconds)
	return seconds, err
}

func (s *Sqlite) CountTimeListenedToItem(ctx context.Context, opts db.TimeListenedOpts) (int64, error) {
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)
	var seconds int64
	var err error
	switch {
	case opts.ArtistID > 0:
		err = s.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(t.duration), 0)
			FROM listens l JOIN tracks t ON l.track_id = t.id
			JOIN artist_tracks at2 ON t.id = at2.track_id
			WHERE l.listened_at BETWEEN ? AND ? AND at2.artist_id = ?`,
			t1.Unix(), t2.Unix(), opts.ArtistID).Scan(&seconds)
	case opts.AlbumID > 0:
		err = s.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(t.duration), 0)
			FROM listens l JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?`,
			t1.Unix(), t2.Unix(), opts.AlbumID).Scan(&seconds)
	case opts.TrackID > 0:
		err = s.db.QueryRowContext(ctx, `
			SELECT COALESCE(SUM(t.duration), 0)
			FROM listens l JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.id = ?`,
			t1.Unix(), t2.Unix(), opts.TrackID).Scan(&seconds)
	default:
		return 0, errors.New("CountTimeListenedToItem: an id must be provided")
	}
	if err != nil {
		return 0, fmt.Errorf("CountTimeListenedToItem: %w", err)
	}
	return seconds, nil
}
