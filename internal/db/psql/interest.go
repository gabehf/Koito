package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/repository"
)

func (d *Psql) GetInterest(ctx context.Context, opts db.GetInterestOpts) ([]db.InterestBucket, error) {
	if opts.Buckets == 0 {
		return nil, errors.New("GetInterest: bucket count must be provided")
	}

	ret := make([]db.InterestBucket, opts.Buckets)

	if opts.ArtistID != 0 {
		resp, err := d.q.GetGroupedListensFromArtist(ctx, repository.GetGroupedListensFromArtistParams{
			ArtistID:    opts.ArtistID,
			BucketCount: opts.Buckets,
		})
		if err != nil {
			return nil, fmt.Errorf("GetInterest: GetGroupedListensFromArtist: %w", err)
		}
		for i, v := range resp {
			ret[i] = db.InterestBucket{
				BucketStart: v.BucketStart,
				BucketEnd:   v.BucketEnd,
				ListenCount: v.ListenCount,
			}
		}
		return ret, nil
	} else if opts.AlbumID != 0 {
		resp, err := d.q.GetGroupedListensFromRelease(ctx, repository.GetGroupedListensFromReleaseParams{
			ReleaseID:   opts.AlbumID,
			BucketCount: opts.Buckets,
		})
		if err != nil {
			return nil, fmt.Errorf("GetInterest: GetGroupedListensFromRelease: %w", err)
		}
		for i, v := range resp {
			ret[i] = db.InterestBucket{
				BucketStart: v.BucketStart,
				BucketEnd:   v.BucketEnd,
				ListenCount: v.ListenCount,
			}
		}
		return ret, nil
	} else if opts.TrackID != 0 {
		resp, err := d.q.GetGroupedListensFromTrack(ctx, repository.GetGroupedListensFromTrackParams{
			ID:          opts.TrackID,
			BucketCount: opts.Buckets,
		})
		if err != nil {
			return nil, fmt.Errorf("GetInterest: GetGroupedListensFromTrack: %w", err)
		}
		for i, v := range resp {
			ret[i] = db.InterestBucket{
				BucketStart: v.BucketStart,
				BucketEnd:   v.BucketEnd,
				ListenCount: v.ListenCount,
			}
		}
		return ret, nil
	} else {
		return nil, errors.New("GetInterest: artist id, album id, or track id must be provided")
	}
}
