package mbz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gabehf/koito/internal/logger"
	"github.com/google/uuid"
)

// MusicBrainz search API response types

type musicBrainzSearchResponse struct {
	Recordings []musicBrainzSearchRecording `json:"recordings"`
}

type musicBrainzSearchRecording struct {
	ID           string                    `json:"id"`
	Title        string                    `json:"title"`
	Length       int                       `json:"length"`
	ArtistCredit []MusicBrainzArtistCredit `json:"artist-credit"`
	Releases     []musicBrainzSearchRelease `json:"releases"`
}

type musicBrainzSearchRelease struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	Status       string `json:"status"`
	ReleaseGroup struct {
		ID string `json:"id"`
	} `json:"release-group"`
}

// MusicBrainzSearchResult holds the resolved IDs from a search-by-name query.
type MusicBrainzSearchResult struct {
	RecordingID    uuid.UUID
	ReleaseID      uuid.UUID
	ReleaseGroupID uuid.UUID
	ReleaseTitle   string
	DurationMs     int
}

// SearchRecording searches MusicBrainz for a recording by artist and track name.
// It returns the best match that passes a confidence filter (case-insensitive exact
// match on title and at least one artist credit), or nil if no confident match is found.
func (c *MusicBrainzClient) SearchRecording(ctx context.Context, artist string, track string) (*MusicBrainzSearchResult, error) {
	l := logger.FromContext(ctx)

	query := fmt.Sprintf("artist:\"%s\" AND recording:\"%s\"",
		escapeLucene(artist), escapeLucene(track))
	url := fmt.Sprintf("%s/ws/2/recording/?query=%s&limit=5&fmt=json",
		c.url, queryEscape(query))

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("SearchRecording: %w", err)
	}

	body, err := c.queue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("SearchRecording: %w", err)
	}

	var resp musicBrainzSearchResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		l.Err(err).Str("body", string(body)).Msg("Failed to unmarshal MusicBrainz search response")
		return nil, fmt.Errorf("SearchRecording: %w", err)
	}

	for _, rec := range resp.Recordings {
		if !strings.EqualFold(rec.Title, track) {
			continue
		}
		if !hasArtistCredit(rec.ArtistCredit, artist) {
			continue
		}

		recordingID, err := uuid.Parse(rec.ID)
		if err != nil {
			continue
		}

		release := pickBestRelease(rec.Releases)
		if release == nil {
			continue
		}

		releaseID, err := uuid.Parse(release.ID)
		if err != nil {
			continue
		}
		releaseGroupID, err := uuid.Parse(release.ReleaseGroup.ID)
		if err != nil {
			continue
		}

		l.Debug().Msgf("MBZ search matched: '%s' by '%s' -> recording=%s release=%s",
			track, artist, recordingID, releaseID)

		return &MusicBrainzSearchResult{
			RecordingID:    recordingID,
			ReleaseID:      releaseID,
			ReleaseGroupID: releaseGroupID,
			ReleaseTitle:   release.Title,
			DurationMs:     rec.Length,
		}, nil
	}

	l.Debug().Msgf("MBZ search: no confident match for '%s' by '%s'", track, artist)
	return nil, nil
}

// hasArtistCredit checks whether at least one artist credit name matches (case-insensitive).
func hasArtistCredit(credits []MusicBrainzArtistCredit, artist string) bool {
	for _, ac := range credits {
		if strings.EqualFold(ac.Name, artist) || strings.EqualFold(ac.Artist.Name, artist) {
			return true
		}
	}
	return false
}

// pickBestRelease prefers an Official release, then falls back to the first available.
func pickBestRelease(releases []musicBrainzSearchRelease) *musicBrainzSearchRelease {
	if len(releases) == 0 {
		return nil
	}
	for i := range releases {
		if releases[i].Status == "Official" {
			return &releases[i]
		}
	}
	return &releases[0]
}

// escapeLucene escapes special characters in Lucene query syntax.
func escapeLucene(s string) string {
	replacer := strings.NewReplacer(
		`\`, `\\`,
		`+`, `\+`,
		`-`, `\-`,
		`!`, `\!`,
		`(`, `\(`,
		`)`, `\)`,
		`{`, `\{`,
		`}`, `\}`,
		`[`, `\[`,
		`]`, `\]`,
		`^`, `\^`,
		`"`, `\"`,
		`~`, `\~`,
		`*`, `\*`,
		`?`, `\?`,
		`:`, `\:`,
		`/`, `\/`,
	)
	return replacer.Replace(s)
}

// queryEscape percent-encodes a query string value for use in a URL.
// We use a simple implementation to avoid importing net/url just for this.
func queryEscape(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if isUnreserved(c) {
			b.WriteByte(c)
		} else {
			fmt.Fprintf(&b, "%%%02X", c)
		}
	}
	return b.String()
}

func isUnreserved(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '_' || c == '.' || c == '~'
}
