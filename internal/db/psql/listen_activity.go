package psql

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

func (d *Psql) GetListenActivity(ctx context.Context, opts db.ListenActivityOpts) ([]db.ListenActivityItem, error) {
	l := logger.FromContext(ctx)
	if opts.Month != 0 && opts.Year == 0 {
		return nil, errors.New("year must be specified with month")
	}
	// Default to range = 12 if not set
	if opts.Range == 0 {
		opts.Range = db.DefaultRange
	}
	t1, t2 := db.ListenActivityOptsToTimes(opts)
	var listenActivity []db.ListenActivityItem
	if opts.AlbumID > 0 {
		l.Debug().Msgf("Fetching listen activity for %d %s(s) from %v to %v for release group %d",
			opts.Range, opts.Step, t1.Format("Jan 02, 2006 15:04:05"), t2.Format("Jan 02, 2006 15:04:05"), opts.AlbumID)
		rows, err := d.q.ListenActivityForRelease(ctx, repository.ListenActivityForReleaseParams{
			Column1:   t1,
			Column2:   t2,
			Column3:   stepToInterval(opts.Step),
			ReleaseID: opts.AlbumID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetListenActivity: ListenActivityForRelease: %w", err)
		}
		listenActivity = make([]db.ListenActivityItem, len(rows))
		for i, row := range rows {
			t := db.ListenActivityItem{
				Start:   row.BucketStart,
				Listens: row.ListenCount,
			}
			listenActivity[i] = t
		}
		l.Debug().Msgf("Database responded with %d steps", len(rows))
	} else if opts.ArtistID > 0 {
		l.Debug().Msgf("Fetching listen activity for %d %s(s) from %v to %v for artist %d",
			opts.Range, opts.Step, t1.Format("Jan 02, 2006 15:04:05"), t2.Format("Jan 02, 2006 15:04:05"), opts.ArtistID)
		rows, err := d.q.ListenActivityForArtist(ctx, repository.ListenActivityForArtistParams{
			Column1:  t1,
			Column2:  t2,
			Column3:  stepToInterval(opts.Step),
			ArtistID: opts.ArtistID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetListenActivity: ListenActivityForArtist: %w", err)
		}
		listenActivity = make([]db.ListenActivityItem, len(rows))
		for i, row := range rows {
			t := db.ListenActivityItem{
				Start:   row.BucketStart,
				Listens: row.ListenCount,
			}
			listenActivity[i] = t
		}
		l.Debug().Msgf("Database responded with %d steps", len(rows))
	} else if opts.TrackID > 0 {
		l.Debug().Msgf("Fetching listen activity for %d %s(s) from %v to %v for track %d",
			opts.Range, opts.Step, t1.Format("Jan 02, 2006 15:04:05"), t2.Format("Jan 02, 2006 15:04:05"), opts.TrackID)
		rows, err := d.q.ListenActivityForTrack(ctx, repository.ListenActivityForTrackParams{
			Column1: t1,
			Column2: t2,
			Column3: stepToInterval(opts.Step),
			ID:      opts.TrackID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetListenActivity: ListenActivityForTrack: %w", err)
		}
		listenActivity = make([]db.ListenActivityItem, len(rows))
		for i, row := range rows {
			t := db.ListenActivityItem{
				Start:   row.BucketStart,
				Listens: row.ListenCount,
			}
			listenActivity[i] = t
		}
		l.Debug().Msgf("Database responded with %d steps", len(rows))
	} else {
		l.Debug().Msgf("Fetching listen activity for %d %s(s) from %v to %v",
			opts.Range, opts.Step, t1.Format("Jan 02, 2006 15:04:05"), t2.Format("Jan 02, 2006 15:04:05"))
		rows, err := d.q.ListenActivity(ctx, repository.ListenActivityParams{
			Column1: pgtype.Date{Time: t1, Valid: true},
			Column2: pgtype.Date{Time: t2, Valid: true},
			Column3: stepToInterval(opts.Step),
		})
		if err != nil {
			return nil, fmt.Errorf("GetListenActivity: ListenActivity: %w", err)
		}
		listenActivity = make([]db.ListenActivityItem, len(rows))
		for i, row := range rows {
			t := db.ListenActivityItem{
				Start:   row.BucketedListensBucketStartDate,
				Listens: row.ListenCount,
			}
			listenActivity[i] = t
		}
		l.Debug().Msgf("Database responded with %d steps", len(rows))
	}

	return listenActivity, nil
}
