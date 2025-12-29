package psql

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/repository"
)

func (p *Psql) CountListens(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	var t1, t2 time.Time
	if timeframe.T1u == 0 && timeframe.T2u == 0 {
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(timeframe.Period)
	} else {
		t1 = time.Unix(timeframe.T1u, 0)
		t2 = time.Unix(timeframe.T2u, 0)
	}
	count, err := p.q.CountListens(ctx, repository.CountListensParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return 0, fmt.Errorf("CountListens: %w", err)
	}
	return count, nil
}

func (p *Psql) CountTracks(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	var t1, t2 time.Time
	if timeframe.T1u == 0 && timeframe.T2u == 0 {
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(timeframe.Period)
	} else {
		t1 = time.Unix(timeframe.T1u, 0)
		t2 = time.Unix(timeframe.T2u, 0)
	}
	count, err := p.q.CountTopTracks(ctx, repository.CountTopTracksParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return 0, fmt.Errorf("CountTracks: %w", err)
	}
	return count, nil
}

func (p *Psql) CountAlbums(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	var t1, t2 time.Time
	if timeframe.T1u == 0 && timeframe.T2u == 0 {
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(timeframe.Period)
	} else {
		t1 = time.Unix(timeframe.T1u, 0)
		t2 = time.Unix(timeframe.T2u, 0)
	}
	count, err := p.q.CountTopReleases(ctx, repository.CountTopReleasesParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return 0, fmt.Errorf("CountAlbums: %w", err)
	}
	return count, nil
}

func (p *Psql) CountArtists(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	var t1, t2 time.Time
	if timeframe.T1u == 0 && timeframe.T2u == 0 {
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(timeframe.Period)
	} else {
		t1 = time.Unix(timeframe.T1u, 0)
		t2 = time.Unix(timeframe.T2u, 0)
	}
	count, err := p.q.CountTopArtists(ctx, repository.CountTopArtistsParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return 0, fmt.Errorf("CountArtists: %w", err)
	}
	return count, nil
}

func (p *Psql) CountTimeListened(ctx context.Context, timeframe db.Timeframe) (int64, error) {
	var t1, t2 time.Time
	if timeframe.T1u == 0 && timeframe.T2u == 0 {
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(timeframe.Period)
	} else {
		t1 = time.Unix(timeframe.T1u, 0)
		t2 = time.Unix(timeframe.T2u, 0)
	}
	count, err := p.q.CountTimeListened(ctx, repository.CountTimeListenedParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return 0, fmt.Errorf("CountTimeListened: %w", err)
	}
	return count, nil
}

func (p *Psql) CountTimeListenedToItem(ctx context.Context, opts db.TimeListenedOpts) (int64, error) {
	t2 := time.Now()
	t1 := db.StartTimeFromPeriod(opts.Period)

	if opts.ArtistID > 0 {
		count, err := p.q.CountTimeListenedToArtist(ctx, repository.CountTimeListenedToArtistParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ArtistID:     opts.ArtistID,
		})
		if err != nil {
			return 0, fmt.Errorf("CountTimeListenedToItem (Artist): %w", err)
		}
		return count, nil
	} else if opts.AlbumID > 0 {
		count, err := p.q.CountTimeListenedToRelease(ctx, repository.CountTimeListenedToReleaseParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ReleaseID:    opts.AlbumID,
		})
		if err != nil {
			return 0, fmt.Errorf("CountTimeListenedToItem (Album): %w", err)
		}
		return count, nil
	} else if opts.TrackID > 0 {
		count, err := p.q.CountTimeListenedToTrack(ctx, repository.CountTimeListenedToTrackParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ID:           opts.TrackID,
		})
		if err != nil {
			return 0, fmt.Errorf("CountTimeListenedToItem (Track): %w", err)
		}
		return count, nil
	}
	return 0, errors.New("CountTimeListenedToItem: an id must be provided")
}
