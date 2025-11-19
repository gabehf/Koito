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
	"github.com/jackc/pgx/v5/pgtype"
)

func (d *Psql) GetAlbum(ctx context.Context, opts db.GetAlbumOpts) (*models.Album, error) {
	l := logger.FromContext(ctx)
	var err error
	var ret = new(models.Album)

	if opts.ID != 0 {
		l.Debug().Msgf("Fetching album from DB with id %d", opts.ID)
		row, err := d.q.GetRelease(ctx, opts.ID)
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: %w", err)
		}
		ret.ID = row.ID
		ret.MbzID = row.MusicBrainzID
		ret.Title = row.Title
		ret.Image = row.Image
		ret.VariousArtists = row.VariousArtists
		err = json.Unmarshal(row.Artists, &ret.Artists)
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: json.Unmarshal: %w", err)
		}
	} else if opts.MusicBrainzID != uuid.Nil {
		l.Debug().Msgf("Fetching album from DB with MusicBrainz Release ID %s", opts.MusicBrainzID)
		row, err := d.q.GetReleaseByMbzID(ctx, &opts.MusicBrainzID)
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: %w", err)
		}
		ret.ID = row.ID
		ret.MbzID = row.MusicBrainzID
		ret.Title = row.Title
		ret.Image = row.Image
		ret.VariousArtists = row.VariousArtists
	} else if opts.ArtistID != 0 && opts.Title != "" {
		l.Debug().Msgf("Fetching album from DB with artist_id %d and title %s", opts.ArtistID, opts.Title)
		row, err := d.q.GetReleaseByArtistAndTitle(ctx, repository.GetReleaseByArtistAndTitleParams{
			ArtistID: opts.ArtistID,
			Title:    opts.Title,
		})
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: %w", err)
		}
		ret.ID = row.ID
		ret.MbzID = row.MusicBrainzID
		ret.Title = row.Title
		ret.Image = row.Image
		ret.VariousArtists = row.VariousArtists
	} else if opts.ArtistID != 0 && len(opts.Titles) > 0 {
		l.Debug().Msgf("Fetching release group from DB with artist_id %d and titles %v", opts.ArtistID, opts.Titles)
		row, err := d.q.GetReleaseByArtistAndTitles(ctx, repository.GetReleaseByArtistAndTitlesParams{
			ArtistID: opts.ArtistID,
			Column1:  opts.Titles,
		})
		if err != nil {
			return nil, fmt.Errorf("GetAlbum: %w", err)
		}
		ret.ID = row.ID
		ret.MbzID = row.MusicBrainzID
		ret.Title = row.Title
		ret.Image = row.Image
		ret.VariousArtists = row.VariousArtists
	} else {
		return nil, errors.New("GetAlbum: insufficient information to get album")
	}

	count, err := d.q.CountListensFromRelease(ctx, repository.CountListensFromReleaseParams{
		ListenedAt:   time.Unix(0, 0),
		ListenedAt_2: time.Now(),
		ReleaseID:    ret.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetAlbum: CountListensFromRelease: %w", err)
	}

	seconds, err := d.CountTimeListenedToItem(ctx, db.TimeListenedOpts{
		Period:  db.PeriodAllTime,
		AlbumID: ret.ID,
	})
	if err != nil {
		return nil, fmt.Errorf("GetAlbum: CountTimeListenedToItem: %w", err)
	}

	firstListen, err := d.q.GetFirstListenFromRelease(ctx, ret.ID)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("GetAlbum: GetFirstListenFromRelease: %w", err)
	}

	ret.ListenCount = count
	ret.TimeListened = seconds
	ret.FirstListen = firstListen.ListenedAt.Unix()

	return ret, nil
}

