package catalog

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/images"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AssociateAlbumOpts struct {
	Artists           []*models.Artist
	ReleaseMbzID      uuid.UUID
	ReleaseGroupMbzID uuid.UUID
	ReleaseName       string
	TrackName         string // required
	Mbzc              mbz.MusicBrainzCaller
	SkipCacheImage    bool
}

func AssociateAlbum(ctx context.Context, d db.DB, opts AssociateAlbumOpts) (*models.Album, error) {
	l := logger.FromContext(ctx)
	if opts.TrackName == "" {
		return nil, errors.New("AssociateAlbum: required parameter TrackName missing")
	}
	releaseTitle := opts.ReleaseName
	if releaseTitle == "" {
		releaseTitle = opts.TrackName
	}
	if opts.ReleaseMbzID != uuid.Nil {
		l.Debug().Msgf("Associating album '%s' by MusicBrainz release ID", releaseTitle)
		return matchAlbumByMbzReleaseID(ctx, d, opts)
	} else {
		l.Debug().Msgf("Associating album '%s' by title and artist", releaseTitle)
		return matchAlbumByTitle(ctx, d, opts)
	}
}

func matchAlbumByMbzReleaseID(ctx context.Context, d db.DB, opts AssociateAlbumOpts) (*models.Album, error) {
	l := logger.FromContext(ctx)
	a, err := d.GetAlbum(ctx, db.GetAlbumOpts{MusicBrainzID: opts.ReleaseMbzID})
	if err == nil {
		l.Debug().Msgf("Found release '%s' by MusicBrainz Release ID", a.Title)
		return &models.Album{
			ID:             a.ID,
			MbzID:          &opts.ReleaseMbzID,
			Title:          a.Title,
			VariousArtists: a.VariousArtists,
			Image:          a.Image,
		}, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("matchAlbumByMbzReleaseID: %w", err)
	} else {
		l.Debug().Msgf("Album '%s' could not be found by MusicBrainz Release ID", opts.ReleaseName)
		rg, err := createOrUpdateAlbumWithMbzReleaseID(ctx, d, opts)
		if err != nil {
			return matchAlbumByTitle(ctx, d, opts)
		}
		return rg, nil
	}
}

