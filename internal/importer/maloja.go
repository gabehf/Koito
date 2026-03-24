package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/utils"
)

type MalojaAlbum struct {
	Title string `json:"albumtitle"`
}

type MalojaFile struct {
	Scrobbles []MalojaExportItem `json:"scrobbles"`
	List      []MalojaExportItem `json:"list"`
}
type MalojaExportItem struct {
	Time  int64       `json:"time"`
	Track MalojaTrack `json:"track"`
}
type MalojaTrack struct {
	Artists []string     `json:"artists"`
	Title   string       `json:"title"`
	Album   *MalojaAlbum `json:"album"`
}

func ImportMalojaFile(ctx context.Context, store db.DB, mbzc mbz.MusicBrainzCaller, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning maloja import on file: %s", filename)
	file, err := os.Open(path.Join(cfg.ConfigDir(), "import", filename))
	if err != nil {
		l.Err(err).Msgf("Failed to read import file: %s", filename)
		return fmt.Errorf("ImportMalojaFile: %w", err)
	}
	defer file.Close()
	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}
	export := new(MalojaFile)
	err = json.NewDecoder(file).Decode(&export)
	if err != nil {
		return fmt.Errorf("ImportMalojaFile: %w", err)
	}
	items := export.Scrobbles
	if len(items) == 0 {
		items = export.List
	}
	count := 0
	total := len(items)
	for i, item := range items {
		martists := make([]string, 0)
		// Maloja has a tendency to have the the artist order ['feature', 'main ● feature'], so
		// here we try to turn that artist array into ['main', 'feature']
		item.Track.Artists = utils.MoveFirstMatchToFront(item.Track.Artists, " \u2022 ")
		for _, an := range item.Track.Artists {
			ans := strings.Split(an, " \u2022 ")
			martists = append(martists, ans...)
		}
		artists := utils.UniqueIgnoringCase(martists)
		if len(item.Track.Artists) < 1 || item.Track.Title == "" {
			l.Debug().Msg("Skipping invalid maloja import item")
			continue
		}
		ts := time.Unix(item.Time, 0)
		if !inImportTimeWindow(ts) {
			l.Debug().Msgf("Skipping import due to import time rules")
			continue
		}
		releaseTitle := ""
		if item.Track.Album != nil {
			releaseTitle = item.Track.Album.Title
		}
		opts := catalog.SubmitListenOpts{
			MbzCaller:      mbzc,
			Artist:         item.Track.Artists[0],
			ArtistNames:    artists,
			TrackTitle:     item.Track.Title,
			ReleaseTitle:   releaseTitle,
			Time:           ts.Local(),
			Client:         "maloja",
			UserID:         1,
			SkipCacheImage: !cfg.FetchImagesDuringImport(),
		}
		err = catalog.SubmitListen(ctx, store, opts)
		if err != nil {
			l.Err(err).Msgf("Failed to import maloja item %d/%d", i+1, total)
			continue
		}
		count++
		if count%500 == 0 {
			l.Info().Msgf("Maloja import progress: %d/%d", count, total)
		}
		throttleFunc()
	}
	return finishImport(ctx, filename, count)
}
