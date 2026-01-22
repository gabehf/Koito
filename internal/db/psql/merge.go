package psql

import (
	"context"
	"fmt"

	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5"
)

func (d *Psql) MergeTracks(ctx context.Context, fromId, toId int32) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Merging track %d into track %d", fromId, toId)
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("MergeTracks: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	from, err := qtx.GetTrack(ctx, fromId)
	if err != nil {
		return fmt.Errorf("MergeTracks: GetTrack: %w", err)
	}
	to, err := qtx.GetTrack(ctx, toId)
	if err != nil {
		return fmt.Errorf("MergeTracks: GetTrack: %w", err)
	}
	err = qtx.UpdateTrackIdForListens(ctx, repository.UpdateTrackIdForListensParams{
		TrackID:   fromId,
		TrackID_2: toId,
	})
	if err != nil {
		return fmt.Errorf("MergeTracks: UpdateTrackIdForListens: %w", err)
	}
	if from.ReleaseID != to.ReleaseID {
		// tracks are from different releases, track artist should be associated with to.release
		artists, err := qtx.GetTrackArtists(ctx, fromId)
		if err != nil {
			return fmt.Errorf("MergeTracks: GetTrackArtists: %w", err)
		}
		for _, artist := range artists {
			err = qtx.AssociateArtistToRelease(ctx, repository.AssociateArtistToReleaseParams{
				ArtistID:  artist.ID,
				ReleaseID: to.ReleaseID,
			})
			if err != nil {
				return fmt.Errorf("MergeTracks: AssociateArtistToRelease: %w", err)
			}
		}
	}
	err = qtx.CleanOrphanedEntries(ctx)
	if err != nil {
		l.Err(err).Msg("MergeTracks: Failed to clean orphaned entries")
		return err
	}
	return tx.Commit(ctx)
}

func (d *Psql) MergeAlbums(ctx context.Context, fromId, toId int32, replaceImage bool) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Merging album %d into album %d", fromId, toId)
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("MergeAlbums: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)

	fromArtists, err := qtx.GetReleaseArtists(ctx, fromId)
	if err != nil {
		return fmt.Errorf("MergeAlbums: GetReleaseArtists: %w", err)
	}

	err = qtx.UpdateReleaseForAll(ctx, repository.UpdateReleaseForAllParams{
		ReleaseID:   fromId,
		ReleaseID_2: toId,
	})
	if err != nil {
		return fmt.Errorf("MergeAlbums: %w", err)
	}
	if replaceImage {
		old, err := qtx.GetRelease(ctx, fromId)
		if err != nil {
			return fmt.Errorf("MergeAlbums: %w", err)
		}
		err = qtx.UpdateReleaseImage(ctx, repository.UpdateReleaseImageParams{
			ID:          toId,
			Image:       old.Image,
			ImageSource: old.ImageSource,
		})
		if err != nil {
			return fmt.Errorf("MergeAlbums: %w", err)
		}
	}

	for _, artist := range fromArtists {
		err = qtx.AssociateArtistToRelease(ctx, repository.AssociateArtistToReleaseParams{
			ArtistID:  artist.ID,
			ReleaseID: toId,
		})
		if err != nil {
			return fmt.Errorf("MergeAlbums: AssociateArtistToRelease: %w", err)
		}
	}

	err = qtx.CleanOrphanedEntries(ctx)
	if err != nil {
		l.Err(err).Msg("Failed to clean orphaned entries")
		return fmt.Errorf("MergeAlbums: CleanOrphanedEntries: %w", err)
	}
	return tx.Commit(ctx)
}

func (d *Psql) MergeArtists(ctx context.Context, fromId, toId int32, replaceImage bool) error {
	l := logger.FromContext(ctx)
	l.Info().Msgf("Merging artist %d into artist %d", fromId, toId)
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("MergeArtists: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	err = qtx.DeleteConflictingArtistTracks(ctx, repository.DeleteConflictingArtistTracksParams{
		ArtistID:   fromId,
		ArtistID_2: toId,
	})
	if err != nil {
		l.Err(err).Msg("Failed to delete conflicting artist tracks")
		return fmt.Errorf("MergeArtists: %w", err)
	}
	err = qtx.DeleteConflictingArtistReleases(ctx, repository.DeleteConflictingArtistReleasesParams{
		ArtistID:   fromId,
		ArtistID_2: toId,
	})
	if err != nil {
		l.Err(err).Msg("Failed to delete conflicting artist releases")
		return fmt.Errorf("MergeArtists: %w", err)
	}
	err = qtx.UpdateArtistTracks(ctx, repository.UpdateArtistTracksParams{
		ArtistID:   fromId,
		ArtistID_2: toId,
	})
	if err != nil {
		l.Err(err).Msg("Failed to update artist tracks")
		return fmt.Errorf("MergeArtists: %w", err)
	}
	err = qtx.UpdateArtistReleases(ctx, repository.UpdateArtistReleasesParams{
		ArtistID:   fromId,
		ArtistID_2: toId,
	})
	if err != nil {
		l.Err(err).Msg("Failed to update artist releases")
		return fmt.Errorf("MergeArtists: %w", err)
	}
	if replaceImage {
		old, err := qtx.GetArtist(ctx, fromId)
		if err != nil {
			return fmt.Errorf("MergeAlbums: %w", err)
		}
		err = qtx.UpdateArtistImage(ctx, repository.UpdateArtistImageParams{
			ID:          toId,
			Image:       old.Image,
			ImageSource: old.ImageSource,
		})
		if err != nil {
			return fmt.Errorf("MergeAlbums: %w", err)
		}
	}
	err = qtx.CleanOrphanedEntries(ctx)
	if err != nil {
		l.Err(err).Msg("Failed to clean orphaned entries")
		return fmt.Errorf("MergeArtists: %w", err)
	}
	return tx.Commit(ctx)
}