func (d *Psql) SaveAlbum(ctx context.Context, opts db.SaveAlbumOpts) (*models.Album, error) {
	l := logger.FromContext(ctx)
	var insertMbzID *uuid.UUID
	var insertImage *uuid.UUID
	if opts.MusicBrainzID != uuid.Nil {
		insertMbzID = &opts.MusicBrainzID
	}
	if opts.Image != uuid.Nil {
		insertImage = &opts.Image
	}
	if len(opts.ArtistIDs) < 1 {
		return nil, errors.New("SaveAlbum: required parameter 'ArtistIDs' missing")
	}
	for _, aid := range opts.ArtistIDs {
		if aid == 0 {
			return nil, errors.New("SaveAlbum: none of 'ArtistIDs' may be 0")
		}
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return nil, fmt.Errorf("SaveAlbum: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	l.Debug().Msgf("Inserting release '%s' into DB", opts.Title)
	r, err := qtx.InsertRelease(ctx, repository.InsertReleaseParams{
		MusicBrainzID:  insertMbzID,
		VariousArtists: opts.VariousArtists,
		Image:          insertImage,
		ImageSource:    pgtype.Text{String: opts.ImageSrc, Valid: opts.ImageSrc != ""},
	})
	if err != nil {
		return nil, fmt.Errorf("SaveAlbum: InsertRelease: %w", err)
	}
	for _, artistId := range opts.ArtistIDs {
		l.Debug().Msgf("Associating release '%s' to artist with ID %d", opts.Title, artistId)
		err = qtx.AssociateArtistToRelease(ctx, repository.AssociateArtistToReleaseParams{
			ArtistID:  artistId,
			ReleaseID: r.ID,
			IsPrimary: opts.ArtistIDs[0] == artistId,
		})
		if err != nil {
			return nil, fmt.Errorf("SaveAlbum: AssociateArtistToRelease: %w", err)
		}
	}
	l.Debug().Msgf("Saving canonical alias %s for release %d", opts.Title, r.ID)
	err = qtx.InsertReleaseAlias(ctx, repository.InsertReleaseAliasParams{
		ReleaseID: r.ID,
		Alias:     opts.Title,
		Source:    "Canonical",
		IsPrimary: true,
	})
	if err != nil {
		l.Err(err).Msgf("Failed to save canonical alias for album %d", r.ID)
		return nil, fmt.Errorf("SaveAlbum: InsertReleaseAlias: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, fmt.Errorf("SaveAlbum: Commit: %w", err)
	}

	err = d.SaveAlbumAliases(ctx, r.ID, opts.Aliases, "MusicBrainz")
	if err != nil {
		l.Err(err).Msgf("Failed to save aliases for album %s", opts.Title)
	}

	return &models.Album{
		ID:             r.ID,
		MbzID:          r.MusicBrainzID,
		Title:          opts.Title,
		Image:          r.Image,
		VariousArtists: r.VariousArtists,
	}, nil
}

func (d *Psql) AddArtistsToAlbum(ctx context.Context, opts db.AddArtistsToAlbumOpts) error {
	l := logger.FromContext(ctx)
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("AddArtistsToAlbum: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	for _, id := range opts.ArtistIDs {
		err := qtx.AssociateArtistToRelease(ctx, repository.AssociateArtistToReleaseParams{
			ReleaseID: opts.AlbumID,
			ArtistID:  id,
		})
		if err != nil {
			l.Error().Err(err).Msgf("Failed to associate release %d with artist %d", opts.AlbumID, id)
			return fmt.Errorf("AddArtistsToAlbum: AssociateArtistToRelease: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) UpdateAlbum(ctx context.Context, opts db.UpdateAlbumOpts) error {
	l := logger.FromContext(ctx)
	if opts.ID == 0 {
		return errors.New("missing album id")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("UpdateAlbum: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	if opts.MusicBrainzID != uuid.Nil {
		l.Debug().Msgf("Updating release with ID %d with MusicBrainz ID %s", opts.ID, opts.MusicBrainzID)
		err := qtx.UpdateReleaseMbzID(ctx, repository.UpdateReleaseMbzIDParams{
			ID:            opts.ID,
			MusicBrainzID: &opts.MusicBrainzID,
		})
		if err != nil {
			return fmt.Errorf("UpdateAlbum: UpdateReleaseMbzID: %w", err)
		}
	}
	if opts.Image != uuid.Nil {
		l.Debug().Msgf("Updating release with ID %d with image %s", opts.ID, opts.Image)
		err := qtx.UpdateReleaseImage(ctx, repository.UpdateReleaseImageParams{
			ID:          opts.ID,
			Image:       &opts.Image,
			ImageSource: pgtype.Text{String: opts.ImageSrc, Valid: opts.ImageSrc != ""},
		})
		if err != nil {
			return fmt.Errorf("UpdateAlbum: UpdateReleaseImage: %w", err)
		}
	}
	if opts.VariousArtistsUpdate {
		l.Debug().Msgf("Updating release with ID %d with image %s", opts.ID, opts.Image)
		err := qtx.UpdateReleaseVariousArtists(ctx, repository.UpdateReleaseVariousArtistsParams{
			ID:             opts.ID,
			VariousArtists: opts.VariousArtistsValue,
		})
		if err != nil {
			return fmt.Errorf("UpdateAlbum: UpdateReleaseVariousArtists: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) SaveAlbumAliases(ctx context.Context, id int32, aliases []string, source string) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("album id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SaveAlbumAliases: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	existing, err := qtx.GetAllReleaseAliases(ctx, id)
	if err != nil {
		return fmt.Errorf("SaveAlbumAliases: GetAllReleaseAliases: %w", err)
	}
	for _, v := range existing {
		aliases = append(aliases, v.Alias)
	}
	utils.Unique(&aliases)
	for _, alias := range aliases {
		if strings.TrimSpace(alias) == "" {
			return errors.New("SaveAlbumAliases: aliases cannot be blank")
		}
		err = qtx.InsertReleaseAlias(ctx, repository.InsertReleaseAliasParams{
			Alias:     strings.TrimSpace(alias),
			ReleaseID: id,
			Source:    source,
			IsPrimary: false,
		})
		if err != nil {
			return fmt.Errorf("SaveAlbumAliases: InsertReleaseAlias: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) DeleteAlbum(ctx context.Context, id int32) error {
	return d.q.DeleteRelease(ctx, id)
}
func (d *Psql) DeleteAlbumAlias(ctx context.Context, id int32, alias string) error {
	return d.q.DeleteReleaseAlias(ctx, repository.DeleteReleaseAliasParams{
		ReleaseID: id,
		Alias:     alias,
	})
}

func (d *Psql) GetAllAlbumAliases(ctx context.Context, id int32) ([]models.Alias, error) {
	rows, err := d.q.GetAllReleaseAliases(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetAllAlbumAliases: GetAllReleaseAliases: %w", err)
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

func (d *Psql) SetPrimaryAlbumAlias(ctx context.Context, id int32, alias string) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SetPrimaryAlbumAlias: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	// get all aliases
	aliases, err := qtx.GetAllReleaseAliases(ctx, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumAlias: GetAllReleaseAliases: %w", err)
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
		return errors.New("SetPrimaryAlbumAlias: alias does not exist")
	}
	err = qtx.SetReleaseAliasPrimaryStatus(ctx, repository.SetReleaseAliasPrimaryStatusParams{
		ReleaseID: id,
		Alias:     alias,
		IsPrimary: true,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumAlias: SetReleaseAliasPrimaryStatus: %w", err)
	}
	err = qtx.SetReleaseAliasPrimaryStatus(ctx, repository.SetReleaseAliasPrimaryStatusParams{
		ReleaseID: id,
		Alias:     primary,
		IsPrimary: false,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumAlias: SetReleaseAliasPrimaryStatus: %w", err)
	}
	return tx.Commit(ctx)
}

func (d *Psql) SetPrimaryAlbumArtist(ctx context.Context, id int32, artistId int32, value bool) error {
	l := logger.FromContext(ctx)
	if id == 0 {
		return errors.New("artist id not specified")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("SetPrimaryAlbumArtist: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	// get all artists
	artists, err := qtx.GetReleaseArtists(ctx, id)
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumArtist: GetReleaseArtists: %w", err)
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
	l.Debug().Msgf("Marking artist with id %d as 'primary = %v' on album with id %d", artistId, value, id)
	err = qtx.UpdateReleasePrimaryArtist(ctx, repository.UpdateReleasePrimaryArtistParams{
		ReleaseID: id,
		ArtistID:  artistId,
		IsPrimary: value,
	})
	if err != nil {
		return fmt.Errorf("SetPrimaryAlbumArtist: UpdateReleasePrimaryArtist: %w", err)
	}
	if value && primary != 0 {
		// if we were marking a new one as primary and there was already one marked as primary,
		// unmark that one as there can only be one
		l.Debug().Msgf("Unmarking artist with id %d as primary on album with id %d", primary, id)
		err = qtx.UpdateReleasePrimaryArtist(ctx, repository.UpdateReleasePrimaryArtistParams{
			ReleaseID: id,
			ArtistID:  primary,
			IsPrimary: false,
		})
		if err != nil {
			return fmt.Errorf("SetPrimaryAlbumArtist: UpdateReleasePrimaryArtist: %w", err)
		}
	}
	return tx.Commit(ctx)
}
