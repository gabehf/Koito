package catalog

import (
	"context"
	"errors"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/gabehf/koito/internal/models"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type AssociateTrackOpts struct {
	ArtistIDs  []int32
	AlbumID    int32
	TrackMbzID uuid.UUID
	TrackName  string
	Duration   int32
	Mbzc       mbz.MusicBrainzCaller
}

func AssociateTrack(ctx context.Context, d db.DB, opts AssociateTrackOpts) (*models.Track, error) {
	l := logger.FromContext(ctx)
	if opts.TrackName == "" {
		return nil, errors.New("AssociateTrack: missing required parameter 'opts.TrackName'")
	}
	if len(opts.ArtistIDs) < 1 {
		return nil, errors.New("AssociateTrack: at least one artist id must be specified")
	}
	if opts.AlbumID == 0 {
		return nil, errors.New("AssociateTrack: release group id must be specified")
	}
	// first, try to match track Mbz ID
	if opts.TrackMbzID != uuid.Nil {
		l.Debug().Msgf("Associating track '%s' by MusicBrainz recording ID", opts.TrackName)
		return matchTrackByMbzID(ctx, d, opts)
	} else {
		l.Debug().Msgf("Associating track '%s' by title and artist", opts.TrackName)
		return matchTrackByTrackInfo(ctx, d, opts)
	}
}

// If no match is found, will call matchTrackByTitleAndArtist and associate the Mbz ID with the result
func matchTrackByMbzID(ctx context.Context, d db.DB, opts AssociateTrackOpts) (*models.Track, error) {
	l := logger.FromContext(ctx)
	track, err := d.GetTrack(ctx, db.GetTrackOpts{
		MusicBrainzID: opts.TrackMbzID,
	})
	if err == nil {
		l.Debug().Msgf("Found track '%s' by MusicBrainz ID", track.Title)
		return track, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("matchTrackByMbzID: %w", err)
	} else {
		l.Debug().Msgf("Track '%s' could not be found by MusicBrainz ID", opts.TrackName)
		track, err := matchTrackByTrackInfo(ctx, d, opts)
		if err != nil {
			return nil, fmt.Errorf("matchTrackByMbzID: %w", err)
		}
		l.Debug().Msgf("Updating track '%s' with MusicBrainz ID %s", opts.TrackName, opts.TrackMbzID)
		if track.MbzID == nil || *track.MbzID == uuid.Nil {
			err := d.UpdateTrack(ctx, db.UpdateTrackOpts{
				ID:            track.ID,
				MusicBrainzID: opts.TrackMbzID,
			})
			if err != nil {
				l.Err(err).Msg("matchArtistsByMBIDMappings: failed to update track with MusicBrainz ID")
				return nil, fmt.Errorf("matchArtistsByMBIDMappings: %w", err)
			}
			l.Debug().Msgf("Updated track '%s' with MusicBrainz ID", track.Title)
		} else {
			l.Warn().Msgf("Attempted to update track %s with MusicBrainz ID, but an existing ID was already found", track.Title)
		}
		if err != nil {
			return nil, fmt.Errorf("matchTrackByMbzID: %w", err)
		}
		track.MbzID = &opts.TrackMbzID
		return track, nil
	}
}

func matchTrackByTrackInfo(ctx context.Context, d db.DB, opts AssociateTrackOpts) (*models.Track, error) {
	l := logger.FromContext(ctx)
	// try provided track title
	track, err := d.GetTrack(ctx, db.GetTrackOpts{
		Title:     opts.TrackName,
		ReleaseID: opts.AlbumID,
		ArtistIDs: opts.ArtistIDs,
	})
	if err == nil {
		l.Debug().Msgf("Track '%s' found by title, release and artist match", track.Title)
		return track, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("matchTrackByTrackInfo: %w", err)
	} else {
		if opts.TrackMbzID != uuid.Nil {
			mbzTrack, err := opts.Mbzc.GetTrack(ctx, opts.TrackMbzID)
			if err == nil {
				track, err := d.GetTrack(ctx, db.GetTrackOpts{
					Title:     mbzTrack.Title,
					ReleaseID: opts.AlbumID,
					ArtistIDs: opts.ArtistIDs,
				})
				if err == nil {
					l.Debug().Msgf("Track '%s' found by MusicBrainz title, release and artist match", opts.TrackName)
					return track, nil
				}
			}
		}
		l.Debug().Msgf("Track '%s' could not be found by title and artist match", opts.TrackName)
		t, err := d.SaveTrack(ctx, db.SaveTrackOpts{
			RecordingMbzID: opts.TrackMbzID,
			AlbumID:        opts.AlbumID,
			Title:          opts.TrackName,
			ArtistIDs:      opts.ArtistIDs,
			Duration:       opts.Duration,
		})
		if err != nil {
			return nil, fmt.Errorf("matchTrackByTrackInfo: %w", err)
		}
		if opts.TrackMbzID == uuid.Nil {
			l.Info().Msgf("Created track '%s' with title and artist", opts.TrackName)
		} else {
			l.Info().Msgf("Created track '%s' with MusicBrainz Recording ID", opts.TrackName)
		}
		return t, nil
	}
}
