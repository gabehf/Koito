package psql

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
)

func (d *Psql) GetExportPage(ctx context.Context, opts db.GetExportPageOpts) ([]*db.ExportItem, error) {
	rows, err := d.q.GetListensExportPage(ctx, repository.GetListensExportPageParams{
		UserID:     opts.UserID,
		TrackID:    opts.TrackID,
		Limit:      opts.Limit,
		ListenedAt: opts.ListenedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("GetExportPage: %w", err)
	}
	ret := make([]*db.ExportItem, len(rows))
	for i, row := range rows {

		var trackAliases []models.Alias
		err = json.Unmarshal(row.TrackAliases, &trackAliases)
		if err != nil {
			return nil, fmt.Errorf("GetExportPage: json.Unmarshal trackAliases: %w", err)
		}
		var albumAliases []models.Alias
		err = json.Unmarshal(row.ReleaseAliases, &albumAliases)
		if err != nil {
			return nil, fmt.Errorf("GetExportPage: json.Unmarshal albumAliases: %w", err)
		}
		var artists []models.ArtistWithFullAliases
		err = json.Unmarshal(row.Artists, &artists)
		if err != nil {
			return nil, fmt.Errorf("GetExportPage: json.Unmarshal artists: %w", err)
		}

		ret[i] = &db.ExportItem{
			TrackID:            row.TrackID,
			ListenedAt:         row.ListenedAt,
			UserID:             row.UserID,
			Client:             row.Client,
			TrackMbid:          row.TrackMbid,
			TrackDuration:      row.TrackDuration,
			TrackAliases:       trackAliases,
			ReleaseID:          row.ReleaseID,
			ReleaseMbid:        row.ReleaseMbid,
			ReleaseImageSource: row.ReleaseImageSource.String,
			VariousArtists:     row.VariousArtists,
			ReleaseAliases:     albumAliases,
			Artists:            artists,
		}
	}
	return ret, nil
}