func createOrUpdateAlbumWithMbzReleaseID(ctx context.Context, d db.DB, opts AssociateAlbumOpts) (*models.Album, error) {
	l := logger.FromContext(ctx)

	release, err := opts.Mbzc.GetRelease(ctx, opts.ReleaseMbzID)
	if err != nil {
		l.Warn().Msg("createOrUpdateAlbumWithMbzReleaseID: MusicBrainz unreachable, falling back to release title matching")
		return matchAlbumByTitle(ctx, d, opts)
	}

	var album *models.Album
	titles := []string{release.Title, opts.ReleaseName}
	utils.Unique(&titles)

	l.Debug().Msgf("Searching for albums '%v' from artist id %d in DB", titles, opts.Artists[0].ID)
	album, err = d.GetAlbum(ctx, db.GetAlbumOpts{
		ArtistID: opts.Artists[0].ID,
		Titles:   titles,
	})
	if err == nil {
		l.Debug().Msgf("Found album %s, updating with MusicBrainz Release ID...", album.Title)
		if album.MbzID == nil {
			err := d.UpdateAlbum(ctx, db.UpdateAlbumOpts{
				ID:            album.ID,
				MusicBrainzID: opts.ReleaseMbzID,
			})
			if err != nil {
				l.Err(err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to update album with MusicBrainz Release ID")
				return nil, fmt.Errorf("createOrUpdateAlbumWithMbzReleaseID: %w", err)
			}
			l.Debug().Msgf("Updated album '%s' with MusicBrainz Release ID", album.Title)
		} else {
			l.Warn().Msgf("Attempted to update album %s with MusicBrainz ID, but an existing ID was already found", album.Title)
		}

		if opts.ReleaseGroupMbzID != uuid.Nil {
			aliases, err := opts.Mbzc.GetReleaseTitles(ctx, opts.ReleaseGroupMbzID)
			if err == nil {
				l.Debug().Msgf("Associating aliases '%s' with Release '%s'", aliases, album.Title)
				err = d.SaveAlbumAliases(ctx, album.ID, aliases, "MusicBrainz")
				if err != nil {
					l.Err(err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to save aliases")
				}
			} else {
				l.Info().AnErr("err", err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to get release group from MusicBrainz")
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		l.Err(err).Msg("createOrUpdateAlbumWithMbzReleaseID: error while searching for album by MusicBrainz Release ID")
		return nil, fmt.Errorf("createOrUpdateAlbumWithMbzReleaseID: %w", err)
	} else {
		l.Debug().Msgf("Album %s could not be found. Creating...", release.Title)

		var variousArtists bool
		for _, artistCredit := range release.ArtistCredit {
			if artistCredit.Name == "Various Artists" {
				l.Debug().Msgf("MusicBrainz release group '%s' detected as being a Various Artists compilation release", release.Title)
				variousArtists = true
			}
		}

		l.Debug().Msg("Searching for album images...")
		var imgid uuid.UUID
		imgUrl, err := images.GetAlbumImage(ctx, images.AlbumImageOpts{
			Artists:      utils.UniqueIgnoringCase(slices.Concat(utils.FlattenMbzArtistCreditNames(release.ArtistCredit), utils.FlattenArtistNames(opts.Artists))),
			Album:        release.Title,
			ReleaseMbzID: &opts.ReleaseMbzID,
		})

		if err == nil && imgUrl != "" {
			imgid = uuid.New()
			if !opts.SkipCacheImage {
				var size ImageSize
				if cfg.FullImageCacheEnabled() {
					size = ImageSizeFull
				} else {
					size = ImageSizeLarge
				}
				l.Debug().Msg("Downloading album image from source...")
				err = DownloadAndCacheImage(ctx, imgid, imgUrl, size)
				if err != nil {
					l.Err(err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to cache image")
				}
			}
		}

		if err != nil {
			l.Debug().Msgf("createOrUpdateAlbumWithMbzReleaseID: failed to get album images for %s: %s", release.Title, err.Error())
		}

		album, err = d.SaveAlbum(ctx, db.SaveAlbumOpts{
			Title:          release.Title,
			MusicBrainzID:  opts.ReleaseMbzID,
			ArtistIDs:      utils.FlattenArtistIDs(opts.Artists),
			VariousArtists: variousArtists,
			Image:          imgid,
			ImageSrc:       imgUrl,
		})
		if err != nil {
			return nil, fmt.Errorf("createOrUpdateAlbumWithMbzReleaseID: %w", err)
		}

		if opts.ReleaseGroupMbzID != uuid.Nil {
			aliases, err := opts.Mbzc.GetReleaseTitles(ctx, opts.ReleaseGroupMbzID)
			if err == nil {
				l.Debug().Msgf("Associating aliases '%s' with Release '%s'", aliases, album.Title)
				err = d.SaveAlbumAliases(ctx, album.ID, aliases, "MusicBrainz")
				if err != nil {
					l.Err(err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to save aliases")
				}
			} else {
				l.Info().AnErr("err", err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to get release group from MusicBrainz")
			}
		}

		l.Info().Msgf("Created album '%s' with MusicBrainz Release ID", album.Title)
	}

	return &models.Album{
		ID:             album.ID,
		MbzID:          &opts.ReleaseMbzID,
		Title:          album.Title,
		VariousArtists: album.VariousArtists,
	}, nil
}

func matchAlbumByTitle(ctx context.Context, d db.DB, opts AssociateAlbumOpts) (*models.Album, error) {
	l := logger.FromContext(ctx)

	var releaseName string
	if opts.ReleaseName != "" {
		releaseName = opts.ReleaseName
	} else {
		releaseName = opts.TrackName
	}

	a, err := d.GetAlbum(ctx, db.GetAlbumOpts{
		Title:    releaseName,
		ArtistID: opts.Artists[0].ID,
	})
	if err == nil {
		l.Debug().Msgf("Found album '%s' by artist and title", a.Title)
		if a.MbzID == nil && opts.ReleaseMbzID != uuid.Nil {
			l.Debug().Msgf("Updating album with id %d with MusicBrainz ID %s", a.ID, opts.ReleaseMbzID)
			err = d.UpdateAlbum(ctx, db.UpdateAlbumOpts{
				ID:            a.ID,
				MusicBrainzID: opts.ReleaseMbzID,
			})
			if err != nil {
				l.Err(err).Msg("matchAlbumByTitle: failed to associate existing release with MusicBrainz ID")
			}
		}
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("matchAlbumByTitle: %w", err)
	} else {
		var imgid uuid.UUID
		imgUrl, err := images.GetAlbumImage(ctx, images.AlbumImageOpts{
			Artists:      utils.FlattenArtistNames(opts.Artists),
			Album:        opts.ReleaseName,
			ReleaseMbzID: &opts.ReleaseMbzID,
		})
		if err == nil && imgUrl != "" {
			imgid = uuid.New()
			if !opts.SkipCacheImage {
				var size ImageSize
				if cfg.FullImageCacheEnabled() {
					size = ImageSizeFull
				} else {
					size = ImageSizeLarge
				}
				l.Debug().Msg("Downloading album image from source...")
				err = DownloadAndCacheImage(ctx, imgid, imgUrl, size)
				if err != nil {
					l.Err(err).Msg("createOrUpdateAlbumWithMbzReleaseID: failed to cache image")
				}
			}
		}
		if err != nil {
			l.Debug().AnErr("error", err).Msgf("matchAlbumByTitle: failed to get album images for %s", opts.ReleaseName)
		}

		a, err = d.SaveAlbum(ctx, db.SaveAlbumOpts{
			Title:         releaseName,
			ArtistIDs:     utils.FlattenArtistIDs(opts.Artists),
			Image:         imgid,
			MusicBrainzID: opts.ReleaseMbzID,
			ImageSrc:      imgUrl,
		})
		if err != nil {
			return nil, fmt.Errorf("matchAlbumByTitle: %w", err)
		}
		l.Info().Msgf("Created album '%s' with artist and title", a.Title)
	}

	return &models.Album{
		ID:    a.ID,
		Title: a.Title,
	}, nil
}
