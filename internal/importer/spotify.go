package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
)

type SpotifyExportItem struct {
	Timestamp  time.Time `json:"ts"`
	TrackName  string    `json:"master_metadata_track_name"`
	ArtistName string    `json:"master_metadata_album_artist_name"`
	AlbumName  string    `json:"master_metadata_album_album_name"`
	ReasonEnd  string    `json:"reason_end"`
	MsPlayed   int32     `json:"ms_played"`
}

func ImportSpotifyFile(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning spotify import on file: %s", filename)
	file, err := os.Open(path.Join(cfg.ConfigDir(), "import", filename))
	if err != nil {
		l.Err(err).Msgf("Failed to read import file: %s", filename)
		return fmt.Errorf("ImportSpotifyFile: %w", err)
	}
	defer file.Close()
	export := make([]SpotifyExportItem, 0)
	err = json.NewDecoder(file).Decode(&export)
	if err != nil {
		return fmt.Errorf("ImportSpotifyFile: %w", err)
	}

	bs := NewBulkSubmitter(ctx, BulkSubmitterOpts{
		Store: store,
		Mbzc:  mbzc,
	})

	for _, item := range export {
		if item.ReasonEnd != "trackdone" {
			continue
		}
		if !inImportTimeWindow(item.Timestamp) {
			continue
		}
		if item.TrackName == "" || item.ArtistName == "" {
			l.Debug().Msg("Skipping non-track item")
			continue
		}
		bs.Accept(catalog.SubmitListenOpts{
			MbzCaller:      mbzc,
			Artist:         item.ArtistName,
			TrackTitle:     item.TrackName,
			ReleaseTitle:   item.AlbumName,
			Duration:       item.MsPlayed / 1000,
			Time:           item.Timestamp,
			Client:         "spotify",
			UserID:         1,
			SkipCacheImage: true,
		})
	}

	count, err := bs.Flush()
	if err != nil {
		return fmt.Errorf("ImportSpotifyFile: %w", err)
	}
	return finishImport(ctx, filename, count)
}
