package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/gabehf/koito/internal/utils"
)

func (d *Psql) GetListensPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[*models.Listen], error) {
	l := logger.FromContext(ctx)
	offset := (opts.Page - 1) * opts.Limit
	var t1 time.Time
	var t2 time.Time
	if opts.From != 0 && opts.To != 0 {
		t1 = time.Unix(int64(opts.From), 0)
		t2 = time.Unix(int64(opts.To), 0)
	} else {
		t1R, t2R, err := utils.DateRange(opts.Week, opts.Month, opts.Year)
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: %w", err)
		}
		t1 = t1R
		t2 = t2R
		if opts.Month == 0 && opts.Year == 0 {
			// use period, not date range
			t2 = time.Now()
			t1 = db.StartTimeFromPeriod(opts.Period)
		}
	}
	if opts.Limit == 0 {
		opts.Limit = DefaultItemsPerPage
	}
	var listens []*models.Listen
	var count int64
	if opts.TrackID > 0 {
		l.Debug().Msgf("Fetching %d listens with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetLastListensFromTrackPaginated(ctx, repository.GetLastListensFromTrackPaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
			ID:           int32(opts.TrackID),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: GetLastListensFromTrackPaginated: %w", err)
		}
		listens = make([]*models.Listen, len(rows))
		for i, row := range rows {
			t := &models.Listen{
				Track: models.Track{
					Title: row.TrackTitle,
					ID:    row.TrackID,
				},
				Time: row.ListenedAt,
			}
			err = json.Unmarshal(row.Artists, &t.Track.Artists)
			if err != nil {
				return nil, fmt.Errorf("GetListensPaginated: Unmarshal: %w", err)
			}
			listens[i] = t
		}
		count, err = d.q.CountListensFromTrack(ctx, repository.CountListensFromTrackParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			TrackID:      int32(opts.TrackID),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: CountListensFromTrack: %w", err)
		}
	} else if opts.AlbumID > 0 {
		l.Debug().Msgf("Fetching %d listens with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetLastListensFromReleasePaginated(ctx, repository.GetLastListensFromReleasePaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
			ReleaseID:    int32(opts.AlbumID),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: GetLastListensFromReleasePaginated: %w", err)
		}
		listens = make([]*models.Listen, len(rows))
		for i, row := range rows {
			t := &models.Listen{
				Track: models.Track{
					Title: row.TrackTitle,
					ID:    row.TrackID,
				},
				Time: row.ListenedAt,
			}
			err = json.Unmarshal(row.Artists, &t.Track.Artists)
			if err != nil {
				return nil, fmt.Errorf("GetListensPaginated: Unmarshal: %w", err)
			}
			listens[i] = t
		}
		count, err = d.q.CountListensFromRelease(ctx, repository.CountListensFromReleaseParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ReleaseID:    int32(opts.AlbumID),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: CountListensFromRelease: %w", err)
		}
	} else if opts.ArtistID > 0 {
		l.Debug().Msgf("Fetching %d listens with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetLastListensFromArtistPaginated(ctx, repository.GetLastListensFromArtistPaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
			ArtistID:     int32(opts.ArtistID),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: GetLastListensFromArtistPaginated: %w", err)
		}
		listens = make([]*models.Listen, len(rows))
		for i, row := range rows {
			t := &models.Listen{
				Track: models.Track{
					Title: row.TrackTitle,
					ID:    row.TrackID,
				},
				Time: row.ListenedAt,
			}
			err = json.Unmarshal(row.Artists, &t.Track.Artists)
			if err != nil {
				return nil, fmt.Errorf("GetListensPaginated: Unmarshal: %w", err)
			}
			listens[i] = t
		}
		count, err = d.q.CountListensFromArtist(ctx, repository.CountListensFromArtistParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			ArtistID:     int32(opts.ArtistID),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: CountListensFromArtist: %w", err)
		}
	} else {
		l.Debug().Msgf("Fetching %d listens with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetLastListensPaginated(ctx, repository.GetLastListensPaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: GetLastListensPaginated: %w", err)
		}
		listens = make([]*models.Listen, len(rows))
		for i, row := range rows {
			t := &models.Listen{
				Track: models.Track{
					Title: row.TrackTitle,
					ID:    row.TrackID,
				},
				Time: row.ListenedAt,
			}
			err = json.Unmarshal(row.Artists, &t.Track.Artists)
			if err != nil {
				return nil, fmt.Errorf("GetListensPaginated: Unmarshal: %w", err)
			}
			listens[i] = t
		}
		count, err = d.q.CountListens(ctx, repository.CountListensParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
		})
		if err != nil {
			return nil, fmt.Errorf("GetListensPaginated: CountListens: %w", err)
		}
		l.Debug().Msgf("Database responded with %d tracks out of a total %d", len(rows), count)
	}

	return &db.PaginatedResponse[*models.Listen]{
		Items:        listens,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(listens)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}

func (d *Psql) SaveListen(ctx context.Context, opts db.SaveListenOpts) error {
	l := logger.FromContext(ctx)
	if opts.TrackID == 0 {
		return errors.New("required parameter TrackID missing")
	}
	if opts.Time.IsZero() {
		opts.Time = time.Now()
	}
	var client *string
	if opts.Client != "" {
		client = &opts.Client
	}
	l.Debug().Msgf("Inserting listen for track with id %d at time %v into DB", opts.TrackID, opts.Time)
	return d.q.InsertListen(ctx, repository.InsertListenParams{
		TrackID:    opts.TrackID,
		ListenedAt: opts.Time,
		UserID:     opts.UserID,
		Client:     client,
	})
}

func (d *Psql) DeleteListen(ctx context.Context, trackId int32, listenedAt time.Time) error {
	l := logger.FromContext(ctx)
	if trackId == 0 {
		return errors.New("required parameter 'trackId' missing")
	}
	l.Debug().Msgf("Deleting listen from track %d at time %s from DB", trackId, listenedAt)
	return d.q.DeleteListen(ctx, repository.DeleteListenParams{
		TrackID:    trackId,
		ListenedAt: listenedAt,
	})
}
