package importer

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/catalog"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
	mbz "github.com/gabehf/koito/internal/mbz"
)

// appleMusicLibraryTrack represents a track from Apple Music Library Tracks.json
type appleMusicLibraryTrack struct {
	Title  string `json:"Title"`
	Artist string `json:"Artist"`
	Album  string `json:"Album"`
}

// appleMusicLibraryLookup is a map of lowercase title -> artist/album
// for quick lookup during import
type appleMusicLibraryLookup map[string]struct {
	Artist string
	Album  string
}

// Apple Music Play Activity CSV column indices (0-based).
// These are computed from the header at parse time.
type appleMusicColumns struct {
	songName            int
	eventTimestamp      int
	eventType           int
	playDurationMs      int
	endReasonType       int
	containerArtistName int
	containerAlbumName  int
}

// findLibraryTracksFile checks if the Library Tracks JSON file exists in the import directory.
// Returns the filename if found, empty string otherwise.
func findLibraryTracksFile(ctx context.Context) string {
	_ = ctx
	libDir := path.Join(cfg.ConfigDir(), "import")
	files, err := os.ReadDir(libDir)
	if err != nil {
		return ""
	}

	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".json") {
			continue
		}
		if strings.Contains(strings.ToLower(f.Name()), "library tracks") {
			return f.Name()
		}
	}

	return ""
}

// loadAppleMusicLibrary loads the Library Tracks JSON file from the given filename.
// Returns a lookup map keyed by lowercase song title -> artist/album.
func loadAppleMusicLibrary(ctx context.Context, filename string) appleMusicLibraryLookup {
	l := logger.FromContext(ctx)
	result := make(appleMusicLibraryLookup)

	libDir := path.Join(cfg.ConfigDir(), "import")
	filePath := path.Join(libDir, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		l.Err(err).Str("file", filename).Msg("Failed to read library tracks file")
		return result
	}

	var tracks []appleMusicLibraryTrack
	if err := json.Unmarshal(data, &tracks); err != nil {
		l.Err(err).Str("file", filename).Msg("Failed to parse library tracks file")
		return result
	}

	for _, t := range tracks {
		title := strings.TrimSpace(t.Title)
		artist := strings.TrimSpace(t.Artist)
		album := strings.TrimSpace(t.Album)
		if title != "" {
			result[title] = struct {
				Artist string
				Album  string
			}{Artist: artist, Album: album}
		}
	}
	l.Info().Msgf("Loaded %d tracks from %s for artist lookup", len(result), filename)

	return result
}

// parseAppleMusicColumns reads the CSV header and returns column indices for the fields we need.
func parseAppleMusicColumns(r *csv.Reader) (appleMusicColumns, error) {
	header, err := r.Read()
	if err != nil {
		return appleMusicColumns{}, fmt.Errorf("parseAppleMusicColumns: %w", err)
	}

	cols := appleMusicColumns{}
	for i, name := range header {
		switch name {
		case "Song Name":
			cols.songName = i
		case "Event Timestamp":
			cols.eventTimestamp = i
		case "Event Type":
			cols.eventType = i
		case "Play Duration Milliseconds":
			cols.playDurationMs = i
		case "End Reason Type":
			cols.endReasonType = i
		case "Container Artist Name":
			cols.containerArtistName = i
		case "Container Album Name":
			cols.containerAlbumName = i
		}
	}
	return cols, nil
}

// parseISO8601Time parses an ISO 8601 timestamp from Apple Music exports.
// Examples: "2024-10-20T18:31:29.435Z", "2024-09-25T12:37:25.478Z"
func parseISO8601Time(ts string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, ts)
	if err != nil {
		return time.Time{}
	}
	return parsed
}

// parseAppleMusicDuration converts milliseconds to seconds.
func parseAppleMusicDuration(ms string) int32 {
	var val int64
	fmt.Sscanf(ms, "%d", &val)
	return int32(val / 1000)
}

