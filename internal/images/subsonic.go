package images

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/queue"
	"github.com/google/uuid"
)

type SubsonicClient struct {
	url          string
	userAgent    string
	authParams   string
	requestQueue *queue.RequestQueue
}

type SubsonicAlbumResponse struct {
	SubsonicResponse struct {
		Status        string `json:"status"`
		SearchResult3 struct {
			Album []struct {
				CoverArt string `json:"coverArt"`
				Artist   string `json:"artist"`
				MBID     string `json:"musicBrainzId"`
			} `json:"album"`
		} `json:"searchResult3"`
	} `json:"subsonic-response"`
}

type SubsonicArtistResponse struct {
	SubsonicResponse struct {
		Status        string `json:"status"`
		SearchResult3 struct {
			Artist []struct {
				ArtistImageUrl string `json:"artistImageUrl"`
			} `json:"artist"`
		} `json:"searchResult3"`
	} `json:"subsonic-response"`
}

const (
	subsonicAlbumSearchFmtStr  = "/rest/search3?%s&f=json&query=%s&v=1.13.0&c=koito&artistCount=0&songCount=0&albumCount=10"
	subsonicArtistSearchFmtStr = "/rest/search3?%s&f=json&query=%s&v=1.13.0&c=koito&artistCount=1&songCount=0&albumCount=0"
	subsonicCoverArtFmtStr     = "/rest/getCoverArt?%s&id=%s&v=1.13.0&c=koito"
)

func NewSubsonicClient() *SubsonicClient {
	ret := new(SubsonicClient)
	ret.url = cfg.SubsonicUrl()
	ret.userAgent = cfg.UserAgent()
	ret.authParams = cfg.SubsonicParams()
	ret.requestQueue = queue.NewRequestQueue(5, 5)
	return ret
}

func (c *SubsonicClient) queue(ctx context.Context, req *http.Request) ([]byte, error) {
	l := logger.FromContext(ctx)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resultChan := c.requestQueue.Enqueue(func(client *http.Client, done chan<- queue.RequestResult) {
		resp, err := client.Do(req)
		if err != nil {
			l.Debug().Err(err).Str("url", req.RequestURI).Msg("Failed to contact ImageSrc")
			done <- queue.RequestResult{Err: err}
			return
		} else if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			err = fmt.Errorf("recieved non-ok status from Subsonic: %s", resp.Status)
			done <- queue.RequestResult{Body: nil, Err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		done <- queue.RequestResult{Body: body, Err: err}
	})

	result := <-resultChan
	return result.Body, result.Err
}

func (c *SubsonicClient) getEntity(ctx context.Context, endpoint string, result any) error {
	l := logger.FromContext(ctx)
	url := c.url + endpoint
	l.Debug().Msgf("Sending request to ImageSrc: GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("getEntity: %w", err)
	}
	l.Debug().Msg("Adding ImageSrc request to queue")
	body, err := c.queue(ctx, req)
	if err != nil {
		l.Err(err).Msg("Subsonic request failed")
		return fmt.Errorf("getEntity: %w", err)
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		l.Err(err).Msg("Failed to unmarshal Subsonic response")
		return fmt.Errorf("getEntity: %w", err)
	}

	return nil
}

func (c *SubsonicClient) GetAlbumImage(ctx context.Context, mbid *uuid.UUID, artist, album string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(SubsonicAlbumResponse)
	l.Debug().Msgf("Finding album image for %s from artist %s", album, artist)
	// first try mbid search
	if mbid != nil {
		l.Debug().Str("mbid", mbid.String()).Msg("Searching album image by MBID")
		err := c.getEntity(ctx, fmt.Sprintf(subsonicAlbumSearchFmtStr, c.authParams, url.QueryEscape(mbid.String())), resp)
		if err != nil {
			return "", fmt.Errorf("GetAlbumImage: %v", err)
		}
		l.Debug().Any("subsonic_response", resp).Msg("")
		if len(resp.SubsonicResponse.SearchResult3.Album) >= 1 {
			return cfg.SubsonicUrl() + fmt.Sprintf(subsonicCoverArtFmtStr, c.authParams, url.QueryEscape(resp.SubsonicResponse.SearchResult3.Album[0].CoverArt)), nil
		}
	}
	// else do artist match
	l.Debug().Str("title", album).Str("artist", artist).Msg("Searching album image by title and artist")
	err := c.getEntity(ctx, fmt.Sprintf(subsonicAlbumSearchFmtStr, c.authParams, url.QueryEscape(album)), resp)
	if err != nil {
		return "", fmt.Errorf("GetAlbumImage: %v", err)
	}
	l.Debug().Any("subsonic_response", resp).Msg("")
	if len(resp.SubsonicResponse.SearchResult3.Album) < 1 {
		return "", fmt.Errorf("GetAlbumImage: failed to get album art from subsonic")
	}
	for _, album := range resp.SubsonicResponse.SearchResult3.Album {
		if album.Artist == artist {
			return cfg.SubsonicUrl() + fmt.Sprintf(subsonicCoverArtFmtStr, c.authParams, url.QueryEscape(resp.SubsonicResponse.SearchResult3.Album[0].CoverArt)), nil
		}
	}
	return "", fmt.Errorf("GetAlbumImage: failed to get album art from subsonic")
}

func (c *SubsonicClient) GetArtistImage(ctx context.Context, mbid *uuid.UUID, artist string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(SubsonicArtistResponse)
	l.Debug().Msgf("Finding artist image for %s", artist)
	// first try mbid search
	if mbid != nil {
		l.Debug().Str("mbid", mbid.String()).Msg("Searching artist image by MBID")
		err := c.getEntity(ctx, fmt.Sprintf(subsonicArtistSearchFmtStr, c.authParams, url.QueryEscape(mbid.String())), resp)
		if err != nil {
			return "", fmt.Errorf("GetArtistImage: %v", err)
		}
		l.Debug().Any("subsonic_response", resp).Msg("")
		if len(resp.SubsonicResponse.SearchResult3.Artist) < 1 || resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl == "" {
			return "", fmt.Errorf("GetArtistImage: failed to get artist art")
		}
		// Subsonic seems to have a tendency to return an artist image even though the url is a 404
		if err = ValidateImageURL(resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl); err != nil {
			return "", fmt.Errorf("GetArtistImage: failed to get validate image url")
		}
	}
	l.Debug().Str("artist", artist).Msg("Searching artist image by name")
	err := c.getEntity(ctx, fmt.Sprintf(subsonicArtistSearchFmtStr, c.authParams, url.QueryEscape(artist)), resp)
	if err != nil {
		return "", fmt.Errorf("GetArtistImage: %v", err)
	}
	l.Debug().Any("subsonic_response", resp).Msg("")
	if len(resp.SubsonicResponse.SearchResult3.Artist) < 1 || resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl == "" {
		return "", fmt.Errorf("GetArtistImage: failed to get artist art")
	}
	// Subsonic seems to have a tendency to return an artist image even though the url is a 404
	if err = ValidateImageURL(resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl); err != nil {
		return "", fmt.Errorf("GetArtistImage: failed to get validate image url")
	}
	return resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl, nil
}
