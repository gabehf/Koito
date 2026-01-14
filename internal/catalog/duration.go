package catalog

import (
	"context"
	"fmt"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/mbz"
	"github.com/google/uuid"
)

func BackfillTrackDurationsFromMusicBrainz(
	ctx context.Context,
	store db.DB,
	mbzCaller mbz.MusicBrainzCaller,
) error {
	l := logger.FromContext(ctx)
	l.Info().Msg("BackfillTrackDurationsFromMusicBrainz: Starting backfill of track durations from MusicBrainz")

	var from int32 = 0

	for {
		tracks, err := store.GetTracksWithNoDurationButHaveMbzID(ctx, from)
		if err != nil {
			return fmt.Errorf("BackfillTrackDurationsFromMusicBrainz: failed to fetch tracks for duration backfill: %w", err)
		}

		// nil, nil means no more results
		if len(tracks) == 0 {
			if from == 0 {
				l.Info().Msg("BackfillTrackDurationsFromMusicBrainz: No tracks need updating. Skipping backfill...")
			} else {
				l.Info().Msg("BackfillTrackDurationsFromMusicBrainz: Backfill complete")
			}
			return nil
		}

		for _, track := range tracks {
			from = track.ID

			if track.MbzID == nil || *track.MbzID == uuid.Nil {
				continue
			}

			l.Debug().
				Str("title", track.Title).
				Str("mbz_id", track.MbzID.String()).
				Msg("BackfillTrackDurationsFromMusicBrainz: Backfilling duration from MusicBrainz")

			mbzTrack, err := mbzCaller.GetTrack(ctx, *track.MbzID)
			if err != nil {
				l.Err(err).
					Str("title", track.Title).
					Msg("BackfillTrackDurationsFromMusicBrainz: Failed to fetch track from MusicBrainz")
				continue
			}

			if mbzTrack.LengthMs <= 0 {
				l.Debug().
					Str("title", track.Title).
					Msg("BackfillTrackDurationsFromMusicBrainz: MusicBrainz track has no duration")
				continue
			}

			durationSeconds := int32(mbzTrack.LengthMs / 1000)

			err = store.UpdateTrack(ctx, db.UpdateTrackOpts{
				ID:       track.ID,
				Duration: durationSeconds,
			})
			if err != nil {
				l.Err(err).
					Str("title", track.Title).
					Msg("BackfillTrackDurationsFromMusicBrainz: Failed to update track duration")
			} else {
				l.Info().
					Str("title", track.Title).
					Int32("duration_seconds", durationSeconds).
					Msg("BackfillTrackDurationsFromMusicBrainz: Track duration backfilled successfully")
			}
		}
	}
}
