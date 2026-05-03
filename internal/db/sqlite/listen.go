package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
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

// listenRow is an intermediate scan target used to decouple the main rows
// query from the per-row artistsForTrack sub-query. With MaxOpenConns(1),
// both the count query and the artistsForTrack call would deadlock if
// executed while the outer *sql.Rows is still holding the only connection.
type listenRow struct {
	listenedAt int64
	trackID    int32
	title      string
}

func (s *Sqlite) GetListensPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[*models.Listen], error) {
	if opts.Limit == 0 {
		opts.Limit = defaultItemsPerPage
	}
	offset := (opts.Page - 1) * opts.Limit
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)

	// Count queries run first, before any main rows query is opened, so
	// they never compete with an open *sql.Rows for the single connection.
	var count int64
	switch {
	case opts.TrackID > 0:
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM listens WHERE listened_at BETWEEN ? AND ? AND track_id = ?`,
			t1.Unix(), t2.Unix(), opts.TrackID).Scan(&count)
	case opts.AlbumID > 0:
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens l JOIN tracks t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?`,
			t1.Unix(), t2.Unix(), opts.AlbumID).Scan(&count)
	case opts.ArtistID > 0:
		s.db.QueryRowContext(ctx, `
			SELECT COUNT(*) FROM listens l JOIN artist_tracks at2 ON l.track_id = at2.track_id
			WHERE l.listened_at BETWEEN ? AND ? AND at2.artist_id = ?`,
			t1.Unix(), t2.Unix(), opts.ArtistID).Scan(&count)
	default:
		s.db.QueryRowContext(ctx,
			`SELECT COUNT(*) FROM listens WHERE listened_at BETWEEN ? AND ?`,
			t1.Unix(), t2.Unix()).Scan(&count)
	}

	// Open and fully drain the main rows query, then close before any
	// sub-queries so the connection is free for artistsForTrack.
	var rows *sql.Rows
	var err error
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
	case opts.AlbumID > 0:
		rows, err = s.db.QueryContext(ctx, `
			SELECT l.listened_at, l.track_id, t.title
			FROM listens l
			JOIN tracks_with_title t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ? AND t.release_id = ?
			ORDER BY l.listened_at DESC LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.AlbumID, opts.Limit, offset,
		)
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
	default:
		rows, err = s.db.QueryContext(ctx, `
			SELECT l.listened_at, l.track_id, t.title
			FROM listens l
			JOIN tracks_with_title t ON l.track_id = t.id
			WHERE l.listened_at BETWEEN ? AND ?
			ORDER BY l.listened_at DESC LIMIT ? OFFSET ?`,
			t1.Unix(), t2.Unix(), opts.Limit, offset,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("GetListensPaginated: %w", err)
	}

	var raw []listenRow
	for rows.Next() {
		var r listenRow
		if err := rows.Scan(&r.listenedAt, &r.trackID, &r.title); err != nil {
			rows.Close()
			return nil, err
		}
		raw = append(raw, r)
	}
	// Explicitly close before calling artistsForTrack so the connection is free.
	if err := rows.Close(); err != nil {
		return nil, err
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	listens := make([]*models.Listen, 0, len(raw))
	for _, r := range raw {
		l := &models.Listen{
			Time: time.Unix(r.listenedAt, 0).UTC(),
			Track: models.Track{
				ID:    r.trackID,
				Title: r.title,
			},
		}
		l.Track.Artists, err = s.artistsForTrack(ctx, r.trackID)
		if err != nil {
			return nil, err
		}
		l.Track.Image, err = s.imageForTrack(ctx, l.Track.ID)
		if err != nil {
			return nil, err
		}
		listens = append(listens, l)
	}

	return &db.PaginatedResponse[*models.Listen]{
		Items:        listens,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(listens)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}

func (s *Sqlite) imageForTrack(ctx context.Context, trackId int32) (*uuid.UUID, error) {
	row := s.db.QueryRowContext(ctx, `
		SELECT r.image
		FROM releases r
		JOIN tracks t ON r.id = t.release_id
		WHERE t.id = ? AND r.image IS NOT NULL`,
		trackId,
	)
	var imageid *uuid.UUID
	err := row.Scan(&imageid)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("imageForTrack: %w", err)
	}
	return imageid, nil
}
