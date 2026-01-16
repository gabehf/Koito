package psql

import (
	"context"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
)

func (d *Psql) GetTopArtistsPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[db.RankedItem[*models.Artist]], error) {
	l := logger.FromContext(ctx)
	offset := (opts.Page - 1) * opts.Limit
	t1, t2 := db.TimeframeToTimeRange(opts.Timeframe)
	if opts.Limit == 0 {
		opts.Limit = DefaultItemsPerPage
	}
	l.Debug().Msgf("Fetching top %d artists on page %d from range %v to %v",
		opts.Limit, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
	rows, err := d.q.GetTopArtistsPaginated(ctx, repository.GetTopArtistsPaginatedParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
		Limit:        int32(opts.Limit),
		Offset:       int32(offset),
	})
	if err != nil {
		return nil, fmt.Errorf("GetTopArtistsPaginated: GetTopArtistsPaginated: %w", err)
	}
	rgs := make([]db.RankedItem[*models.Artist], len(rows))
	for i, row := range rows {
		t := &models.Artist{
			Name:        row.Name,
			MbzID:       row.MusicBrainzID,
			ID:          row.ID,
			Image:       row.Image,
			ListenCount: row.ListenCount,
		}
		rgs[i].Item = t
		rgs[i].Rank = row.Rank
	}
	count, err := d.q.CountTopArtists(ctx, repository.CountTopArtistsParams{
		ListenedAt:   t1,
		ListenedAt_2: t2,
	})
	if err != nil {
		return nil, fmt.Errorf("GetTopArtistsPaginated: CountTopArtists: %w", err)
	}
	l.Debug().Msgf("Database responded with %d artists out of a total %d", len(rows), count)

	return &db.PaginatedResponse[db.RankedItem[*models.Artist]]{
		Items:        rgs,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(rgs)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}
