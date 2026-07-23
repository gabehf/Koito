package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/google/uuid"
)

func (s *Sqlite) GetTrackFeedback(ctx context.Context, userID, trackID int32) (int, error) {
	var score int
	err := s.db.QueryRowContext(ctx,
		`SELECT score FROM recording_feedback WHERE user_id = ? AND track_id = ?`,
		userID, trackID,
	).Scan(&score)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	}
	if err != nil {
		return 0, fmt.Errorf("GetTrackFeedback: %w", err)
	}
	return score, nil
}

func (s *Sqlite) SetTrackFeedback(ctx context.Context, userID, trackID int32, score int) error {
	if score != 1 && score != -1 && score != 0 {
		return fmt.Errorf("SetTrackFeedback: invalid score %d", score)
	}
	if score == 0 {
		_, err := s.db.ExecContext(ctx,
			`DELETE FROM recording_feedback WHERE user_id = ? AND track_id = ?`, userID, trackID)
		return err
	}

	now := time.Now().Unix()
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO recording_feedback (user_id, track_id, score, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?)
		ON CONFLICT(user_id, track_id) DO UPDATE SET
			score = excluded.score,
			updated_at = excluded.updated_at`,
		userID, trackID, score, now, now,
	)
	if err != nil {
		return fmt.Errorf("SetTrackFeedback: %w", err)
	}
	return nil
}
func (s *Sqlite) ListUserFeedback(ctx context.Context, opts db.ListUserFeedbackOpts) ([]db.RecordingFeedback, int64, error) {
	where := `WHERE rf.user_id = ?`
	args := []any{opts.UserID}
	if opts.Score != nil {
		where += ` AND rf.score = ?`
		args = append(args, *opts.Score)
	}

	var total int64
	if err := s.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM recording_feedback rf `+where, args...,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	args = append(args, opts.Limit, opts.Offset)
	rows, err := s.db.QueryContext(ctx, `
		SELECT t.musicbrainz_id, rf.score, rf.created_at
		FROM recording_feedback rf
		JOIN tracks t ON t.id = rf.track_id
		`+where+`
		ORDER BY rf.created_at DESC
		LIMIT ? OFFSET ?`, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	items := []db.RecordingFeedback{}
	for rows.Next() {
		var mbid sql.NullString
		var item db.RecordingFeedback
		if err := rows.Scan(&mbid, &item.Score, &item.CreatedAt); err != nil {
			return nil, 0, err
		}
		item.RecordingMBID = parseNullableUUID(mbid)
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (s *Sqlite) GetFeedbackForRecordings(ctx context.Context, opts db.GetFeedbackForRecordingsOpts) ([]db.RecordingFeedback, error) {
	mbids := opts.MBIDs
	out := make([]db.RecordingFeedback, len(mbids))
	for i := range mbids {
		out[i].RecordingMBID = &mbids[i]
	}
	if len(mbids) == 0 {
		return out, nil
	}

	placeholders := make([]string, len(mbids))
	args := make([]any, 0, len(mbids)+1)
	args = append(args, opts.UserID)
	for i, mbid := range mbids {
		placeholders[i] = "?"
		args = append(args, mbid.String())
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT t.musicbrainz_id, COALESCE(rf.score, 0), COALESCE(rf.created_at, 0)
		FROM tracks t
		LEFT JOIN recording_feedback rf ON rf.track_id = t.id AND rf.user_id = ?
		WHERE t.musicbrainz_id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	found := make(map[uuid.UUID]db.RecordingFeedback, len(mbids))
	for rows.Next() {
		var mbidStr string
		var item db.RecordingFeedback
		if err := rows.Scan(&mbidStr, &item.Score, &item.CreatedAt); err != nil {
			return nil, err
		}
		mbid, err := uuid.Parse(mbidStr)
		if err != nil {
			continue
		}
		item.RecordingMBID = &mbid
		found[mbid] = item
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i, mbid := range mbids {
		if item, ok := found[mbid]; ok {
			out[i] = item
			out[i].RecordingMBID = &mbids[i]
		}
	}
	return out, nil
}
