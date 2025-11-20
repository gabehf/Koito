package psql

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/gabehf/koito/internal/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (d *Psql) GetTrack(ctx context.Context, opts db.GetTrackOpts) (*models.Track, error) {
	l := logger.FromContext(ctx)
	var track models.Track

	if opts.ID != 0 {
		l.Debug().Msgf("Fetching track from DB with id %d", opts.ID)
		t, err := d.q.GetTrack(ctx, opts.ID)
		if err != nil {
			return nil, fmt.Errorf("GetTrack: GetTrack By ID: %w", err)
		}
		track = models.Track{
			ID:       t.ID,
			MbzID:    t.MusicBrainzID,
			Title:    t.Title,
			AlbumID:  t.ReleaseID,
			Image:    t.Image,
			Duration: t.Duration,
		}
		err = json.Unmarshal(t.Artists, &track.Artists)
		if err != nil {
			return nil, fmt.Errorf("GetTrack: json.Unmarshal: %w", err)
		}
	} else if opts.MusicBrainzID != uuid.Nil {
		l.Debug().Msgf("Fetching track from DB with MusicBrainz ID %s", opts.MusicBrainzID)
		t, err := d.q.GetTrackByMbzID(ctx, &opts.MusicBrainzID)
		if err != nil {
			return nil, fmt.Errorf("GetTrack: GetTrackByMbzID: %w", err)
		}
		track = models.Track{
			ID:       t.ID,
			MbzID:    t.MusicBrainzID,
			Title:    t.Title,
			AlbumID:  t.ReleaseID,
			Duration: t.Duration,
		}
	} else if len(opts.ArtistIDs) > 0 {
		l.Debug().Msgf("Fetching track from DB with title '%s' and artist id(s) '%v'", opts.Title, opts.ArtistIDs)
		t, err := d.q.GetTrackByTitleAndArtists(ctx, repository.GetTrackByTitleAndArtistsParams{
			Title:   opts.Title,
			Column2: opts.ArtistIDs,
		})
		if err != nil {
			return nil, fmt.Errorf("GetTrack: GetTrackByTitleAndArtists: %w", err)
		}
		track = models.Track{
			ID:       t.ID,
			MbzID:    t.MusicBrainzID,
			Title:    t.Title,
			AlbumID:  t.ReleaseID,
			Duration: t.Duration,
		}
	} else {
		return nil, errors.New("GetTrack: insufficient information to get track")
	}

	count, err := d.q.CountListensFromTrack(ctx, repository.CountListensFromTrackParams{
		ListenedAt:   time.Unix(0, 0),
		ListenedAt_2: time.Now(),
		TrackID:      track.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetTrack: CountListensFromTrack: %w", err)
	}

	seconds, err := d.CountTimeListenedToItem(ctx, db.TimeListenedOpts{
		Period:  db.PeriodAllTime,
		TrackID: track.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetTrack: CountTimeListenedToItem: %w", err)
	}

	firstListen, err := d.q.GetFirstListenFromTrack(ctx, track.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetAlbum: GetFirstListenFromRelease: %w", err)
	}

	track.ListenCount = count
	track.TimeListened = seconds
	track.FirstListen = firstListen.ListenedAt.Unix()

	return &track, nil
}

func (d *Psql) SaveTrack(ctx context.Context, opts db.SaveTrackOpts) (*models.Track, error) {
	// create track in DB
	l := logger.FromContext(ctx)
	var insertMbzID *uuid.UUID
	if opts.RecordingMbzID != uuid.Nil {
		insertMbzID = &opts.RecordingMbzID
	}
	if len(opts.ArtistIDs) < 1 {
		return nil, errors.New("SaveTrack: required parameter 'ArtistIDs' missing")
	}
	for _, aid := range opts.ArtistIDs {
		if aid == 0 {
			return nil, errors.New("SaveTrack: none of 'ArtistIDs' may be 0")
		}
	}
	if opts.AlbumID == 0 {
		return nil, errors.New("SaveTrack: required parameter 'AlbumID' missing")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return nil, fmt.Errorf("SaveTrack: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	l.Debug().Msgf("Inserting new track '%s' into DB", opts.Title)
	trackRow, err := qtx.InsertTrack(ctx, repository.InsertTrackParams{
		MusicBrainzID: insertMbzID,
		ReleaseID:     opts.AlbumID,
		Duration:      opts.Duration,
	})
	if err != nil {
		return nil, fmt.Errorf("SaveTrack: InsertTrack: %w", err)
	}
	// insert associated artists
	for _, aid := range opts.ArtistIDs {
		err = qtx.AssociateArtistToTrack(ctx, repository.AssociateArtistToTrackParams{
			ArtistID:  aid,
			TrackID:   trackRow.ID,
			IsPrimary: opts.ArtistIDs[0] == aid,
		})
		if err != nil {
			return nil, fmt.Errorf("SaveTrack: AssociateArtistToTrack: %w", err)
		}
	}
	// insert primary alias
	err = qtx.InsertTrackAlias(ctx, repository.InsertTrackAliasParams{
		TrackID:   trackRow.ID,
		Alias:     opts.Title,
		Source:    "Canonical",
		IsPrimary: true,
	})
	if err != nil {
		return nil, fmt.Errorf("SaveTrack: InsertTrackAlias: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("SaveTrack: Commit: %w", err)
	}
	return &models.Track{
		ID:       trackRow.ID,
		MbzID:    insertMbzID,
		Title:    opts.Title,
		Duration: opts.Duration,
	}, nil
}

func (d *Psql) UpdateTrack(ctx context.Context, opts db.UpdateTrackOpts) error {
	l := logger.FromContext(ctx)
	if opts.ID == 0 {
		return errors.New("UpdateTrack: track id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("UpdateTrack: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	if opts.MusicBrainzID != uuid.Nil {
		l.Debug().Msgf("Updating MusicBrainz ID for track %d", opts.ID)
		err := qtx.UpdateTrackMbzID(ctx, repository.UpdateTrackMbzIDParams{
			ID:            opts.ID,
			MusicBrainzID: &opts.MusicBrainzID,
		})
		if err != nil {
			return fmt.Errorf("UpdateTrack: UpdateTrackMbzID: %w", err)
		}
	}
	if opts.Duration != 0 {
		l.Debug().Msgf("Updating duration for track %d", opts.ID)
		err := qtx.UpdateTrackDuration(ctx, repository.UpdateTrackDurationParams{
			ID:       opts.ID,
			Duration: opts.Duration,
		})
		if err != nil {
			return fmt.Errorf("UpdateTrack: UpdateTrackDuration: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) SaveTrackAliases(ctx context.Context, id int32, aliases []string, source string) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("SaveTrackAliases: track id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SaveTrackAliases: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	existing, err := qtx.GetAllTrackAliases(ctx, id)
	if err != nil {
		return fmt.Errorf("SaveTrackAliases: GetAllTrackAliases: %w", err)
	}
	for _, v := range existing {
		aliases = append(aliases, v.Alias)
	}
	utils.Unique(&aliases)
	for _, alias := range aliases {
		if strings.TrimSpace(alias) == "" {
			return errors.New("aliases cannot be blank")
		}
		err = qtx.InsertTrackAlias(ctx, repository.InsertTrackAliasParams{
			Alias:     strings.TrimSpace(alias),
			TrackID:   id,
			Source:    source,
			IsPrimary: false,
		})
		if err != nil {
			return fmt.Errorf("SaveTrackAliases: InsertTrackAlias: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) DeleteTrack(ctx context.Context, id int32) error {
	return d.q.DeleteTrack(ctx, id)
}

func (d *Psql) DeleteTrackAlias(ctx context.Context, id int32, alias string) error {
	return d.q.DeleteTrackAlias(ctx, repository.DeleteTrackAliasParams{
		TrackID: id,
		Alias:   alias,
	})
}

func (d *Psql) GetAllTrackAliases(ctx context.Context, id int32) ([]models.Alias, error) {
	rows, err := d.q.GetAllTrackAliases(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetAllTrackAliases: GetAllTrackAliases: %w", err)
	}
	aliases := make([]models.Alias, len(rows))
	for i, row := range rows {
		aliases[i] = models.Alias{
			ID:      id,
			Alias:   row.Alias,
			Source:  row.Source,
			Primary: row.IsPrimary,
		}
	}
	return aliases, nil
}

func (d *Psql) SetPrimaryTrackAlias(ctx context.Context, id int32, alias string) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SetPrimaryTrackAlias: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	// get all aliases
	aliases, err := qtx.GetAllTrackAliases(ctx, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackAlias: GetAllTrackAliases: %w", err)
	}
	primary := ""
	exists := false
	for _, v := range aliases {
		if v.Alias == alias {
			exists = true
		}
		if v.IsPrimary {
			primary = v.Alias
		}
	}
	if primary == alias {
		// no-op rename
		return nil
	}
	if !exists {
		return errors.New("alias does not exist")
	}
	err = qtx.SetTrackAliasPrimaryStatus(ctx, repository.SetTrackAliasPrimaryStatusParams{
		TrackID:   id,
		Alias:     alias,
		IsPrimary: true,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackAlias: SetTrackAliasPrimaryStatus: %w", err)
	}
	err = qtx.SetTrackAliasPrimaryStatus(ctx, repository.SetTrackAliasPrimaryStatusParams{
		TrackID:   id,
		Alias:     primary,
		IsPrimary: false,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackAlias: SetTrackAliasPrimaryStatus: %w", err)
	}
	return tx.Commit(ctx)
}

func (d *Psql) SetPrimaryTrackArtist(ctx context.Context, id int32, artistId int32, value bool) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SetPrimaryTrackArtist: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	// get all artists
	artists, err := qtx.GetTrackArtists(ctx, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackArtist: GetTrackArtists: %w", err)
	}
	var primary int32
	for _, v := range artists {
		// i dont get it??? is_primary is not a nullable column??? why use pgtype.Bool???
		// why not just use boolean??? is sqlc stupid??? am i stupid???????
		if v.IsPrimary.Valid && v.IsPrimary.Bool {
			primary = v.ID
		}
	}
	if value && primary == artistId {
		// no-op
		return nil
	}
	l.Debug().Msgf("Marking artist with id %d as 'primary = %v' on track with id %d", artistId, value, id)
	err = qtx.UpdateTrackPrimaryArtist(ctx, repository.UpdateTrackPrimaryArtistParams{
		TrackID:   id,
		ArtistID:  artistId,
		IsPrimary: value,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryTrackArtist: UpdateTrackPrimaryArtist: %w", err)
	}
	if value && primary != 0 {
		l.Debug().Msgf("Unmarking artist with id %d as primary on track with id %d", primary, id)
		// if we were marking a new one as primary and there was already one marked as primary,
		// unmark that one as there can only be one
		err = qtx.UpdateTrackPrimaryArtist(ctx, repository.UpdateTrackPrimaryArtistParams{
			TrackID:   id,
			ArtistID:  primary,
			IsPrimary: false,
		})
		if err != nil {
			return fmt.Errorf("SetPrimaryTrackArtist: UpdateTrackPrimaryArtist: %w", err)
		}
	}
	return tx.Commit(ctx)
}
