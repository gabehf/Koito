package importer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/export"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func ImportKoitoFile(ctx context.Context, store db.DB, filename string) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Beginning Koito import on file: %s", filename)
	data := new(export.KoitoExport)
	f, err := os.Open(path.Join(cfg.ConfigDir(), "import", filename))
	if err != nil {
		return fmt.Errorf("ImportKoitoFile: os.Open: %w", err)
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(data)
	if err != nil {
		return fmt.Errorf("ImportKoitoFile: Decode: %w", err)
	}

	if data.Version != "1" {
		return fmt.Errorf("ImportKoitoFile: unupported version: %s", data.Version)
	}

	l.Info().Msgf("Beginning data import for user: %s", data.User)

	count := 0

	for i := range data.Listens {
		// use this for save/get mbid for all artist/album/track
		var mbid uuid.UUID

		artistIds := make([]int32, 0)
		for _, ia := range data.Listens[i].Artists {
			mbid = uuid.Nil
			if ia.MBID != nil {
				mbid = *ia.MBID
			}
			artist, err := store.GetArtist(ctx, db.GetArtistOpts{
				MusicBrainzID: mbid,
				Name:          getPrimaryAliasFromAliasSlice(ia.Aliases),
			})
			if errors.Is(err, pgx.ErrNoRows) {
				var imgid = uuid.Nil
				// not a perfect way to check if the image url is an actual source vs manual upload but
				// im like 99% sure it will work perfectly
				if strings.HasPrefix(ia.ImageUrl, "http") {
					imgid = uuid.New()
				}
				// save artist
				artist, err := store.SaveArtist(ctx, db.SaveArtistOpts{
					Name:          getPrimaryAliasFromAliasSlice(ia.Aliases),
					Image:         imgid,
					ImageSrc:      ia.ImageUrl,
					MusicBrainzID: mbid,
					Aliases:       utils.FlattenAliases(ia.Aliases),
				})
				if err != nil {
					return fmt.Errorf("ImportKoitoFile: %w", err)
				}
				artistIds = append(artistIds, artist.ID)
			} else if err != nil {
				return fmt.Errorf("ImportKoitoFile: %w", err)
			} else {
				artistIds = append(artistIds, artist.ID)
			}
		}
		// call associate album
		albumId := int32(0)
		mbid = uuid.Nil
		if data.Listens[i].Album.MBID != nil {
			mbid = *data.Listens[i].Album.MBID
		}
		album, err := store.GetAlbum(ctx, db.GetAlbumOpts{
			MusicBrainzID: mbid,
			Title:         getPrimaryAliasFromAliasSlice(data.Listens[i].Album.Aliases),
			ArtistID:      artistIds[0],
		})
		if errors.Is(err, pgx.ErrNoRows) {
			var imgid = uuid.Nil
			// not a perfect way to check if the image url is an actual source vs manual upload but
			// im like 99% sure it will work perfectly
			if strings.HasPrefix(data.Listens[i].Album.ImageUrl, "http") {
				imgid = uuid.New()
			}
			// save album
			album, err = store.SaveAlbum(ctx, db.SaveAlbumOpts{
				Title:          getPrimaryAliasFromAliasSlice(data.Listens[i].Album.Aliases),
				Image:          imgid,
				ImageSrc:       data.Listens[i].Album.ImageUrl,
				MusicBrainzID:  mbid,
				Aliases:        utils.FlattenAliases(data.Listens[i].Album.Aliases),
				ArtistIDs:      artistIds,
				VariousArtists: data.Listens[i].Album.VariousArtists,
			})
			if err != nil {
				return fmt.Errorf("ImportKoitoFile: %w", err)
			}
			albumId = album.ID
		} else if err != nil {
			return fmt.Errorf("ImportKoitoFile: %w", err)
		} else {
			albumId = album.ID
		}

		// call associate track
		mbid = uuid.Nil
		if data.Listens[i].Track.MBID != nil {
			mbid = *data.Listens[i].Track.MBID
		}
		track, err := store.GetTrack(ctx, db.GetTrackOpts{
			MusicBrainzID: mbid,
			Title:         getPrimaryAliasFromAliasSlice(data.Listens[i].Track.Aliases),
			ReleaseID:     albumId,
			ArtistIDs:     artistIds,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			// save track
			track, err = store.SaveTrack(ctx, db.SaveTrackOpts{
				Title:          getPrimaryAliasFromAliasSlice(data.Listens[i].Track.Aliases),
				RecordingMbzID: mbid,
				Duration:       int32(data.Listens[i].Track.Duration),
				ArtistIDs:      artistIds,
				AlbumID:        albumId,
			})
			if err != nil {
				return fmt.Errorf("ImportKoitoFile: %w", err)
			}
			// save track aliases
			err = store.SaveTrackAliases(ctx, track.ID, utils.FlattenAliases(data.Listens[i].Track.Aliases), "Import")
			if err != nil {
				return fmt.Errorf("ImportKoitoFile: %w", err)
			}
		} else if err != nil {
			return fmt.Errorf("ImportKoitoFile: %w", err)
		}

		// save listen
		err = store.SaveListen(ctx, db.SaveListenOpts{
			TrackID: track.ID,
			Time:    data.Listens[i].ListenedAt,
			UserID:  1,
		})
		if err != nil {
			return fmt.Errorf("ImportKoitoFile: %w", err)
		}

		l.Info().Msgf("ImportKoitoFile: Imported listen for track %s", track.Title)
		count++
	}

	return finishImport(ctx, filename, count)
}
func getPrimaryAliasFromAliasSlice(aliases []models.Alias) string {
	for _, a := range aliases {
		if a.Primary {
			return a.Alias
		}
	}
	return ""
}
