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
	mbz "github.com/gabehf/koito/internal/mbz"
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
	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
	export := make([]SpotifyExportItem, 0)
	err = json.NewDecoder(file).Decode(&export)
	if err != nil {
		return fmt.Errorf("ImportSpotifyFile: %w", err)
	}

	for _, item := range export {
		if item.ReasonEnd != "trackdone" {
			continue
		}
		if !inImportTimeWindow(item.Timestamp) {
			l.Debug().Msgf("Skipping import due to import time rules")
			continue
		}
		dur := item.MsPlayed
		if item.TrackName == "" || item.ArtistName == "" {
			l.Debug().Msg("Skipping non-track item")
			continue
		}
		opts := catalog.SubmitListenOpts{
			MbzCaller:      mbzc,
			Artist:         item.ArtistName,
			TrackTitle:     item.TrackName,
			ReleaseTitle:   item.AlbumName,
			Duration:       dur / 1000,
			Time:           item.Timestamp,
			Client:         "spotify",
			UserID:         1,
			SkipCacheImage: !cfg.FetchImagesDuringImport(),
		}
		err = catalog.SubmitListen(ctx, store, opts)
		if err != nil {
			l.Err(err).Msg("Failed to import spotify playback item")
			return fmt.Errorf("ImportSpotifyFile: %w", err)
		}
		throttleFunc()
	}
	return finishImport(ctx, filename, len(export))
}
