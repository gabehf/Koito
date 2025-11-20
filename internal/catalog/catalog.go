// Package catalog manages the internal metadata of the catalog of music the user has submitted listens for.
// This includes artists, releases (album, single, ep, etc), and tracks, as well as ingesting
// listens submitted both via the API(s) and other methods.
package catalog

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/memkv"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

type GetListensOpts struct {
	ArtistID       int32
	ReleaseGroupID int32
	TrackID        int32
	Limit          int
}

type SaveListenOpts struct {
	TrackID int32
	Time    time.Time
}

type ArtistMbidMap struct {
	Artist string
	Mbid   uuid.UUID
}

type SubmitListenOpts struct {
	// When true, skips registering the listen and only associates or creates the
	// artist, release, release group, and track in DB
	SkipSaveListen bool

	// When true, skips caching the images and only stores the image url in the db
	SkipCacheImage bool

	MbzCaller          mbz.MusicBrainzCaller
	ArtistNames        []string
	Artist             string
	ArtistMbzIDs       []uuid.UUID
	ArtistMbidMappings []ArtistMbidMap
	TrackTitle         string
	RecordingMbzID     uuid.UUID
	Duration           int32 // in seconds
	ReleaseTitle       string
	ReleaseMbzID       uuid.UUID
	ReleaseGroupMbzID  uuid.UUID
	Time               time.Time

	UserID       int32
	Client       string
	IsNowPlaying bool
}

const (
	ImageSourceUserUpload = "User Upload"
)

func SubmitListen(ctx context.Context, store db.DB, opts SubmitListenOpts) error {
	l := logger.FromContext(ctx)

	if opts.Artist == "" || opts.TrackTitle == "" {
		return errors.New("track name and artist are required")
	}

	// bandaid to ensure new activity does not have sub-second precision
	opts.Time = opts.Time.Truncate(time.Second)

	artists, err := AssociateArtists(
		ctx,
		store,
		AssociateArtistsOpts{
			ArtistMbzIDs:   opts.ArtistMbzIDs,
			ArtistNames:    opts.ArtistNames,
			ArtistName:     opts.Artist,
			ArtistMbidMap:  opts.ArtistMbidMappings,
			Mbzc:           opts.MbzCaller,
			TrackTitle:     opts.TrackTitle,
			SkipCacheImage: opts.SkipCacheImage,
		})
	if err != nil {
		l.Err(err).Msg("Failed to associate artists to listen")
		return fmt.Errorf("SubmitListen: %w", err)
	} else if len(artists) < 1 {
		l.Debug().Msg("Failed to associate any artists to release")
	}

	artistIDs := make([]int32, len(artists))

	for i, artist := range artists {
		artistIDs[i] = artist.ID
		l.Debug().Any("artist", artist).Msg("Matched listen to artist")
	}
	rg, err := AssociateAlbum(ctx, store, AssociateAlbumOpts{
		ReleaseMbzID:      opts.ReleaseMbzID,
		ReleaseGroupMbzID: opts.ReleaseGroupMbzID,
		ReleaseName:       opts.ReleaseTitle,
		TrackName:         opts.TrackTitle,
		Mbzc:              opts.MbzCaller,
		Artists:           artists,
		SkipCacheImage:    opts.SkipCacheImage,
	})
	if err != nil {
		l.Error().Err(err).Msg("Failed to associate release group to listen")
		return fmt.Errorf("SubmitListen: %w", err)
	}
	l.Debug().Any("album", rg).Msg("Matched listen to release")

	// ensure artists are associated with release group
	store.AddArtistsToAlbum(ctx, db.AddArtistsToAlbumOpts{
		ArtistIDs: artistIDs,
		AlbumID:   rg.ID,
	})

	track, err := AssociateTrack(ctx, store, AssociateTrackOpts{
		ArtistIDs:  artistIDs,
		AlbumID:    rg.ID,
		TrackMbzID: opts.RecordingMbzID,
		TrackName:  opts.TrackTitle,
		Duration:   opts.Duration,
		Mbzc:       opts.MbzCaller,
	})
	if err != nil {
		l.Error().Err(err).Msg("Failed to associate track to listen")
		return fmt.Errorf("SubmitListen: %w", err)
	}
	l.Debug().Any("track", track).Msg("Matched listen to track")

	if track.Duration == 0 {
		if opts.Duration != 0 {
			l.Debug().Msg("Updating duration using request information")
			err := store.UpdateTrack(ctx, db.UpdateTrackOpts{
				ID:       track.ID,
				Duration: opts.Duration,
			})
			if err != nil {
				l.Err(err).Msgf("Failed to update duration for track %s", track.Title)
			} else {
				l.Info().Msgf("Duration updated to %d for track '%s'", opts.Duration, track.Title)
			}
		} else if track.MbzID != nil && *track.MbzID != uuid.Nil {
			l.Debug().Msg("Attempting to update duration using MusicBrainz ID")
			mbztrack, err := opts.MbzCaller.GetTrack(ctx, *track.MbzID)
			if err != nil {
				l.Err(err).Msg("Failed to make request to MusicBrainz")
			} else {
				err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
					ID:       track.ID,
					Duration: int32(mbztrack.LengthMs / 1000),
				})
				if err != nil {
					l.Err(err).Msgf("Failed to update duration for track %s", track.Title)
				} else {
					l.Info().Msgf("Duration updated to %d for track '%s'", mbztrack.LengthMs/1000, track.Title)
				}
			}
		}
	}

	if opts.IsNowPlaying {
		if track.Duration == 0 {
			memkv.Store.Set(strconv.Itoa(int(opts.UserID)), track.ID)
		} else {
			memkv.Store.Set(strconv.Itoa(int(opts.UserID)), track.ID, time.Duration(track.Duration)*time.Second)
		}
	}

	if opts.SkipSaveListen {
		return nil
	}

	l.Info().Msgf("Received listen: '%s' by %s, from release '%s'", track.Title, buildArtistStr(artists), rg.Title)

	return store.SaveListen(ctx, db.SaveListenOpts{
		TrackID: track.ID,
		Time:    opts.Time,
		UserID:  opts.UserID,
		Client:  opts.Client,
	})
}

