package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
)

func (s *Sqlite) SaveListen(ctx context.Context, opts db.SaveListenOpts) error {
	if opts.TrackID == 0 {
		return errors.New("SaveListen: required parameter TrackID missing")
	}
	if opts.Time.IsZero() {
		opts.Time = time.Now()
	}
	client := ""
	if opts.Client != "" {
		client = opts.Client
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO listens (track_id, listened_at, user_id, client) VALUES (?,?,?,?)`,
		opts.TrackID, opts.Time.Unix(), opts.UserID, client,
	)
	return err
}

func (s *Sqlite) DeleteListen(ctx context.Context, trackId int32, listenedAt time.Time) error {
	if trackId == 0 {
		return errors.New("DeleteListen: required parameter 'trackId' missing")
	}
	_, err := s.db.ExecContext(ctx,
		`DELETE FROM listens WHERE track_id = ? AND listened_at = ?`,
		trackId, listenedAt.Unix(),
	)
	return err
}

func (s *Sqlite) GetListensPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[*models.Listen], error) {
	if opts.Limit == 0 {
		opts.Limit = defaultItemsPerPage
	}
	offset := (opts.Page - 1) * opts.Limit
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)

	var (
		rows  *sql.Rows
		err   error
		count int64
	)

	switch {
	case opts.TrackID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT l.listened_at, l.track_id, t.title
			FROM listens l
			JOIN tracks_with_title t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.id = ?
			ORDER BY l.listened_at DESC LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.TrackID, opts.Limit, offset,
		)
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated (by track): %w", err)
		}
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM listens WHERE listened_at BETWEEN ? AND ? AND track_id = ?`,
			t1.Unix(), t2.Unix(), opts.TrackID).Scan(&count)

	case opts.AlbumID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT l.listened_at, l.track_id, t.title
			FROM listens l
			JOIN tracks_with_title t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?
			ORDER BY l.listened_at DESC LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.AlbumID, opts.Limit, offset,
		)
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated (by album): %w", err)
		}
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens l JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?`,
			t1.Unix(), t2.Unix(), opts.AlbumID).Scan(&count)

	case opts.ArtistID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT l.listened_at, l.track_id, t.title
			FROM listens l
			JOIN tracks_with_title t ON l.track_id = t.id
			JOIN artist_tracks at2 ON t.id = at2.track_id
			WHERE l.listened_at BETWEEN ? AND ? AND at2.artist_id = ?
			ORDER BY l.listened_at DESC LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.ArtistID, opts.Limit, offset,
		)
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated (by artist): %w", err)
		}
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens l JOIN artist_tracks at2 ON l.track_id = at2.track_id
			WHERE l.listened_at BETWEEN ? AND ? AND at2.artist_id = ?`,
			t1.Unix(), t2.Unix(), opts.ArtistID).Scan(&count)

	default:
		rows, err = s.db.QueryContext(ctx, `
			SELECT l.listened_at, l.track_id, t.title
			FROM listens l
			JOIN tracks_with_title t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ?
			ORDER BY l.listened_at DESC LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.Limit, offset,
		)
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: %w", err)
		}
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM listens WHERE listened_at BETWEEN ? AND ?`,
			t1.Unix(), t2.Unix()).Scan(&count)
	}
	defer rows.Close()

	var listens []*models.Listen
	for rows.Next() {
		var l models.Listen
		var listenedAt int64
		if err := rows.Scan(&listenedAt, &l.Track.ID, &l.Track.Title); err != nil {
			return nil, err
		}
		l.Time = time.Unix(listenedAt, 0).UTC()
		artists, err := s.artistsForTrack(ctx, l.Track.ID)
		if err != nil {
			return nil, err
		}
		l.Track.Artists = artists
		listens = append(listens, &l)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if listens == nil {
		listens = []*models.Listen{}
	}
	return &db.PaginatedResponse[*models.Listen]{
		Items:        listens,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(listens)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}
