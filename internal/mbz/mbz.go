// package mbz provides functions for interacting with the musicbrainz api
package mbz

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/queue"
	"github.com/google/uuid"
)

type MusicBrainzArea struct {
	Name           string   `json:"name"`
	Iso3166_1Codes []string `json:"iso-3166-1-codes"`
}

type MusicBrainzClient struct {
	url          string
	userAgent    string
	requestQueue *queue.RequestQueue
}

type MusicBrainzCaller interface {
	GetArtistPrimaryAliases(ctx context.Context, id uuid.UUID) ([]string, error)
	GetReleaseTitles(ctx context.Context, RGID uuid.UUID) ([]string, error)
	GetTrack(ctx context.Context, id uuid.UUID) (*MusicBrainzTrack, error)
	GetReleaseGroup(ctx context.Context, id uuid.UUID) (*MusicBrainzReleaseGroup, error)
	GetRelease(ctx context.Context, id uuid.UUID) (*MusicBrainzRelease, error)
	SearchRecording(ctx context.Context, artist string, track string) (*MusicBrainzSearchResult, error)
	Shutdown()
}

func NewMusicBrainzClient() *MusicBrainzClient {
	ret := new(MusicBrainzClient)
	ret.url = cfg.MusicBrainzUrl()
	ret.userAgent = cfg.UserAgent()
	ret.requestQueue = queue.NewRequestQueue(cfg.MusicBrainzRateLimit(), cfg.MusicBrainzRateLimit())
	return ret
}

func (c *MusicBrainzClient) Shutdown() {
	c.requestQueue.Shutdown()
}

func (c *MusicBrainzClient) getEntity(ctx context.Context, fmtStr string, id uuid.UUID, result any) error {
	l := logger.FromContext(ctx)
	url := fmt.Sprintf(fmtStr, c.url, id.String())
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		l.Err(err).Msg("Failed to build MusicBrainz request")
		return fmt.Errorf("getEntity: %w", err)
	}
	l.Debug().Msg("Adding MusicBrainz request to queue")
	body, err := c.queue(ctx, req)
	if err != nil {
		l.Err(err).Msg("MusicBrainz request failed")
		return fmt.Errorf("getEntity: %w", err)
	}

	err = json.Unmarshal(body, result)
	if err != nil {
		l.Err(err).Str("body", string(body)).Msg("Failed to unmarshal MusicBrainz response body")
		return fmt.Errorf("getEntity: %w", err)
	}

	return nil
}

func (c *MusicBrainzClient) queue(ctx context.Context, req *http.Request) ([]byte, error) {
	l := logger.FromContext(ctx)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resultChan := c.requestQueue.Enqueue(func(client *http.Client, done chan<- queue.RequestResult) {
		resp, err := client.Do(req)
		if err != nil {
			l.Err(err).Str("url", req.RequestURI).Msg("Failed to contact MusicBrainz")
			done <- queue.RequestResult{Err: err}
			return
		} else if resp.StatusCode >= 300 || resp.StatusCode < 200 {
			err = fmt.Errorf("recieved non-ok status from MusicBrainz: %s", resp.Status)
			done <- queue.RequestResult{Body: nil, Err: err}
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		done <- queue.RequestResult{Body: body, Err: err}
	})

	result := <-resultChan
	return result.Body, result.Err
}
