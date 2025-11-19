package psql

import (
	"context"
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
	"github.com/jackc/pgx/v5/pgtype"
)

// this function sucks because sqlc keeps making new types for rows that are the same
func (d *Psql) GetArtist(ctx context.Context, opts db.GetArtistOpts) (*models.Artist, error) {
	l := logger.FromContext(ctx)
	if opts.ID != 0 {
		l.Debug().Msgf("Fetching artist from DB with id %d", opts.ID)
		row, err := d.q.GetArtist(ctx, opts.ID)
		if err != nil {
			return nil, fmt.Errorf("GetArtist: GetArtist by ID: %w", err)
		}
		count, err := d.q.CountListensFromArtist(ctx, repository.CountListensFromArtistParams{
			ListenedAt:   time.Unix(0, 0),
			ListenedAt_2: time.Now(),
			ArtistID:     row.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetArtist: CountListensFromArtist: %w", err)
		}
		seconds, err := d.CountTimeListenedToItem(ctx, db.TimeListenedOpts{
			Period:   db.PeriodAllTime,
			ArtistID: row.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetArtist: CountTimeListenedToItem: %w", err)
		}
		firstListen, err := d.q.GetFirstListenFromArtist(ctx, row.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAlbum: GetFirstListenFromArtist: %w", err)
		}
		return &models.Artist{
			ID:           row.ID,
			MbzID:        row.MusicBrainzID,
			Name:         row.Name,
			Aliases:      row.Aliases,
			Image:        row.Image,
			ListenCount:  count,
			TimeListened: seconds,
			FirstListen:  firstListen.ListenedAt.Unix(),
		}, nil
	} else if opts.MusicBrainzID != uuid.Nil {
		l.Debug().Msgf("Fetching artist from DB with MusicBrainz ID %s", opts.MusicBrainzID)
		row, err := d.q.GetArtistByMbzID(ctx, &opts.MusicBrainzID)
		if err != nil {
			return nil, fmt.Errorf("GetArtist: GetArtistByMbzID: %w", err)
		}
		count, err := d.q.CountListensFromArtist(ctx, repository.CountListensFromArtistParams{
			ListenedAt:   time.Unix(0, 0),
			ListenedAt_2: time.Now(),
			ArtistID:     row.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetArtist: CountListensFromArtist: %w", err)
		}
		seconds, err := d.CountTimeListenedToItem(ctx, db.TimeListenedOpts{
			Period:   db.PeriodAllTime,
			ArtistID: row.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetArtist: CountTimeListenedToItem: %w", err)
		}
		firstListen, err := d.q.GetFirstListenFromArtist(ctx, row.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAlbum: GetFirstListenFromArtist: %w", err)
		}
		return &models.Artist{
			ID:           row.ID,
			MbzID:        row.MusicBrainzID,
			Name:         row.Name,
			Aliases:      row.Aliases,
			Image:        row.Image,
			ListenCount:  count,
			TimeListened: seconds,
			FirstListen:  firstListen.ListenedAt.Unix(),
		}, nil
	} else if opts.Name != "" {
		l.Debug().Msgf("Fetching artist from DB with name '%s'", opts.Name)
		row, err := d.q.GetArtistByName(ctx, opts.Name)
		if err != nil {
			return nil, fmt.Errorf("GetArtist: GetArtistByName: %w", err)
		}
		count, err := d.q.CountListensFromArtist(ctx, repository.CountListensFromArtistParams{
			ListenedAt:   time.Unix(0, 0),
			ListenedAt_2: time.Now(),
			ArtistID:     row.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetArtist: CountListensFromArtist: %w", err)
		}
		seconds, err := d.CountTimeListenedToItem(ctx, db.TimeListenedOpts{
			Period:   db.PeriodAllTime,
			ArtistID: row.ID,
		})
		if err != nil {
			return nil, fmt.Errorf("GetArtist: CountTimeListenedToItem: %w", err)
		}
		firstListen, err := d.q.GetFirstListenFromArtist(ctx, row.ID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("GetAlbum: GetFirstListenFromArtist: %w", err)
		}
		return &models.Artist{
			ID:           row.ID,
			MbzID:        row.MusicBrainzID,
			Name:         row.Name,
			Aliases:      row.Aliases,
			Image:        row.Image,
			ListenCount:  count,
			TimeListened: seconds,
			FirstListen:  firstListen.ListenedAt.Unix(),
		}, nil
	} else {
		return nil, errors.New("insufficient information to get artist")
	}
}

// Inserts all unique aliases into the DB with specified source
func (d *Psql) SaveArtistAliases(ctx context.Context, id int32, aliases []string, source string) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("SaveArtistAliases: artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SaveArtistAliases: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	l.Debug().Msgf("Fetching existing artist aliases for artist %d...", id)
	existing, err := qtx.GetAllArtistAliases(ctx, id)
	if err != nil {
		return fmt.Errorf("SaveArtistAliases: GetAllArtistAliases: %w", err)
	}
	for _, v := range existing {
		aliases = append(aliases, v.Alias)
	}
	l.Debug().Msgf("Ensuring aliases are unique...")
	utils.Unique(&aliases)
	for _, alias := range aliases {
		l.Debug().Msgf("Inserting alias %s for artist with id %d", alias, id)
		alias = strings.TrimSpace(alias)
		if alias == "" {
			return errors.New("SaveArtistAliases: aliases cannot be blank")
		}
		err = qtx.InsertArtistAlias(ctx, repository.InsertArtistAliasParams{
			Alias:     alias,
			ArtistID:  id,
			Source:    source,
			IsPrimary: false,
		})
		if err != nil {
			return fmt.Errorf("SaveArtistAliases: InsertArtistAlias: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) DeleteArtist(ctx context.Context, id int32) error {
	return d.q.DeleteArtist(ctx, id)
}

// Equivalent to Psql.SaveArtist, then Psql.SaveMbzAliases
func (d *Psql) SaveArtist(ctx context.Context, opts db.SaveArtistOpts) (*models.Artist, error) {
	l := logger.FromContext(ctx)
	var insertMbzID *uuid.UUID
	var insertImage *uuid.UUID
	if opts.MusicBrainzID != uuid.Nil {
		insertMbzID = &opts.MusicBrainzID
	}
	if opts.Image != uuid.Nil {
		insertImage = &opts.Image
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return nil, fmt.Errorf("SaveArtist: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	opts.Name = strings.TrimSpace(opts.Name)
	if opts.Name == "" {
		return nil, errors.New("SaveArtist: name must not be blank")
	}
	l.Debug().Msgf("Inserting artist '%s' into DB", opts.Name)
	a, err := qtx.InsertArtist(ctx, repository.InsertArtistParams{
		MusicBrainzID: insertMbzID,
		Image:         insertImage,
		ImageSource:   pgtype.Text{String: opts.ImageSrc, Valid: opts.ImageSrc != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("SaveArtist: InsertArtist: %w", err)
	}
	l.Debug().Msgf("Inserting canonical alias '%s' into DB for artist with id %d", opts.Name, a.ID)
	err = qtx.InsertArtistAlias(ctx, repository.InsertArtistAliasParams{
		ArtistID:  a.ID,
		Alias:     opts.Name,
		Source:    "Canonical",
		IsPrimary: true,
	})
	if err != nil {
		l.Err(err).Msgf("SaveArtist: error inserting canonical alias for artist '%s'", opts.Name)
		return nil, fmt.Errorf("SaveArtist: InsertArtistAlias: %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		l.Err(err).Msg("Failed to commit insert artist transaction")
		return nil, fmt.Errorf("SaveArtist: Commit: %w", err)
	}
	artist := &models.Artist{
		ID:      a.ID,
		Name:    opts.Name,
		Image:   a.Image,
		MbzID:   a.MusicBrainzID,
		Aliases: []string{opts.Name},
	}
	if len(opts.Aliases) > 0 {
		l.Debug().Msgf("Inserting aliases '%v' into DB for artist '%s'", opts.Aliases, opts.Name)
		err = d.SaveArtistAliases(ctx, a.ID, opts.Aliases, "MusicBrainz")
		if err != nil {
			return nil, fmt.Errorf("SaveArtist: SaveArtistAliases: %w", err)
		}
		artist.Aliases = opts.Aliases
	}
	return artist, nil
}

func (d *Psql) UpdateArtist(ctx context.Context, opts db.UpdateArtistOpts) error {
	l := logger.FromContext(ctx)
	if opts.ID == 0 {
		return errors.New("UpdateArtist: artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("UpdateArtist: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	if opts.MusicBrainzID != uuid.Nil {
		l.Debug().Msgf("Updating artist with id %d with MusicBrainz ID %s", opts.ID, opts.MusicBrainzID)
		err := qtx.UpdateArtistMbzID(ctx, repository.UpdateArtistMbzIDParams{
			ID:            opts.ID,
			MusicBrainzID: &opts.MusicBrainzID,
		})
		if err != nil {
			return fmt.Errorf("UpdateArtist: UpdateArtistMbzID: %w", err)
		}
	}
	if opts.Image != uuid.Nil {
		l.Debug().Msgf("Updating artist with id %d with image %s", opts.ID, opts.Image)
		err = qtx.UpdateArtistImage(ctx, repository.UpdateArtistImageParams{
			ID:          opts.ID,
			Image:       &opts.Image,
			ImageSource: pgtype.Text{String: opts.ImageSrc, Valid: opts.ImageSrc != ""},
		})
		if err != nil {
			return fmt.Errorf("UpdateArtist: UpdateArtistImage: %w", err)
		}
	}
	err = tx.Commit(ctx)
	if err != nil {
		l.Err(err).Msg("Failed to commit update artist transaction")
		return fmt.Errorf("UpdateArtist: Commit: %w", err)
	}
	return nil
}

func (d *Psql) DeleteArtistAlias(ctx context.Context, id int32, alias string) error {
	return d.q.DeleteArtistAlias(ctx, repository.DeleteArtistAliasParams{
		ArtistID: id,
		Alias:    alias,
	})
}

func (d *Psql) GetAllArtistAliases(ctx context.Context, id int32) ([]models.Alias, error) {
	rows, err := d.q.GetAllArtistAliases(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetAllArtistAliases: %w", err)
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

func (d *Psql) SetPrimaryArtistAlias(ctx context.Context, id int32, alias string) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("SetPrimaryArtistAlias: artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SetPrimaryArtistAlias: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	aliases, err := qtx.GetAllArtistAliases(ctx, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryArtistAlias: GetAllArtistAliases: %w", err)
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
		return nil
	}
	if !exists {
		return errors.New("SetPrimaryArtistAlias: alias does not exist")
	}
	err = qtx.SetArtistAliasPrimaryStatus(ctx, repository.SetArtistAliasPrimaryStatusParams{
		ArtistID:  id,
		Alias:     alias,
		IsPrimary: true,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryArtistAlias: SetArtistAliasPrimaryStatus (primary): %w", err)
	}
	err = qtx.SetArtistAliasPrimaryStatus(ctx, repository.SetArtistAliasPrimaryStatusParams{
		ArtistID:  id,
		Alias:     primary,
		IsPrimary: false,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryArtistAlias: SetArtistAliasPrimaryStatus (previous primary): %w", err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		l.Err(err).Msg("Failed to commit transaction")
		return fmt.Errorf("SetPrimaryArtistAlias: Commit: %w", err)
	}
	return nil
}
func (d *Psql) GetArtistsForAlbum(ctx context.Context, id int32) ([]*models.Artist, error) {
	l := logger.FromContext(ctx)
	l.Debug().Msgf("Fetching artists for album ID %d", id)

	rows, err := d.q.GetReleaseArtists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetArtistsForAlbum: %w", err)
	}

	artists := make([]*models.Artist, len(rows))
	for i, row := range rows {
		artists[i] = &models.Artist{
			ID:        row.ID,
			Name:      row.Name,
			MbzID:     row.MusicBrainzID,
			Image:     row.Image,
			IsPrimary: row.IsPrimary.Valid && row.IsPrimary.Bool,
		}
	}

	return artists, nil
}

func (d *Psql) GetArtistsForTrack(ctx context.Context, id int32) ([]*models.Artist, error) {
	l := logger.FromContext(ctx)
	l.Debug().Msgf("Fetching artists for track ID %d", id)

	rows, err := d.q.GetTrackArtists(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetArtistsForTrack: %w", err)
	}

	artists := make([]*models.Artist, len(rows))
	for i, row := range rows {
		artists[i] = &models.Artist{
			ID:        row.ID,
			Name:      row.Name,
			MbzID:     row.MusicBrainzID,
			Image:     row.Image,
			IsPrimary: row.IsPrimary.Valid && row.IsPrimary.Bool,
		}
	}

	return artists, nil
}
