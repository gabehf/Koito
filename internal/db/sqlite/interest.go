package sqlite

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
)

// GetInterest computes N evenly-spaced listen-count buckets spanning the entity's
// entire listen history. The PG implementation used generate_series + EXTRACT(EPOCH);
// here we fetch all timestamps in Go and do the bucketing in memory.
func (s *Sqlite) GetInterest(ctx context.Context, opts db.GetInterestOpts) ([]db.InterestBucket, error) {
	if opts.Buckets == 0 {
		return nil, errors.New("GetInterest: bucket count must be provided")
	}

	var query string
	var arg any
	switch {
	case opts.ArtistID != 0:
		query = `
			SELECT l.listened_at FROM listens l
			JOIN tracks t ON t.id = l.track_id
			JOIN artist_tracks at2 ON at2.track_id = t.id
			WHERE at2.artist_id = ?
			ORDER BY l.listened_at`
		arg = opts.ArtistID
	case opts.AlbumID != 0:
		query = `
			SELECT l.listened_at FROM listens l
			JOIN tracks t ON t.id = l.track_id
			WHERE t.release_id = ?
			ORDER BY l.listened_at`
		arg = opts.AlbumID
	case opts.TrackID != 0:
		query = `
			SELECT listened_at FROM listens WHERE track_id = ? ORDER BY listened_at`
		arg = opts.TrackID
	default:
		return nil, errors.New("GetInterest: artist id, album id, or track id must be provided")
	}

	rows, err := s.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, fmt.Errorf("GetInterest: query: %w", err)
	}
	defer rows.Close()

	var timestamps []int64
	for rows.Next() {
		var t int64
		if err := rows.Scan(&t); err != nil {
			return nil, err
		}
		timestamps = append(timestamps, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if len(timestamps) == 0 {
		// return empty buckets spanning now to now
		now := time.Now().Unix()
		buckets := make([]db.InterestBucket, opts.Buckets)
		for i := range buckets {
			buckets[i] = db.InterestBucket{
				BucketStart: time.Unix(now, 0).UTC(),
				BucketEnd:   time.Unix(now, 0).UTC(),
				ListenCount: 0,
			}
		}
		return buckets, nil
	}

	start := timestamps[0]
	end := time.Now().Unix()
	totalSeconds := end - start
	n := int64(opts.Buckets)

	// bucket duration in seconds (may be 0 if all listens at the same instant)
	bucketSize := totalSeconds / n
	if bucketSize == 0 {
		bucketSize = 1
	}

	counts := make([]int64, opts.Buckets)
	for _, ts := range timestamps {
		idx := (ts - start) * n / (totalSeconds + 1)
		if idx >= n {
			idx = n - 1
		}
		counts[idx]++
	}

	buckets := make([]db.InterestBucket, opts.Buckets)
	for i := range buckets {
		bucketStart := start + int64(i)*bucketSize
		bucketEnd := bucketStart + bucketSize
		buckets[i] = db.InterestBucket{
			BucketStart: time.Unix(bucketStart, 0).UTC(),
			BucketEnd:   time.Unix(bucketEnd, 0).UTC(),
			ListenCount: counts[i],
		}
	}
	return buckets, nil
}
