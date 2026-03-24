package psql

import (
	"context"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5"
)

func (d *Psql) GetTrackLookup(ctx context.Context, key string) (*db.TrackLookupResult, error) {
	row, err := d.q.GetTrackLookup(ctx, key)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, err
		}
		return nil, err
	}
	return &db.TrackLookupResult{
		ArtistID: row.ArtistID,
		AlbumID:  row.AlbumID,
		TrackID:  row.TrackID,
	}, nil
}

func (d *Psql) SaveTrackLookup(ctx context.Context, opts db.SaveTrackLookupOpts) error {
	return d.q.InsertTrackLookup(ctx, repository.InsertTrackLookupParams{
		LookupKey: opts.Key,
		ArtistID:  opts.ArtistID,
		AlbumID:   opts.AlbumID,
		TrackID:   opts.TrackID,
	})
}

func (d *Psql) InvalidateTrackLookup(ctx context.Context, opts db.InvalidateTrackLookupOpts) error {
	if opts.ArtistID != 0 {
		if err := d.q.DeleteTrackLookupByArtist(ctx, opts.ArtistID); err != nil {
			return err
		}
	}
	if opts.AlbumID != 0 {
		if err := d.q.DeleteTrackLookupByAlbum(ctx, opts.AlbumID); err != nil {
			return err
		}
	}
	if opts.TrackID != 0 {
		if err := d.q.DeleteTrackLookupByTrack(ctx, opts.TrackID); err != nil {
			return err
		}
	}
	return nil
}
