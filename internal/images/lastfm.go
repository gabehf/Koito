package images

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/queue"
	"github.com/google/uuid"
)

// i told gemini to write this cuz i figured it would be simple enough and
// it looks like it just works? maybe ai is actually worth one quintillion gallons of water

type LastFMClient struct {
	apiKey       string
	baseUrl      string
	userAgent    string
	requestQueue *queue.RequestQueue
}

// LastFM JSON structures use "#text" for the value of XML-mapped fields
type lastFMImage struct {
	URL  string `json:"#text"`
	Size string `json:"size"`
}

type lastFMAlbumResponse struct {
	Album struct {
		Name  string        `json:"name"`
		Image []lastFMImage `json:"image"`
	} `json:"album"`
	Error   int    `json:"error"`
	Message string `json:"message"`
}

type lastFMArtistResponse struct {
	Artist struct {
		Name  string        `json:"name"`
		Image []lastFMImage `json:"image"`
	} `json:"artist"`
	Error   int    `json:"error"`
	Message string `json:"message"`
}

const (
	lastFMApiBaseUrl = "http://ws.audioscrobbler.com/2.0/"
)

func NewLastFMClient() *LastFMClient {
	ret := new(LastFMClient)
	ret.apiKey = cfg.LastFMApiKey()
	ret.baseUrl = lastFMApiBaseUrl
	ret.userAgent = cfg.UserAgent()
	ret.requestQueue = queue.NewRequestQueue(5, 5)
	return ret
}

func (c *LastFMClient) queue(ctx context.Context, req *http.Request) ([]byte, error) {
	l := logger.FromContext(ctx)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resultChan := c.requestQueue.Enqueue(func(client *http.Client, done chan<- queue.RequestResult) {
		resp, err := client.Do(req)
		if err != nil {
			l.Debug().Err(err).Str("url", req.URL.String()).Msg("Failed to contact LastFM")
			done <- queue.RequestResult{Err: err}
			return
		}
		defer resp.Body.Close()

		// LastFM might return 200 OK even for API errors (like "Artist not found"),
		// so we rely on parsing the JSON body for logic errors later,
		// but we still check for HTTP protocol failures here.
		if resp.StatusCode >= 500 {
			err = fmt.Errorf("received server error from LastFM: %s", resp.Status)
			done <- queue.RequestResult{Body: nil, Err: err}
			return
		}

		body, err := io.ReadAll(resp.Body)
		done <- queue.RequestResult{Body: body, Err: err}
	})

	result := <-resultChan
	return result.Body, result.Err
}

func (c *LastFMClient) getEntity(ctx context.Context, params url.Values, result any) error {
	l := logger.FromContext(ctx)

	// Add standard parameters
	params.Set("api_key", c.apiKey)
	params.Set("format", "json")

	// Construct URL
	reqUrl, _ := url.Parse(c.baseUrl)
	reqUrl.RawQuery = params.Encode()

	l.Debug().Msgf("Sending request to LastFM: GET %s", reqUrl.String())

	req, err := http.NewRequest("GET", reqUrl.String(), nil)
	if err != nil {
		return fmt.Errorf("getEntity: %w", err)
	}

	l.Debug().Msg("Adding LastFM request to queue")
	body, err := c.queue(ctx, req)
	if err != nil {
		l.Err(err).Msg("LastFM request failed")
		return fmt.Errorf("getEntity: %w", err)
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		l.Err(err).Msg("Failed to unmarshal LastFM response")
		return fmt.Errorf("getEntity: %w", err)
	}

	return nil
}

// selectBestImage picks the largest available image from the LastFM slice
func (c *LastFMClient) selectBestImage(images []lastFMImage) string {
	// Rank preference: mega > extralarge > large > medium > small
	// Since LastFM usually returns them in order of size, we could take the last one,
	// but a map lookup is safer against API changes.

	imgMap := make(map[string]string)
	for _, img := range images {
		if img.URL != "" {
			imgMap[img.Size] = img.URL
		}
	}

	if url, ok := imgMap["mega"]; ok {
		if err := ValidateImageURL(overrideImgSize(url)); err == nil {
			return overrideImgSize(url)
		} else {
			return url
		}
	}
	if url, ok := imgMap["extralarge"]; ok {
		if err := ValidateImageURL(overrideImgSize(url)); err == nil {
			return overrideImgSize(url)
		} else {
			return url
		}
	}
	if url, ok := imgMap["large"]; ok {
		if err := ValidateImageURL(overrideImgSize(url)); err == nil {
			return overrideImgSize(url)
		} else {
			return url
		}
	}
	if url, ok := imgMap["medium"]; ok {
		return url
	}
	if url, ok := imgMap["small"]; ok {
		return url
	}

	return ""
}

