package catalog

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/gabehf/koito/imagecache"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
)

type AssociateArtistsOpts struct {
	ArtistMbzIDs  []uuid.UUID
	ArtistNames   []string
	ArtistMbidMap []ArtistMbidMap
	ArtistName    string
	TrackTitle    string
	Mbzc          mbz.MusicBrainzCaller

	SkipCacheImage bool
	OriginUrl      string // e.g. the source video URL; used as a last-resort artist image fallback
}

func AssociateArtists(ctx context.Context, d db.ArtistStore, opts AssociateArtistsOpts) ([]*models.Artist, error) {
	l := logger.FromContext(ctx)

	var result []*models.Artist

	// use mbid map first, as it is the most reliable way to get mbid for artists
	if len(opts.ArtistMbidMap) > 0 {
		l.Debug().Msg("Associating artists by MusicBrainz ID(s) mappings")
		mbzMatches, err := matchArtistsByMBIDMappings(ctx, d, opts)
		if err != nil {
			return nil, fmt.Errorf("AssociateArtists: %w", err)
		}
		result = append(result, mbzMatches...)
	}

	if len(opts.ArtistMbzIDs) > len(result) {
		l.Debug().Msg("Associating artists by list of MusicBrainz ID(s)")
		mbzMatches, err := matchArtistsByMBID(ctx, d, opts, result)
		if err != nil {
			return nil, fmt.Errorf("AssociateArtists: %w", err)
		}
		result = append(result, mbzMatches...)
	}

	if len(opts.ArtistNames) > len(result) {
		l.Debug().Msg("Associating artists by list of artist names")
		nameMatches, err := matchArtistsByNames(ctx, opts.ArtistNames, result, d, opts)
		if err != nil {
			return nil, fmt.Errorf("AssociateArtists: %w", err)
		}
		result = append(result, nameMatches...)
	}

	if len(result) < 1 {
		allArtists := slices.Concat(opts.ArtistNames, ParseArtists(opts.ArtistName, opts.TrackTitle, cfg.ArtistSeparators()))
		l.Debug().Msgf("Associating artists by artist name(s) %v and track title '%s'", allArtists, opts.TrackTitle)
		fallbackMatches, err := matchArtistsByNames(ctx, allArtists, nil, d, opts)
		if err != nil {
			return nil, fmt.Errorf("AssociateArtists: %w", err)
		}
		result = append(result, fallbackMatches...)
	}

	return result, nil
}

func matchArtistsByMBIDMappings(ctx context.Context, d db.ArtistStore, opts AssociateArtistsOpts) ([]*models.Artist, error) {
	var result []*models.Artist

	for _, a := range opts.ArtistMbidMap {
		artist, err := matchOrCreateArtistByMbidMapping(ctx, a, d, opts)
		if err != nil {
			return nil, fmt.Errorf("matchArtistsByMBIDMappings: %w", err)
		}
		result = append(result, artist)
	}

	return result, nil
}

