package psql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

const searchItemLimit = 8
const substringSearchLength = 6

func (d *Psql) SearchArtists(ctx context.Context, q string) ([]*models.Artist, error) {
	if len(q) < substringSearchLength {
		rows, err := d.q.SearchArtistsBySubstring(ctx, repository.SearchArtistsBySubstringParams{
			Column1: pgtype.Text{String: q, Valid: true},
			Limit:   searchItemLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("SearchArtist: SearchArtistsBySubstring: %w", err)
		}
		ret := make([]*models.Artist, len(rows))
		for i, row := range rows {
			ret[i] = &models.Artist{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Name:  row.Name,
				Image: catalog.BuildImageList(row.Image),
			}
		}
		return ret, nil
	} else {
		rows, err := d.q.SearchArtists(ctx, repository.SearchArtistsParams{
			Similarity: q,
			Limit:      searchItemLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("SearchArtist: SearchArtists: %w", err)
		}
		ret := make([]*models.Artist, len(rows))
		for i, row := range rows {
			ret[i] = &models.Artist{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Name:  row.Name,
				Image: catalog.BuildImageList(row.Image),
			}
		}
		return ret, nil
	}
}

func (d *Psql) SearchAlbums(ctx context.Context, q string) ([]*models.Album, error) {
	if len(q) < substringSearchLength {
		rows, err := d.q.SearchReleasesBySubstring(ctx, repository.SearchReleasesBySubstringParams{
			Column1: pgtype.Text{String: q, Valid: true},
			Limit:   searchItemLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("SearchAlbums: SearchReleasesBySubstring: %w", err)
		}
		ret := make([]*models.Album, len(rows))
		for i, row := range rows {
			ret[i] = &models.Album{
				ID:             row.ID,
				MbzID:          row.MusicBrainzID,
				Title:          row.Title,
				VariousArtists: row.VariousArtists,
				Image:          catalog.BuildImageList(row.Image),
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, fmt.Errorf("SearchAlbums: Unmarshal: %w", err)
			}
		}
		return ret, nil
	} else {
		rows, err := d.q.SearchReleases(ctx, repository.SearchReleasesParams{
			Similarity: q,
			Limit:      searchItemLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("SearchAlbums: SearchReleases: %w", err)
		}
		ret := make([]*models.Album, len(rows))
		for i, row := range rows {
			ret[i] = &models.Album{
				ID:             row.ID,
				MbzID:          row.MusicBrainzID,
				Title:          row.Title,
				VariousArtists: row.VariousArtists,
				Image:          catalog.BuildImageList(row.Image),
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, fmt.Errorf("SearchAlbums: Unmarshal: %w", err)
			}
		}
		return ret, nil
	}
}

func (d *Psql) SearchTracks(ctx context.Context, q string) ([]*models.Track, error) {
	if len(q) < substringSearchLength {
		rows, err := d.q.SearchTracksBySubstring(ctx, repository.SearchTracksBySubstringParams{
			Column1: pgtype.Text{String: q, Valid: true},
			Limit:   searchItemLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("SearchTracks: SearchTracksBySubstring: %w", err)
		}
		ret := make([]*models.Track, len(rows))
		for i, row := range rows {
			ret[i] = &models.Track{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Title: row.Title,
				Image: row.Image,
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, fmt.Errorf("SearchTracks: Unmarshal: %w", err)
			}
		}
		return ret, nil
	} else {
		rows, err := d.q.SearchTracks(ctx, repository.SearchTracksParams{
			Similarity: q,
			Limit:      searchItemLimit,
		})
		if err != nil {
			return nil, fmt.Errorf("SearchTracks: SearchTracks: %w", err)
		}
		ret := make([]*models.Track, len(rows))
		for i, row := range rows {
			ret[i] = &models.Track{
				ID:    row.ID,
				MbzID: row.MusicBrainzID,
				Title: row.Title,
				Image: row.Image,
			}
			err = json.Unmarshal(row.Artists, &ret[i].Artists)
			if err != nil {
				return nil, fmt.Errorf("SearchTracks: Unmarshal: %w", err)
			}
		}
		return ret, nil
	}
}
