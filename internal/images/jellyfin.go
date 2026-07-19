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

type JellyfinClient struct {
	url          string
	apiKey       string
	userID       string
	userAgent    string
	requestQueue *queue.RequestQueue
}

type jellyfinItemsResponse struct {
	Items []jellyfinItem `json:"Items"`
}

type jellyfinItem struct {
	ID          string            `json:"Id"`
	Name        string            `json:"Name"`
	AlbumArtist string            `json:"AlbumArtist"`
	ImageTags   map[string]string `json:"ImageTags"`
	ProviderIds map[string]string `json:"ProviderIds"`
}

func NewJellyfinClient() *JellyfinClient {
	return &JellyfinClient{
		url:          cfg.JellyfinUrl(),
		apiKey:       cfg.JellyfinApiKey(),
		userID:       cfg.JellyfinUserID(),
		userAgent:    cfg.UserAgent(),
		requestQueue: queue.NewRequestQueue(5, 5),
	}
}

func (c *JellyfinClient) queue(ctx context.Context, req *http.Request) ([]byte, error) {
	l := logger.FromContext(ctx)
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-Emby-Token", c.apiKey)

	resultChan := c.requestQueue.Enqueue(func(client *http.Client, done chan<- queue.RequestResult) {
		resp, err := client.Do(req)
		if err != nil {
			l.Debug().Err(err).Str("url", req.URL.String()).Msg("Failed to contact Jellyfin")
			done <- queue.RequestResult{Err: err}
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			done <- queue.RequestResult{Err: fmt.Errorf("received non-ok status from Jellyfin: %s", resp.Status)}
			return
		}
		body, err := io.ReadAll(resp.Body)
		done <- queue.RequestResult{Body: body, Err: err}
	})

	result := <-resultChan
	return result.Body, result.Err
}

func (c *JellyfinClient) itemsPath() string {
	if c.userID != "" {
		return "/Users/" + url.PathEscape(c.userID) + "/Items"
	}
	return "/Items"
}

func (c *JellyfinClient) searchItems(ctx context.Context, itemType, searchTerm string) ([]jellyfinItem, error) {
	l := logger.FromContext(ctx)
	q := url.Values{}
	q.Set("Recursive", "true")
	q.Set("IncludeItemTypes", itemType)
	q.Set("SearchTerm", searchTerm)
	q.Set("Fields", "ProviderIds")
	q.Set("EnableImageTypes", "Primary")
	q.Set("ImageTypeLimit", "1")
	q.Set("Limit", "25")

	endpoint := c.url + c.itemsPath() + "?" + q.Encode()
	l.Debug().Msgf("Sending request to Jellyfin: GET %s", endpoint)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("searchItems: %w", err)
	}
	body, err := c.queue(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("searchItems: %w", err)
	}
	var resp jellyfinItemsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("searchItems: %w", err)
	}
	return resp.Items, nil
}

func (c *JellyfinClient) primaryImageURL(item jellyfinItem) (string, error) {
	if item.ID == "" {
		return "", fmt.Errorf("primaryImageURL: missing item id")
	}
	if len(item.ImageTags) > 0 {
		if _, ok := item.ImageTags["Primary"]; !ok {
			return "", fmt.Errorf("primaryImageURL: item has no primary image")
		}
	}
	// api_key query param so imagecache can fetch without custom headers
	return fmt.Sprintf("%s/Items/%s/Images/Primary?api_key=%s",
		c.url, url.PathEscape(item.ID), url.QueryEscape(c.apiKey)), nil
}

func providerIDEquals(ids map[string]string, keys []string, want string) bool {
	if ids == nil || want == "" {
		return false
	}
	want = strings.ToLower(want)
	for _, key := range keys {
		if v, ok := ids[key]; ok && strings.EqualFold(v, want) {
			return true
		}
	}
	return false
}

