// package imagesrc defines interfaces for album and artist image providers
package images

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gabehf/koito/internal/logger"
	"github.com/google/uuid"
)

type ImageSource struct {
	deezerEnabled   bool
	deezerC         *DeezerClient
	subsonicEnabled bool
	subsonicC       *SubsonicClient
	caaEnabled      bool
}
type ImageSourceOpts struct {
	UserAgent      string
	EnableCAA      bool
	EnableDeezer   bool
	EnableSubsonic bool
}

var once sync.Once
var imgsrc ImageSource

type ArtistImageOpts struct {
	Aliases []string
}

type AlbumImageOpts struct {
	Artists           []string
	Album             string
	ReleaseMbzID      *uuid.UUID
	ReleaseGroupMbzID *uuid.UUID
}

const caaBaseUrl = "https://coverartarchive.org"

// all functions are no-op if no providers are enabled
func Initialize(opts ImageSourceOpts) {
	once.Do(func() {
		if opts.EnableCAA {
			imgsrc.caaEnabled = true
		}
		if opts.EnableDeezer {
			imgsrc.deezerEnabled = true
			imgsrc.deezerC = NewDeezerClient()
		}
		if opts.EnableSubsonic {
			imgsrc.subsonicEnabled = true
			imgsrc.subsonicC = NewSubsonicClient()
		}
	})
}

func Shutdown() {
	imgsrc.deezerC.Shutdown()
}

func GetArtistImage(ctx context.Context, opts ArtistImageOpts) (string, error) {
	l := logger.FromContext(ctx)
	if imgsrc.subsonicEnabled {
		img, err := imgsrc.subsonicC.GetArtistImage(ctx, opts.Aliases[0])
		if err != nil {
			l.Debug().Err(err).Msg("GetArtistImage: Could not find artist image from Subsonic")
		} else if img != "" {
			return img, nil
		}
	} else {
		l.Debug().Msg("GetArtistImage: Subsonic image fetching is disabled")
	}
	if imgsrc.deezerEnabled {
		img, err := imgsrc.deezerC.GetArtistImages(ctx, opts.Aliases)
		if err != nil {
			l.Debug().Err(err).Msg("GetArtistImage: Could not find artist image from Deezer")
			return "", err
		} else if img != "" {
			return img, nil
		}
	} else {
		l.Debug().Msg("GetArtistImage: Deezer image fetching is disabled")
	}
	l.Warn().Msg("GetArtistImage: No image providers are enabled")
	return "", nil
}
func GetAlbumImage(ctx context.Context, opts AlbumImageOpts) (string, error) {
	l := logger.FromContext(ctx)
	if imgsrc.subsonicEnabled {
		img, err := imgsrc.subsonicC.GetAlbumImage(ctx, opts.Artists[0], opts.Album)
		if err != nil {
			l.Debug().Err(err).Msg("GetAlbumImage: Could not find artist image from Subsonic")
		}
		if img != "" {
			return img, nil
		}
		l.Debug().Msg("Could not find album cover from Subsonic")
	}
	if imgsrc.caaEnabled {
		l.Debug().Msg("Attempting to find album image from CoverArtArchive")
		if opts.ReleaseMbzID != nil && *opts.ReleaseMbzID != uuid.Nil {
			url := fmt.Sprintf(caaBaseUrl+"/release/%s/front", opts.ReleaseMbzID.String())
			resp, err := http.DefaultClient.Head(url)
			if err != nil {
				l.Debug().Err(err).Msg("GetAlbumImage: Could not find artist image from CoverArtArchive with Release MBID")
			}
			if resp.StatusCode == 200 {
				return url, nil
			}
		}
		if opts.ReleaseGroupMbzID != nil && *opts.ReleaseGroupMbzID != uuid.Nil {
			url := fmt.Sprintf(caaBaseUrl+"/release-group/%s/front", opts.ReleaseGroupMbzID.String())
			resp, err := http.DefaultClient.Head(url)
			if err != nil {
				l.Debug().Err(err).Msg("GetAlbumImage: Could not find artist image from CoverArtArchive with Release Group MBID")
			}
			if resp.StatusCode == 200 {
				return url, nil
			}
		}
	}
	if imgsrc.deezerEnabled {
		l.Debug().Msg("Attempting to find album image from Deezer")
		img, err := imgsrc.deezerC.GetAlbumImages(ctx, opts.Artists, opts.Album)
		if err != nil {
			l.Debug().Err(err).Msg("GetAlbumImage: Could not find artist image from Deezer")
			return "", err
		}
		return img, nil
	}
	l.Warn().Msg("GetAlbumImage: No image providers are enabled")
	return "", nil
}

// ValidateImageURL checks if the URL points to a valid image by performing a HEAD request.
func ValidateImageURL(url string) error {
	resp, err := http.Head(url)
	if err != nil {
		return fmt.Errorf("ValidateImageURL: http.Head: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ValidateImageURL: HEAD request failed, status code: %d", resp.StatusCode)
	}

	contentType := resp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return fmt.Errorf("ValidateImageURL: URL does not point to an image, content type: %s", contentType)
	}

	return nil
}