func buildArtistStr(artists []*models.Artist) string {
	artistNames := make([]string, len(artists))
	for i, artist := range artists {
		artistNames[i] = artist.Name
	}
	return strings.Join(artistNames, " & ")
}

var (
	// Bracketed feat patterns
	bracketFeatPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)\([fF]eat\. ([^)]*)\)`),
		regexp.MustCompile(`(?i)\[[fF]eat\. ([^\]]*)\]`),
	}
	// Inline feat (not in brackets)
	inlineFeatPattern = regexp.MustCompile(`(?i)[fF]eat\. ([^()\[\]]+)$`)

	// Delimiters only used inside feat. sections
	featSplitDelimiters = regexp.MustCompile(`(?i)\s*(?:,|&|and|·)\s*`)
)

// ParseArtists extracts all contributing artist names from the artist and title strings
func ParseArtists(artist string, title string, addlSeparators []*regexp.Regexp) []string {
	seen := make(map[string]struct{})
	var out []string

	add := func(name string) {
		name = strings.TrimSpace(name)
		if name == "" {
			return
		}
		if _, exists := seen[name]; !exists {
			seen[name] = struct{}{}
			out = append(out, name)
		}
	}

	// Extract bracketed features from artist
	for _, re := range bracketFeatPatterns {
		if matches := re.FindStringSubmatch(artist); matches != nil {
			artist = strings.Replace(artist, matches[0], "", 1)
			for _, name := range featSplitDelimiters.Split(matches[1], -1) {
				add(name)
			}
		}
	}
	// Extract inline feat. from artist
	if matches := inlineFeatPattern.FindStringSubmatch(artist); matches != nil {
		artist = strings.Replace(artist, matches[0], "", 1)
		for _, name := range featSplitDelimiters.Split(matches[1], -1) {
			add(name)
		}
	}

	// Add base artist(s)
	l1 := len(out)
	for _, re := range addlSeparators {
		for _, name := range re.Split(artist, -1) {
			if name == artist {
				continue
			}
			add(name)
		}
	}
	// Only add the full artist string if no splitters were matched
	if l1 == len(out) {
		add(artist)
	}

	// Extract features from title
	for _, re := range bracketFeatPatterns {
		if matches := re.FindStringSubmatch(title); matches != nil {
			for _, name := range featSplitDelimiters.Split(matches[1], -1) {
				add(name)
			}
		}
	}
	if matches := inlineFeatPattern.FindStringSubmatch(title); matches != nil {
		for _, name := range featSplitDelimiters.Split(matches[1], -1) {
			add(name)
		}
	}

	return out
}