func (c *JellyfinClient) pickAlbumImage(items []jellyfinItem, mbid *uuid.UUID, artist string) (string, bool) {
	mbidKeys := []string{"MusicBrainzAlbum", "MusicBrainzReleaseGroup"}
	if mbid != nil && *mbid != uuid.Nil {
		for _, item := range items {
			if providerIDEquals(item.ProviderIds, mbidKeys, mbid.String()) {
				if img, err := c.primaryImageURL(item); err == nil {
					return img, true
				}
			}
		}
	}
	for _, item := range items {
		if artist != "" && !strings.EqualFold(item.AlbumArtist, artist) {
			continue
		}
		if img, err := c.primaryImageURL(item); err == nil {
			return img, true
		}
	}
	for _, item := range items {
		if img, err := c.primaryImageURL(item); err == nil {
			return img, true
		}
	}
	return "", false
}

func (c *JellyfinClient) GetAlbumImage(ctx context.Context, mbid *uuid.UUID, artist, album string) (string, error) {
	l := logger.FromContext(ctx)
	l.Debug().Msgf("Finding album image for %s from artist %s via Jellyfin", album, artist)

	if album != "" {
		l.Debug().Str("title", album).Str("artist", artist).Msg("Searching Jellyfin album image by title")
		items, err := c.searchItems(ctx, "MusicAlbum", album)
		if err != nil {
			return "", fmt.Errorf("GetAlbumImage: %w", err)
		}
		if img, ok := c.pickAlbumImage(items, mbid, artist); ok {
			return img, nil
		}
	}

	if mbid != nil && *mbid != uuid.Nil {
		l.Debug().Str("mbid", mbid.String()).Msg("Searching Jellyfin album image by MBID")
		items, err := c.searchItems(ctx, "MusicAlbum", mbid.String())
		if err != nil {
			return "", fmt.Errorf("GetAlbumImage: %w", err)
		}
		if img, ok := c.pickAlbumImage(items, mbid, artist); ok {
			return img, nil
		}
	}

	return "", fmt.Errorf("GetAlbumImage: failed to get album art from jellyfin")
}

func (c *JellyfinClient) GetArtistImage(ctx context.Context, mbid *uuid.UUID, artist string) (string, error) {
	l := logger.FromContext(ctx)
	l.Debug().Msgf("Finding artist image for %s via Jellyfin", artist)

	mbidKeys := []string{"MusicBrainzArtist"}

	if mbid != nil && *mbid != uuid.Nil {
		l.Debug().Str("mbid", mbid.String()).Msg("Searching Jellyfin artist image by MBID")
		items, err := c.searchItems(ctx, "MusicArtist", mbid.String())
		if err != nil {
			return "", fmt.Errorf("GetArtistImage: %w", err)
		}
		for _, item := range items {
			if providerIDEquals(item.ProviderIds, mbidKeys, mbid.String()) {
				img, err := c.primaryImageURL(item)
				if err != nil {
					continue
				}
				return img, nil
			}
		}
	}

	if artist == "" {
		return "", fmt.Errorf("GetArtistImage: failed to get artist art from jellyfin")
	}

	l.Debug().Str("artist", artist).Msg("Searching Jellyfin artist image by name")
	items, err := c.searchItems(ctx, "MusicArtist", artist)
	if err != nil {
		return "", fmt.Errorf("GetArtistImage: %w", err)
	}
	if len(items) < 1 {
		return "", fmt.Errorf("GetArtistImage: failed to get artist art from jellyfin")
	}

	if mbid != nil && *mbid != uuid.Nil {
		for _, item := range items {
			if providerIDEquals(item.ProviderIds, mbidKeys, mbid.String()) {
				img, err := c.primaryImageURL(item)
				if err != nil {
					continue
				}
				return img, nil
			}
		}
	}

	for _, item := range items {
		if !strings.EqualFold(item.Name, artist) {
			continue
		}
		img, err := c.primaryImageURL(item)
		if err != nil {
			continue
		}
		return img, nil
	}

	for _, item := range items {
		img, err := c.primaryImageURL(item)
		if err != nil {
			continue
		}
		return img, nil
	}

	return "", fmt.Errorf("GetArtistImage: failed to get artist art from jellyfin")
}
