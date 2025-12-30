package psql

import (
	"context"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/gabehf/koito/internal/utils"
)

func (d *Psql) GetTopArtistsPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[*models.Artist], error) {
	l := logger.FromContext(ctx)
	offset := (opts.Page - 1) * opts.Limit
	t1, t2, err := utils.DateRange(opts.Week, opts.Month, opts.Year)
	if err != nil {
		return nil, fmt.Errorf("GetTopArtistsPaginated: %w", err)
	}
	if opts.Month == 0 && opts.Year == 0 {
		// use period, not date range
		t2 = time.Now()
		t1 = db.StartTimeFromPeriod(opts.Period)
	}
	if opts.From != 0 || opts.To != 0 {
		t1 = time.Unix(opts.From, 0)
		t2 = time.Unix(opts.To, 0)
	}
	if opts.Limit == 0 {
		opts.Limit = DefaultItemsPerPage
	}
	l.Debug().Msgf("Fetching top %d artists with period %s on page %d from range %v to %v",
		opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
	rows, err := d.q.GetTopArtistsPaginated(ctx, repository.GetTopArtistsPaginatedParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
		Limit:        int32(opts.Limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("GetTopArtistsPaginated: GetTopArtistsPaginated: %w", err)
	}
	rgs := make([]*models.Artist, len(rows))
	for i, row := range rows {
		t := &models.Artist{
			Name:        row.Name,
			MbzID:       row.MusicBrainzID,
			ID:          row.ID,
			Image:       row.Image,
			ListenCount: row.ListenCount,
		}
		rgs[i] = t
	}
	count, err := d.q.CountTopArtists(ctx, repository.CountTopArtistsParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return nil, fmt.Errorf("GetTopArtistsPaginated: CountTopArtists: %w", err)
	}
	l.Debug().Msgf("Database responded with %d artists out of a total %d", len(rows), count)

	return &db.PaginatedResponse[*models.Artist]{
		Items:        rgs,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(rgs)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}
