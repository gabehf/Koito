package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
)

func (s *Sqlite) SaveTrackFeedback(ctx context.Context, opts db.SaveTrackFeedbackOpts) error {
	switch opts.Score {
	case 0:
		_, err := s.db.ExecContext(ctx,
			`DELETE FROM track_feedback WHERE track_id = ? AND user_id = ?`,
			opts.TrackID, opts.UserID,
		)
		if err != nil {
			return fmt.Errorf("SaveTrackFeedback: %w", err)
		}
		return nil
	case 1, -1:
		_, err := s.db.ExecContext(ctx, `
			INSERT INTO track_feedback (track_id, user_id, score, created_at) VALUES (?, ?, ?, ?)
			ON CONFLICT (track_id, user_id) DO UPDATE SET score = excluded.score, created_at = excluded.created_at`,
			opts.TrackID, opts.UserID, opts.Score, time.Now().Unix(),
		)
		if err != nil {
			return fmt.Errorf("SaveTrackFeedback: %w", err)
		}
		return nil
	default:
		return errors.New("SaveTrackFeedback: score must be -1, 0 or 1")
	}
}

// GetTrackFeedback returns a user's feedback on tracks, most recent first, and the total number of entries.
func (s *Sqlite) GetTrackFeedback(ctx context.Context, userId int32, limit, offset int) ([]db.TrackFeedback, int64, error) {
	var total int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM track_feedback WHERE user_id = ?`, userId,
	).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetTrackFeedback: %w", err)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT track_id, score, created_at
		FROM track_feedback
		WHERE user_id = ?
		ORDER BY created_at DESC, track_id DESC
		LIMIT ? OFFSET ?`,
		userId, limit, offset,
	)
	if err != nil {
		return nil, 0, fmt.Errorf("GetTrackFeedback: %w", err)
	}

	type rawFeedback struct {
		trackID   int32
		score     int
		createdAt int64
	}
	var raw []rawFeedback
	for rows.Next() {
		var r rawFeedback
		if err := rows.Scan(&r.trackID, &r.score, &r.createdAt); err != nil {
			rows.Close()
			return nil, 0, fmt.Errorf("GetTrackFeedback: %w", err)
		}
		raw = append(raw, r)
	}
	// close before issuing track lookups: the connection is shared (MaxOpenConns(1))
	if err := rows.Close(); err != nil {
		return nil, 0, fmt.Errorf("GetTrackFeedback: %w", err)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("GetTrackFeedback: %w", err)
	}

	items := make([]db.TrackFeedback, 0, len(raw))
	for _, r := range raw {
		track, err := s.GetTrackDetail(ctx, r.trackID)
		if err != nil {
			return nil, 0, fmt.Errorf("GetTrackFeedback: %w", err)
		}
		items = append(items, db.TrackFeedback{
			CreatedAt: time.Unix(r.createdAt, 0).UTC(),
			Score:     r.score,
			Track:     *track,
		})
	}
	return items, total, nil
}
