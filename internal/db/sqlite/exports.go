package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gabehf/koito/internal/db"
)

func (s *Sqlite) GetExportPage(ctx context.Context, opts db.GetExportPageOpts) ([]*db.ExportItem, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT l.listened_at, l.user_id, l.client,
		       t.id AS track_id, t.musicbrainz_id AS track_mbid, t.duration,
		       t.release_id,
		       r.musicbrainz_id AS release_mbid, r.image, r.image_source, r.various_artists
		FROM listens l
		JOIN tracks t ON l.track_id = t.id
		JOIN releases r ON t.release_id = r.id
		WHERE l.user_id = ?
		  AND (l.listened_at > ? OR (l.listened_at = ? AND l.track_id > ?))
		ORDER BY l.listened_at, l.track_id
		LIMIT ?`,
		opts.UserID,
		opts.ListenedAt.Unix(), opts.ListenedAt.Unix(), opts.TrackID,
		opts.Limit,
	)
	if err != nil {
		return nil, fmt.Errorf("GetExportPage: %w", err)
	}
	defer rows.Close()

	var items []*db.ExportItem
	for rows.Next() {
		var item db.ExportItem
		var listenedAt int64
		var client sql.NullString
		var trackMbid, releaseMbid, releaseImage, releaseImageSrc sql.NullString
		var variousArtists int

		if err := rows.Scan(
			&listenedAt, &item.UserID, &client,
			&item.TrackID, &trackMbid, &item.TrackDuration,
			&item.ReleaseID,
			&releaseMbid, &releaseImage, &releaseImageSrc, &variousArtists,
		); err != nil {
			return nil, fmt.Errorf("GetExportPage: scan: %w", err)
		}

		item.ListenedAt = time.Unix(listenedAt, 0).UTC()
		if client.Valid && client.String != "" {
			item.Client = &client.String
		}
		item.TrackMbid = parseNullableUUID(trackMbid)
		item.ReleaseMbid = parseNullableUUID(releaseMbid)
		item.ReleaseImage = parseNullableUUID(releaseImage)
		item.ReleaseImageSource = releaseImageSrc.String
		item.VariousArtists = variousArtists == 1

		trackAliases, err := s.getAliasesForEntity(ctx, "track_aliases", "track_id", item.TrackID)
		if err != nil {
			return nil, fmt.Errorf("GetExportPage: track aliases: %w", err)
		}
		item.TrackAliases = trackAliases

		releaseAliases, err := s.getAliasesForEntity(ctx, "release_aliases", "release_id", item.ReleaseID)
		if err != nil {
			return nil, fmt.Errorf("GetExportPage: release aliases: %w", err)
		}
		item.ReleaseAliases = releaseAliases

		artists, err := s.artistsWithAliasesForTrack(ctx, item.TrackID)
		if err != nil {
			return nil, fmt.Errorf("GetExportPage: artists: %w", err)
		}
		item.Artists = artists

		items = append(items, &item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

