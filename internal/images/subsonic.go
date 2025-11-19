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
	subsonicAlbumSearchFmtStr  = "/rest/search3?%s&f=json&query=%s&v=1.13.0&c=koito&artistCount=0&songCount=0&albumCount=1"
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

func (c *SubsonicClient) GetAlbumImage(ctx context.Context, artist, album string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(SubsonicAlbumResponse)
	l.Debug().Msgf("Finding album image for %s from artist %s", album, artist)
	err := c.getEntity(ctx, fmt.Sprintf(subsonicAlbumSearchFmtStr, c.authParams, url.QueryEscape(artist+" "+album)), resp)
	if err != nil {
		return "", fmt.Errorf("GetAlbumImage: %v", err)
	}
	l.Debug().Any("subsonic_response", resp).Send()
	if len(resp.SubsonicResponse.SearchResult3.Album) < 1 || resp.SubsonicResponse.SearchResult3.Album[0].CoverArt == "" {
		return "", fmt.Errorf("GetAlbumImage: failed to get album art")
	}
	return cfg.SubsonicUrl() + fmt.Sprintf(subsonicCoverArtFmtStr, c.authParams, url.QueryEscape(resp.SubsonicResponse.SearchResult3.Album[0].CoverArt)), nil
}

func (c *SubsonicClient) GetArtistImage(ctx context.Context, artist string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(SubsonicArtistResponse)
	l.Debug().Msgf("Finding artist image for %s", artist)
	err := c.getEntity(ctx, fmt.Sprintf(subsonicArtistSearchFmtStr, c.authParams, url.QueryEscape(artist)), resp)
	if err != nil {
		return "", fmt.Errorf("GetArtistImage: %v", err)
	}
	l.Debug().Any("subsonic_response", resp).Send()
	if len(resp.SubsonicResponse.SearchResult3.Artist) < 1 || resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl == "" {
		return "", fmt.Errorf("GetArtistImage: failed to get artist art")
	}
	return resp.SubsonicResponse.SearchResult3.Artist[0].ArtistImageUrl, nil
}
