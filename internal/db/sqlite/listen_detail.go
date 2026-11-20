package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
)

// GetTrackDetail returns the track, its release and its artists with their MusicBrainz IDs.
func (s *Sqlite) GetTrackDetail(ctx context.Context, id int32) (*db.TrackDetail, error) {
	var d db.TrackDetail
	var releaseTitle sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT tt.id, tt.title, t.musicbrainz_id, t.duration, r.musicbrainz_id, rwt.title
		FROM tracks t
		JOIN tracks_with_title tt ON tt.id = t.id
		JOIN releases r ON r.id = t.release_id
		LEFT JOIN releases_with_title rwt ON rwt.id = r.id
		WHERE t.id = ?`,
		id,
	).Scan(&d.TrackID, &d.TrackTitle, &d.TrackMbid, &d.Duration, &d.ReleaseMbid, &releaseTitle)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("GetTrackDetail: %w", db.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("GetTrackDetail: %w", err)
	}
	d.ReleaseTitle = releaseTitle.String

	rows, err := s.db.QueryContext(ctx, `
		SELECT a.musicbrainz_id, awn.name
		FROM artist_tracks at2
		JOIN artists a ON a.id = at2.artist_id
		JOIN artists_with_name awn ON awn.id = a.id
		WHERE at2.track_id = ?
		ORDER BY at2.is_primary DESC, awn.name`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTrackDetail: artists: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var mbid *uuid.UUID
		var name string
		if err := rows.Scan(&mbid, &name); err != nil {
			return nil, fmt.Errorf("GetTrackDetail: artists: %w", err)
		}
		d.ArtistNames = append(d.ArtistNames, name)
		if mbid != nil {
			d.ArtistMbids = append(d.ArtistMbids, *mbid)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetTrackDetail: artists: %w", err)
	}
	return &d, nil
}

// GetListensInRange returns listens strictly after opts.From and strictly before opts.To
// (unix seconds; zero leaves a side unbounded), newest first.
func (s *Sqlite) GetListensInRange(ctx context.Context, opts db.ListenRangeOpts) ([]db.ListenDetail, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT listened_at, client, track_id
		FROM listens
		WHERE (? = 0 OR listened_at > ?) AND (? = 0 OR listened_at < ?)
		ORDER BY listened_at DESC
		LIMIT ?`,
		opts.From, opts.From, opts.To, opts.To, opts.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("GetListensInRange: %w", err)
	}

	type rawListen struct {
		listenedAt int64
		client     string
		trackID    int32
	}
	var raw []rawListen
	for rows.Next() {
		var r rawListen
		if err := rows.Scan(&r.listenedAt, &r.client, &r.trackID); err != nil {
			rows.Close()
			return nil, fmt.Errorf("GetListensInRange: %w", err)
		}
		raw = append(raw, r)
	}
	// close before issuing track lookups: the connection is shared (MaxOpenConns(1))
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("GetListensInRange: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetListensInRange: %w", err)
	}

	listens := make([]db.ListenDetail, 0, len(raw))
	for _, r := range raw {
		track, err := s.GetTrackDetail(ctx, r.trackID)
		if err != nil {
			return nil, fmt.Errorf("GetListensInRange: %w", err)
		}
		listens = append(listens, db.ListenDetail{
			ListenedAt: time.Unix(r.listenedAt, 0).UTC(),
			Client:     r.client,
			Track:      *track,
		})
	}
	return listens, nil
}

func (s *Sqlite) GetLatestListenUnix(ctx context.Context) (int64, error) {
	var unix sql.NullInt64
	err := s.db.QueryRowContext(ctx, `SELECT MAX(listened_at) FROM listens`).Scan(&unix)
	if err != nil {
		return 0, fmt.Errorf("GetLatestListenUnix: %w", err)
	}
	return unix.Int64, nil
}
