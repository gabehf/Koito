package psql

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/gabehf/koito/internal/utils"
)

func (d *Psql) GetTopAlbumsPaginated(ctx context.Context, opts db.GetItemsOpts) (*db.PaginatedResponse[*models.Album], error) {
	l := logger.FromContext(ctx)
	offset := (opts.Page - 1) * opts.Limit
	t1, t2, err := utils.DateRange(opts.Week, opts.Month, opts.Year)
	if err != nil {
		return nil, fmt.Errorf("GetTopAlbumsPaginated: %w", err)
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

	var rgs []*models.Album
	var count int64

	if opts.ArtistID != 0 {
		l.Debug().Msgf("Fetching top %d albums from artist id %d with period %s on page %d from range %v to %v",
			opts.Limit, opts.ArtistID, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))

		rows, err := d.q.GetTopReleasesFromArtist(ctx, repository.GetTopReleasesFromArtistParams{
			ArtistID:     int32(opts.ArtistID),
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
			ListenedAt:   t1,
			ListenedAt_2: t2,
		})
		if err != nil {
			return nil, fmt.Errorf("GetTopAlbumsPaginated: GetTopReleasesFromArtist: %w", err)
		}
		rgs = make([]*models.Album, len(rows))
		l.Debug().Msgf("Database responded with %d items", len(rows))
		for i, v := range rows {
			artists := make([]models.SimpleArtist, 0)
			err = json.Unmarshal(v.Artists, &artists)
			if err != nil {
				l.Err(err).Msgf("Error unmarshalling artists for release group with id %d", v.ID)
				return nil, fmt.Errorf("GetTopAlbumsPaginated: Unmarshal: %w", err)
			}
			rgs[i] = &models.Album{
				ID:             v.ID,
				MbzID:          v.MusicBrainzID,
				Title:          v.Title,
				Image:          v.Image,
				Artists:        artists,
				VariousArtists: v.VariousArtists,
				ListenCount:    v.ListenCount,
			}
		}
		count, err = d.q.CountReleasesFromArtist(ctx, int32(opts.ArtistID))
		if err != nil {
			return nil, fmt.Errorf("GetTopAlbumsPaginated: CountReleasesFromArtist: %w", err)
		}
	} else {
		l.Debug().Msgf("Fetching top %d albums with period %s on page %d from range %v to %v",
			opts.Limit, opts.Period, opts.Page, t1.Format("Jan 02, 2006"), t2.Format("Jan 02, 2006"))
		rows, err := d.q.GetTopReleasesPaginated(ctx, repository.GetTopReleasesPaginatedParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
			Limit:        int32(opts.Limit),
			Offset:       int32(offset),
		})
		if err != nil {
			return nil, fmt.Errorf("GetTopAlbumsPaginated: GetTopReleasesPaginated: %w", err)
		}
		rgs = make([]*models.Album, len(rows))
		l.Debug().Msgf("Database responded with %d items", len(rows))
		for i, row := range rows {
			artists := make([]models.SimpleArtist, 0)
			err = json.Unmarshal(row.Artists, &artists)
			if err != nil {
				l.Err(err).Msgf("Error unmarshalling artists for release group with id %d", row.ID)
				return nil, fmt.Errorf("GetTopAlbumsPaginated: Unmarshal: %w", err)
			}
			t := &models.Album{
				Title:          row.Title,
				MbzID:          row.MusicBrainzID,
				ID:             row.ID,
				Image:          row.Image,
				Artists:        artists,
				VariousArtists: row.VariousArtists,
				ListenCount:    row.ListenCount,
			}
			rgs[i] = t
		}
		count, err = d.q.CountTopReleases(ctx, repository.CountTopReleasesParams{
			ListenedAt:   t1,
			ListenedAt_2: t2,
		})
		if err != nil {
			return nil, fmt.Errorf("GetTopAlbumsPaginated: CountTopReleases: %w", err)
		}
		l.Debug().Msgf("Database responded with %d albums out of a total %d", len(rows), count)
	}
	return &db.PaginatedResponse[*models.Album]{
		Items:        rgs,
		TotalCount:   count,
		ItemsPerPage: int32(opts.Limit),
		HasNextPage:  int64(offset+len(rgs)) < count,
		CurrentPage:  int32(opts.Page),
	}, nil
}
