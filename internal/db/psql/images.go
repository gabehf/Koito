package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (d *Psql) ImageHasAssociation(ctx context.Context, image uuid.UUID) (bool, error) {
	_, err := d.q.GetReleaseByImageID(ctx, &image)
	if err == nil {
		return true, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("ImageHasAssociation: GetReleaseByImageID: %w", err)
	}
	_, err = d.q.GetArtistByImage(ctx, &image)
	if err == nil {
		return true, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return false, fmt.Errorf("ImageHasAssociation: GetArtistByImage: %w", err)
	}
	return false, nil
}

func (d *Psql) GetImageSource(ctx context.Context, image uuid.UUID) (string, error) {
	r, err := d.q.GetReleaseByImageID(ctx, &image)
	if err == nil {
		return r.ImageSource.String, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("GetImageSource: GetReleaseByImageID: %w", err)
	}
	rr, err := d.q.GetArtistByImage(ctx, &image)
	if err == nil {
		return rr.ImageSource.String, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return "", fmt.Errorf("GetImageSource: GetArtistByImage: %w", err)
	}
	return "", nil
}

func (d *Psql) AlbumsWithoutImages(ctx context.Context, from int32) ([]*models.Album, error) {
	l := logger.FromContext(ctx)
	rows, err := d.q.GetReleasesWithoutImages(ctx, repository.GetReleasesWithoutImagesParams{
		Limit: 20,
		ID:    from,
	})
	if err != nil {
		return nil, fmt.Errorf("AlbumsWithoutImages: GetReleasesWithoutImages: %w", err)
	}
	albums := make([]*models.Album, len(rows))
	for i, row := range rows {
		var artists []models.SimpleArtist
		if err := json.Unmarshal(row.Artists, &artists); err != nil {
			l.Err(err).Msgf("AlbumsWithoutImages: error unmarshalling artists for release group with id %d", row.ID)
			artists = nil
		}
		albums[i] = &models.Album{
			ID:             row.ID,
			Image:          row.Image,
			Title:          row.Title,
			MbzID:          row.MusicBrainzID,
			VariousArtists: row.VariousArtists,
			Artists:        artists,
		}
	}
	return albums, nil
}

// returns nil, nil on no results
func (d *Psql) ArtistsWithoutImages(ctx context.Context, from int32) ([]*models.Artist, error) {
	rows, err := d.q.GetArtistsWithoutImages(ctx, repository.GetArtistsWithoutImagesParams{
		Limit: 20,
		ID:    from,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("ArtistsWithoutImages: %w", err)
	}

	ret := make([]*models.Artist, len(rows))
	for i, row := range rows {
		ret[i] = &models.Artist{
			ID:    row.ID,
			Name:  row.Name,
			MbzID: row.MusicBrainzID,
		}
	}
	return ret, nil
}