func matchOrCreateArtistByMbidMapping(ctx context.Context, a ArtistMbidMap, d db.ArtistStore, opts AssociateArtistsOpts) (*models.Artist, error) {
	l := logger.FromContext(ctx)

	unlock := lockKey("artist:" + strings.ToLower(a.Artist))
	defer unlock()

	artist, err := d.GetArtist(ctx, db.GetArtistOpts{
		MusicBrainzID: a.Mbid,
	})
	if err == nil {
		l.Debug().Msgf("Artist '%s' found by MusicBrainz ID", artist.Name)
		return artist, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	artist, err = d.GetArtist(ctx, db.GetArtistOpts{
		Name: a.Artist,
	})
	if err == nil {
		l.Debug().Msgf("Artist '%s' found by Name", a.Artist)
		if artist.MbzID == nil {
			if err := d.UpdateArtist(ctx, db.UpdateArtistOpts{
				ID:            artist.ID,
				MusicBrainzID: a.Mbid,
			}); err != nil {
				l.Err(err).Msg("matchOrCreateArtistByMbidMapping: failed to update artist with MusicBrainz ID")
				return nil, err
			}
			artist.MbzID = &a.Mbid
			l.Debug().Msgf("Updated artist '%s' with MusicBrainz ID", artist.Name)
		} else {
			l.Warn().Msgf("Attempted to update artist %s with MusicBrainz ID, but an existing ID was already found", artist.Name)
		}
		return artist, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	artist, err = resolveAliasOrCreateArtist(ctx, a.Mbid, opts.ArtistNames, d, opts)
	if err != nil {
		l.Warn().AnErr("error", err).Msg("matchOrCreateArtistByMbidMapping: MusicBrainz unreachable, creating new artist with provided MusicBrainz ID mapping")

		var imgid uuid.UUID
		var imgUrl string
		if !opts.SkipCacheImage {
			var imgErr error
			imgUrl, imgErr = images.GetArtistImage(ctx, images.ArtistImageOpts{
				Aliases:   []string{a.Artist},
				OriginUrl: opts.OriginUrl,
			})
			if imgErr == nil && imgUrl != "" {
				imgid = uuid.New()
				l.Debug().Msg("Downloading artist image from source...")
				if dlErr := imagecache.DownloadImage(imgid, imgUrl); dlErr != nil {
					l.Err(dlErr).Msg("Failed to cache image")
				}
			} else {
				l.Err(imgErr).Msgf("matchOrCreateArtistByMbidMapping: Failed to get artist image for artist '%s'", a.Artist)
			}
		}

		artist, err = d.SaveArtist(ctx, db.SaveArtistOpts{
			Name:          a.Artist,
			MusicBrainzID: a.Mbid,
			Image:         imgid,
			ImageSrc:      imgUrl,
		})
		if err != nil {
			l.Err(err).Msgf("matchOrCreateArtistByMbidMapping: Failed to create artist '%s' in database", a.Artist)
			return nil, err
		}
	}

	return artist, nil
}

func matchArtistsByMBID(ctx context.Context, d db.ArtistStore, opts AssociateArtistsOpts, existing []*models.Artist) ([]*models.Artist, error) {
	l := logger.FromContext(ctx)
	var result []*models.Artist

	for _, id := range opts.ArtistMbzIDs {
		if artistExistsByMbzID(id, existing) || artistExistsByMbzID(id, result) {
			l.Debug().Msgf("Artist with MusicBrainz ID %s already found, skipping...", id)
			continue
		}
		if id == uuid.Nil {
			l.Warn().Msg("Provided artist has uuid.Nil MusicBrainzID")
			return matchArtistsByNames(ctx, opts.ArtistNames, result, d, opts)
		}

		if len(opts.ArtistNames) < 1 {
			opts.ArtistNames = slices.Concat(opts.ArtistNames, ParseArtists(opts.ArtistName, opts.TrackTitle, cfg.ArtistSeparators()))
		}

		a, err := matchOrCreateArtistByMbzID(ctx, id, d, opts)
		if err != nil {
			l.Warn().Msg("MusicBrainz unreachable, falling back to artist name matching")
			return matchArtistsByNames(ctx, opts.ArtistNames, result, d, opts)
		}

		result = append(result, a)
	}

	return result, nil
}

func matchOrCreateArtistByMbzID(ctx context.Context, id uuid.UUID, d db.ArtistStore, opts AssociateArtistsOpts) (*models.Artist, error) {
	l := logger.FromContext(ctx)

	unlock := lockKey("artist:mbid:" + id.String())
	defer unlock()

	a, err := d.GetArtist(ctx, db.GetArtistOpts{
		MusicBrainzID: id,
	})
	if err == nil {
		l.Debug().Msgf("Artist '%s' found by MusicBrainz ID", a.Name)
		return a, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	return resolveAliasOrCreateArtist(ctx, id, opts.ArtistNames, d, opts)
}

func resolveAliasOrCreateArtist(ctx context.Context, mbzID uuid.UUID, names []string, d db.ArtistStore, opts AssociateArtistsOpts) (*models.Artist, error) {
	l := logger.FromContext(ctx)

	aliases, err := opts.Mbzc.GetArtistPrimaryAliases(ctx, mbzID)
	if err != nil {
		return nil, fmt.Errorf("resolveAliasOrCreateArtist: %w", err)
	}
	l.Debug().Msgf("Got aliases %v from MusicBrainz", aliases)

	for _, alias := range aliases {
		a, err := d.GetArtist(ctx, db.GetArtistOpts{
			Name: alias,
		})
		if err == nil && (a.MbzID == nil || *a.MbzID == uuid.Nil) {
			a.MbzID = &mbzID
			l.Debug().Msgf("Alias '%s' found in DB. Associating with MusicBrainz ID...", alias)
			if updateErr := d.UpdateArtist(ctx, db.UpdateArtistOpts{ID: a.ID, MusicBrainzID: mbzID}); updateErr != nil {
				return nil, fmt.Errorf("resolveAliasOrCreateArtist: %w", updateErr)
			}
			if saveAliasErr := d.SaveArtistAliases(ctx, a.ID, aliases, "MusicBrainz"); saveAliasErr != nil {
				return nil, fmt.Errorf("resolveAliasOrCreateArtist: %w", saveAliasErr)
			}
			return a, nil
		}
	}

	canonical := aliases[0]
	for _, alias := range aliases {
		for _, name := range names {
			if strings.EqualFold(alias, name) {
				l.Debug().Msgf("Canonical name for artist is '%s'", alias)
				canonical = alias
				break
			}
		}
	}

	var imgid uuid.UUID
	var imgUrl string
	if !opts.SkipCacheImage {
		var err error
		imgUrl, err = images.GetArtistImage(ctx, images.ArtistImageOpts{
			Aliases:   aliases,
			OriginUrl: opts.OriginUrl,
		})
		if err == nil && imgUrl != "" {
			imgid = uuid.New()
			l.Debug().Msg("Downloading artist image from source...")
			if dlErr := imagecache.DownloadImage(imgid, imgUrl); dlErr != nil {
				l.Err(dlErr).Msg("Failed to cache image")
			}
		} else if err != nil {
			l.Warn().AnErr("error", err).Msg("Failed to get artist image from ImageSrc")
		}
	}

	u, err := d.SaveArtist(ctx, db.SaveArtistOpts{
		MusicBrainzID: mbzID,
		Name:          canonical,
		Aliases:       aliases,
		Image:         imgid,
		ImageSrc:      imgUrl,
	})
	if err != nil {
		return nil, fmt.Errorf("resolveAliasOrCreateArtist: %w", err)
	}
	l.Info().Msgf("Created artist '%s' with MusicBrainz Artist ID", canonical)
	return u, nil
}

func matchArtistsByNames(ctx context.Context, names []string, existing []*models.Artist, d db.ArtistStore, opts AssociateArtistsOpts) ([]*models.Artist, error) {
	l := logger.FromContext(ctx)
	var result []*models.Artist

	for _, name := range names {
		if artistExists(name, existing) || artistExists(name, result) {
			l.Debug().Msgf("Artist '%s' already found, skipping...", name)
			continue
		}

		unlock := lockKey("artist:" + strings.ToLower(name))
		a, err := matchOrCreateArtistByName(ctx, name, d, opts)
		unlock()
		if err != nil {
			return nil, fmt.Errorf("matchArtistsByNames: %w", err)
		}
		result = append(result, a)
	}
	return result, nil
}

func matchOrCreateArtistByName(ctx context.Context, name string, d db.ArtistStore, opts AssociateArtistsOpts) (*models.Artist, error) {
	l := logger.FromContext(ctx)
	a, err := d.GetArtist(ctx, db.GetArtistOpts{
		Name: name,
	})
	if err == nil {
		l.Debug().Msgf("Artist '%s' found in DB", name)
		return a, nil
	}
	if !errors.Is(err, db.ErrNotFound) {
		return nil, err
	}

	var imgid uuid.UUID
	var imgUrl string
	if !opts.SkipCacheImage {
		var imgErr error
		imgUrl, imgErr = images.GetArtistImage(ctx, images.ArtistImageOpts{
			Aliases:   []string{name},
			OriginUrl: opts.OriginUrl,
		})
		if imgErr == nil && imgUrl != "" {
			imgid = uuid.New()
			l.Debug().Msg("Downloading artist image from source...")
			if dlErr := imagecache.DownloadImage(imgid, imgUrl); dlErr != nil {
				l.Err(dlErr).Msg("Failed to cache image")
			}
		} else if imgErr != nil {
			l.Debug().AnErr("error", imgErr).Msgf("Failed to get artist images for %s", name)
		}
	}

	a, err = d.SaveArtist(ctx, db.SaveArtistOpts{Name: name, Image: imgid, ImageSrc: imgUrl})
	if err != nil {
		return nil, err
	}
	l.Info().Msgf("Created artist '%s' with artist name", name)
	return a, nil
}

func artistExists(name string, artists []*models.Artist) bool {
	for _, a := range artists {
		allAliases := append(a.Aliases, a.Name)
		for _, alias := range allAliases {
			if strings.EqualFold(name, alias) {
				return true
			}
		}
	}
	return false
}
func artistExistsByMbzID(id uuid.UUID, artists []*models.Artist) bool {
	for _, a := range artists {
		if a.MbzID != nil && *a.MbzID == id {
			return true
		}
	}
	return false
}
