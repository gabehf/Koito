package images

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/utils"
	"github.com/gabehf/koito/queue"
)

type DeezerClient struct {
	url          string
	userAgent    string
	requestQueue *queue.RequestQueue
}

type DeezerAlbumResponse struct {
	Data []DeezerAlbum `json:"data"`
}
type DeezerAlbum struct {
	Title    string `json:"title"`
	CoverXL  string `json:"cover_xl"`
	CoverSm  string `json:"cover_small"`
	CoverMd  string `json:"cover_medium"`
	CoverBig string `json:"cover_big"`
}
type DeezerArtistResponse struct {
	Data []DeezerArtist `json:"data"`
}
type DeezerArtist struct {
	Name       string `json:"name"`
	PictureXL  string `json:"picture_xl"`
	PictureSm  string `json:"picture_small"`
	PictureMd  string `json:"picture_medium"`
	PictureBig string `json:"picture_big"`
}

const (
	deezerBaseUrl       = "https://api.deezer.com"
	albumImageEndpoint  = "/search/album?q=%s"
	artistImageEndpoint = "/search/artist?q=%s"
)

func NewDeezerClient() *DeezerClient {
	ret := new(DeezerClient)
	ret.url = deezerBaseUrl
	ret.userAgent = cfg.UserAgent()
	ret.requestQueue = queue.NewRequestQueue(5, 5)
	return ret
}

func (c *DeezerClient) Shutdown() {
	c.requestQueue.Shutdown()
}

func (c *DeezerClient) queue(ctx context.Context, req *http.Request) ([]byte, error) {
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
			err = fmt.Errorf("recieved non-ok status from Deezer: %s", resp.Status)
			done <- queue.RequestResult{Body: nil, Err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		done <- queue.RequestResult{Body: body, Err: err}
	})

	result := <-resultChan
	return result.Body, result.Err
}

func (c *DeezerClient) getEntity(ctx context.Context, endpoint string, result any) error {
	l := logger.FromContext(ctx)
	url := deezerBaseUrl + endpoint
	l.Debug().Msgf("Sending request to ImageSrc: GET %s", url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("getEntity: %w", err)
	}
	l.Debug().Msg("Adding ImageSrc request to queue")
	body, err := c.queue(ctx, req)
	if err != nil {
		l.Err(err).Msg("Deezer request failed")
		return fmt.Errorf("getEntity: %w", err)
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		l.Err(err).Msg("Failed to unmarshal Deezer response")
		return fmt.Errorf("getEntity: %w", err)
	}

	return nil
}

// Deezer behavior is that it serves a default image when it can't find one for an artist, so
// this function will just download the default image thinking that it is an actual artist image.
// I don't know how to fix this yet.
func (c *DeezerClient) GetArtistImages(ctx context.Context, aliases []string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(DeezerArtistResponse)

	aliasesUniq := utils.UniqueIgnoringCase(aliases)
	aliasesAscii := utils.RemoveNonAscii(aliasesUniq)

	// Deezer very often uses romanized names for foreign artists, so check those first
	for _, a := range aliasesAscii {
		err := c.getEntity(ctx, fmt.Sprintf(artistImageEndpoint, url.QueryEscape(fmt.Sprintf("artist:\"%s\"", a))), resp)
		if err != nil {
			return "", fmt.Errorf("GetArtistImages: %w", err)
		}
		if len(resp.Data) < 1 {
			return "", errors.New("GetArtistImages: artist image not found")
		}
		for _, v := range resp.Data {
			if strings.EqualFold(v.Name, a) {
				img := v.PictureXL
				l.Debug().Msgf("Found artist images for %s: %v", a, img)
				return img, nil
			}
		}
	}

	// if no romanized name exists or couldn't be found, check the rest
	for _, a := range utils.RemoveInBoth(aliasesUniq, aliasesAscii) {
		err := c.getEntity(ctx, fmt.Sprintf(artistImageEndpoint, url.QueryEscape(fmt.Sprintf("artist:\"%s\"", a))), resp)
		if err != nil {
			return "", fmt.Errorf("GetArtistImages: %w", err)
		}
		if len(resp.Data) < 1 {
			return "", errors.New("GetArtistImages: artist image not found")
		}
		for _, v := range resp.Data {
			if strings.EqualFold(v.Name, a) {
				img := v.PictureXL
				l.Debug().Msgf("Found artist images for %s: %v", a, img)
				return img, nil
			}
		}
	}
	return "", errors.New("GetArtistImages: artist image not found")
}

func (c *DeezerClient) GetAlbumImages(ctx context.Context, artists []string, album string) (string, error) {
	l := logger.FromContext(ctx)
	resp := new(DeezerAlbumResponse)
	l.Debug().Msgf("Finding album image for %s from artist(s) %v", album, artists)
	// try to find artist + album match for all artists
	for _, alias := range artists {
		err := c.getEntity(ctx, fmt.Sprintf(albumImageEndpoint, url.QueryEscape(fmt.Sprintf("artist:\"%s\"album:\"%s\"", alias, album))), resp)
		if err != nil {
			return "", fmt.Errorf("GetAlbumImages: %w", err)
		}
		if len(resp.Data) > 0 {
			for _, v := range resp.Data {
				if strings.EqualFold(v.Title, album) {
					img := v.CoverXL
					l.Debug().Msgf("Found album images for %s: %v", album, img)
					return img, nil
				}
			}
		}
	}

	// if none are found, try to find an album just by album title
	err := c.getEntity(ctx, fmt.Sprintf(albumImageEndpoint, url.QueryEscape(fmt.Sprintf("album:\"%s\"", album))), resp)
	if err != nil {
		return "", fmt.Errorf("GetAlbumImages: %w", err)
	}
	for _, v := range resp.Data {
		if strings.EqualFold(v.Title, album) {
			img := v.CoverXL
			l.Debug().Msgf("Found album images for %s: %v", album, img)
			return img, nil
		}
	}

	return "", errors.New("GetAlbumImages: album image not found")
}
