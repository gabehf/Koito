package importer

import (
	"context"
	"fmt"
	"path"
	"strconv"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
	mbz "github.com/gabehf/koito/internal/mbz"
	"github.com/xuri/excelize/v2"
)

// DeezerExportItem maps a row from the listeningHistory sheet.
type DeezerExportItem struct {
	Timestamp  time.Time
	TrackName  string
	ArtistName string
	AlbumName  string
	MsPlayed   int32
	Platform   string
}

func ImportDeezerFile(ctx context.Context, store importStore, mbzc mbz.MusicBrainzCaller, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning deezer import on file: %s", filename)

	filePath := path.Join(cfg.ConfigDir(), "import", filename)
	f, err := excelize.OpenFile(filePath)
	if err != nil {
		l.Err(err).Msgf("Failed to open xlsx file: %s", filename)
		return fmt.Errorf("ImportDeezerFile: %w", err)
	}
	defer f.Close()

	// Read country from customizationData (row 2, column G)
	country, _ := f.GetCellValue("2_customizationData", "G2")
	if country == "" {
		country = "unknown"
	}

	// Read listeningHistory sheet
	rows, err := f.GetRows("10_listeningHistory")
	if err != nil {
		return fmt.Errorf("ImportDeezerFile: failed to read rows: %w", err)
	}

	var items []DeezerExportItem
	for i, row := range rows {
		if i == 0 {
			continue // skip header
		}
		if len(row) < 9 {
			continue
		}

		track := row[0]
		artist := row[1]
		// row[2] = ISRC (not used by koito)
		album := row[3]
		// row[4] = IP address (not used)
		listeningTime := row[5]
		platform := row[6]
		// row[7] = Platform Model (not used)
		dateStr := row[8]

		if track == "" || artist == "" {
			continue // skip non-track rows
		}

		// Parse date: "2013-09-07 12:42:21"
		ts, err := time.Parse("2006-01-02 15:04:05", dateStr)
		if err != nil {
			l.Debug().Msgf("Skipping row %d: bad date format: %s", i+1, dateStr)
			continue
		}

		// Parse listening time: "-1" means unknown → 0
		msPlayed := int32(0)
		if listeningTime != "-1" && listeningTime != "" {
			secs, err := strconv.ParseInt(listeningTime, 10, 32)
			if err == nil {
				msPlayed = int32(secs)
			}
		}

		// Normalize platform
		if platform == "None" || platform == "" {
			platform = "deezer"
		}

		items = append(items, DeezerExportItem{
			Timestamp:  ts,
			TrackName:  track,
			ArtistName: artist,
			AlbumName:  album,
			MsPlayed:   msPlayed,
			Platform:   platform,
		})
	}

	l.Info().Msgf("Parsed %d listening records (country: %s)", len(items), country)

	// Throttle if configured
	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}

	// Import each item
	count := 0
	for _, item := range items {
		if !inImportTimeWindow(item.Timestamp) {
			continue
		}

		opts := catalog.SubmitListenOpts{
			MbzCaller:      mbzc,
			Artist:         item.ArtistName,
			TrackTitle:     item.TrackName,
			ReleaseTitle:   item.AlbumName,
			Duration:       item.MsPlayed, // already in seconds
			Time:           item.Timestamp,
			Client:         "deezer",
			UserID:         1,
			SkipCacheImage: !cfg.FetchImagesDuringImport(),
		}

		err := catalog.SubmitListen(ctx, store, opts)
		if err != nil {
			l.Err(err).Msgf("Failed to import listen: %s by %s", item.TrackName, item.ArtistName)
			continue // don't abort on single failure
		}

		throttleFunc()
		count++
	}

	return finishImport(ctx, filename, count)
}
