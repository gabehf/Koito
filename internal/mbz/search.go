package mbz

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabehf/koito/internal/logger"
)

// musicBrainzSearchRecording represents a single recording result from MusicBrainz search.
type musicBrainzSearchRecording struct {
	Title        string                    `json:"title"`
	ArtistCredit []musicBrainzArtistCredit `json:"artist-credit"`
}

// musicBrainzArtistCredit is the simplified version used in search results.
type musicBrainzArtistCredit struct {
	Name   string `json:"name"`
	Artist struct {
		Name string `json:"name"`
	} `json:"artist"`
}

// musicBrainzSearchResponse is the wrapper for MusicBrainz search results.
type musicBrainzSearchResponse struct {
	Recordings []musicBrainzSearchRecording `json:"recordings"`
}

// SearchRecording searches MusicBrainz for recordings matching the given title.
// If artist is provided, it narrows the search to that artist.
// Returns the first matching artist name, or empty string if not found.
func (c *MusicBrainzClient) SearchRecording(ctx context.Context, title string, artist string) (string, error) {
	l := logger.FromContext(ctx)

	// Build the MusicBrainz query
	query := url.QueryEscape("recording:" + title)
	if artist != "" {
		query += "+" + url.QueryEscape("artist:" + artist)
	}

	fullURL := fmt.Sprintf("%s/ws/2/recording/?query=%s&fmt=json&limit=1", c.url, query)
	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		l.Err(err).Msg("Failed to build MusicBrainz search request")
		return "", fmt.Errorf("SearchRecording: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	body, err := c.queue(ctx, req)
	if err != nil {
		return "", fmt.Errorf("SearchRecording: %w", err)
	}

	var result musicBrainzSearchResponse
	if err := json.Unmarshal(body, &result); err != nil {
		l.Err(err).Msg("Failed to unmarshal MusicBrainz search response")
		return "", fmt.Errorf("SearchRecording: %w", err)
	}

	if len(result.Recordings) < 1 {
		l.Debug().Str("title", title).Msg("No recordings found in MusicBrainz")
		return "", nil
	}

	recording := result.Recordings[0]

	// Try to get artist from the artist credit
	if len(recording.ArtistCredit) > 0 {
		name := recording.ArtistCredit[0].Artist.Name
		if name == "" {
			name = recording.ArtistCredit[0].Name
		}
		if name != "" {
			l.Debug().Str("title", title).Str("artist", name).Msg("Found artist from MusicBrainz")
			return name, nil
		}
	}

	// Fallback: use the title's artist part if present (format: "Artist - Title")
	if idx := strings.Index(recording.Title, " - "); idx > 0 {
		artistPart := strings.TrimSpace(recording.Title[:idx])
		if artistPart != "" && artistPart != title {
			l.Debug().Str("title", title).Str("artist", artistPart).Msg("Extracted artist from recording title")
			return artistPart, nil
		}
	}

	l.Debug().Str("title", title).Msg("Recording found but no artist info available")
	return "", nil
}