// ImportAppleMusicPlayActivityFile imports an Apple Music Play Activity CSV export.
// The file is expected to be in the import directory under cfg.ConfigDir().
//
// Artist resolution follows this priority:
//  1. Container Artist Name from CSV (always empty in current exports)
//  2. Library Tracks JSON lookup by song title (exact match)
//  3. MusicBrainz search as fallback
//
// The Play Activity CSV has 200+ columns. This importer maps:
//   - Song Name -> TrackTitle
//   - Library Tracks JSON -> Artist + Album
//   - Event Timestamp -> Time
//   - Play Duration Milliseconds -> Duration
//
// Only PLAY_END events are imported (equivalent to Spotify's "trackdone" filter).
func ImportAppleMusicPlayActivityFile(ctx context.Context, store importStore, mbzc mbz.MusicBrainzCaller, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning Apple Music import on file: %s", filename)

	// Check that Library Tracks JSON file exists
	libraryFile := findLibraryTracksFile(ctx)
	if libraryFile == "" {
		l.Warn().Msg("Apple Music import skipped: Apple Music Library Tracks.json not found in import directory.")
		return nil
	}
	l.Info().Msgf("Using library file: %s for artist lookup", libraryFile)

	// Load library tracks for artist lookup
	library := loadAppleMusicLibrary(ctx, libraryFile)

	file, err := os.Open(path.Join(cfg.ConfigDir(), "import", filename))
	if err != nil {
		l.Err(err).Msgf("Failed to read import file: %s", filename)
		return fmt.Errorf("ImportAppleMusicPlayActivityFile: %w", err)
	}
	defer file.Close()

	var throttleFunc = func() {}
	if ms := cfg.ThrottleImportMs(); ms > 0 {
		throttleFunc = func() {
			time.Sleep(time.Duration(ms) * time.Millisecond)
		}
	}

	reader := csv.NewReader(bufio.NewReader(file))
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true

	cols, err := parseAppleMusicColumns(reader)
	if err != nil {
		return fmt.Errorf("ImportAppleMusicPlayActivityFile: %w", err)
	}

	count := 0
	imported := 0

	for {
		record, err := reader.Read()
		if err != nil {
			break
		}
		count++

		// Filter: only PLAY_END events (equivalent to Spotify's "trackdone")
		if cols.eventType >= len(record) || record[cols.eventType] != "PLAY_END" {
			continue
		}

		// Get the song name
		if cols.songName >= len(record) || record[cols.songName] == "" {
			continue
		}
		songName := record[cols.songName]

		// Get the timestamp
		var eventTime time.Time
		if cols.eventTimestamp < len(record) && record[cols.eventTimestamp] != "" {
			eventTime = parseISO8601Time(record[cols.eventTimestamp])
		}
		if eventTime.IsZero() {
			l.Debug().Str("song", songName).Msg("Skipping import with zero timestamp")
			continue
		}

		// Time window check
		if !inImportTimeWindow(eventTime) {
			l.Debug().Msgf("Skipping import due to import time rules")
			continue
		}

		// Resolve artist and album - priority: CSV -> Library JSON -> MusicBrainz
		artist := ""
		releaseTitle := ""

		// 1. Container Artist Name from CSV (always empty in current exports)
		if cols.containerArtistName < len(record) && record[cols.containerArtistName] != "" {
			artist = record[cols.containerArtistName]
			if cols.containerAlbumName < len(record) && record[cols.containerAlbumName] != "" {
				releaseTitle = record[cols.containerAlbumName]
			}
		} else if library != nil {
			// 2. Library Tracks JSON lookup
			if lookup, ok := library[songName]; ok {
				artist = lookup.Artist
				releaseTitle = lookup.Album
			} else {
				l.Debug().Str("title", songName).Msg("Song not found in library tracks, attempting MusicBrainz lookup")
			}
		}

		// 3. MusicBrainz fallback (if artist still empty)
		if artist == "" {
			artist, err = mbzc.SearchRecording(ctx, songName, "")
			if err != nil {
				l.Debug().AnErr("error", err).Str("title", songName).Msg("MusicBrainz lookup failed")
				continue
			}
			if artist == "" {
				l.Debug().Str("title", songName).Msg("No artist found via MusicBrainz, skipping")
				continue
			}
		}

		// Get duration from CSV if available
		duration := int32(0)
		if cols.playDurationMs < len(record) && record[cols.playDurationMs] != "" {
			duration = parseAppleMusicDuration(record[cols.playDurationMs])
		}

		opts := catalog.SubmitListenOpts{
			MbzCaller:      mbzc,
			Artist:         artist,
			TrackTitle:     songName,
			ReleaseTitle:   releaseTitle,
			Duration:       duration,
			Time:           eventTime.Truncate(time.Second),
			Client:         "applemusic",
			UserID:         1,
			SkipCacheImage: !cfg.FetchImagesDuringImport(),
		}

		err = catalog.SubmitListen(ctx, store, opts)
		if err != nil {
			l.Err(err).Msg("Failed to import Apple Music playback item")
			continue
		}
		imported++
		throttleFunc()
	}

	l.Info().Msgf("Finished importing %s; imported %d items out of %d total rows", filename, imported, count)
	return finishImport(ctx, filename, count)
}