// lastfm seems to only return a 300x300 image even for "mega" and "extralarge" images, so I'm cheating
func overrideImgSize(url string) string {
	return strings.Replace(url, "300x300", "600x600", 1)
}

func (c *LastFMClient) GetAlbumImage(ctx context.Context, mbid *uuid.UUID, artist, album string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(lastFMAlbumResponse)
	l.Debug().Msgf("Finding album image for %s from artist %s", album, artist)

	// Helper to run the fetch
	fetch := func(query paramsBuilder) error {
		params := url.Values{}
		params.Set("method", "album.getInfo")
		query(params)
		return c.getEntity(ctx, params, resp)
	}

	// 1. Try MBID search first
	if mbid != nil {
		l.Debug().Str("mbid", mbid.String()).Msg("Searching album image by MBID")
		err := fetch(func(p url.Values) {
			p.Set("mbid", mbid.String())
		})

		// If success and no API error code
		if err == nil && resp.Error == 0 && len(resp.Album.Image) > 0 {
			best := c.selectBestImage(resp.Album.Image)
			if best != "" {
				return best, nil
			}
		} else if resp.Error != 0 {
			l.Debug().Int("api_error", resp.Error).Msg("LastFM MBID lookup failed, falling back to name")
		}
	}

	// 2. Fallback to Artist + Album name match
	l.Debug().Str("title", album).Str("artist", artist).Msg("Searching album image by title and artist")

	// Clear previous response structure just in case
	resp = new(lastFMAlbumResponse)

	err := fetch(func(p url.Values) {
		p.Set("artist", artist)
		p.Set("album", album)
		// Auto-correct spelling is useful for name lookups
		p.Set("autocorrect", "1")
	})

	if err != nil {
		return "", fmt.Errorf("GetAlbumImage: %v", err)
	}

	if resp.Error != 0 {
		return "", fmt.Errorf("GetAlbumImage: LastFM API error %d: %s", resp.Error, resp.Message)
	}

	best := c.selectBestImage(resp.Album.Image)
	if best == "" {
		return "", fmt.Errorf("GetAlbumImage: no suitable image found")
	}

	return best, nil
}

func (c *LastFMClient) GetArtistImage(ctx context.Context, mbid *uuid.UUID, artist string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(lastFMArtistResponse)
	l.Debug().Msgf("Finding artist image for %s", artist)

	fetch := func(query paramsBuilder) error {
		params := url.Values{}
		params.Set("method", "artist.getInfo")
		query(params)
		return c.getEntity(ctx, params, resp)
	}

	// 1. Try MBID search
	if mbid != nil {
		l.Debug().Str("mbid", mbid.String()).Msg("Searching artist image by MBID")
		err := fetch(func(p url.Values) {
			p.Set("mbid", mbid.String())
		})

		if err == nil && resp.Error == 0 && len(resp.Artist.Image) > 0 {
			best := c.selectBestImage(resp.Artist.Image)
			if best != "" {
				// Validate to match Subsonic implementation behavior
				if err := ValidateImageURL(best); err == nil {
					return best, nil
				}
			}
		}
	}

	// 2. Fallback to Artist name
	l.Debug().Str("artist", artist).Msg("Searching artist image by name")
	resp = new(lastFMArtistResponse)

	err := fetch(func(p url.Values) {
		p.Set("artist", artist)
		p.Set("autocorrect", "1")
	})

	if err != nil {
		return "", fmt.Errorf("GetArtistImage: %v", err)
	}

	if resp.Error != 0 {
		return "", fmt.Errorf("GetArtistImage: LastFM API error %d: %s", resp.Error, resp.Message)
	}

	best := c.selectBestImage(resp.Artist.Image)
	if best == "" {
		return "", fmt.Errorf("GetArtistImage: no suitable image found")
	}

	if err := ValidateImageURL(best); err != nil {
		return "", fmt.Errorf("GetArtistImage: failed to validate image url")
	}

	return best, nil
}

type paramsBuilder func(url.Values)
