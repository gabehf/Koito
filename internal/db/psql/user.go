package psql

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/logger"
	"github.com/gabehf/koito/internal/models"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

// Returns nil, nil when no database entries are found
func (d *Psql) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	row, err := d.q.GetUserByUsername(ctx, strings.ToLower(username))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("GetUserByUsername: %w", err)
	}
	return &models.User{
		ID:       row.ID,
		Username: row.Username,
		Password: row.Password,
		Role:     models.UserRole(row.Role),
	}, nil
}

// Returns nil, nil when no database entries are found
func (d *Psql) GetUserByApiKey(ctx context.Context, key string) (*models.User, error) {
	row, err := d.q.GetUserByApiKey(ctx, key)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("GetUserByApiKey: %w", err)
	}
	return &models.User{
		ID:       row.ID,
		Username: row.Username,
		Password: row.Password,
		Role:     models.UserRole(row.Role),
	}, nil
}

func (d *Psql) SaveUser(ctx context.Context, opts db.SaveUserOpts) (*models.User, error) {
	l := logger.FromContext(ctx)
	err := ValidateUsername(opts.Username)
	if err != nil {
		l.Debug().AnErr("validator_notice", err).Msgf("Username failed validation: %s", opts.Username)
		return nil, fmt.Errorf("SaveUser: ValidateUsername: %w", err)
	}
	pw, err := ValidateAndNormalizePassword(opts.Password)
	if err != nil {
		l.Debug().AnErr("validator_notice", err).Msgf("Password failed validation")
		return nil, fmt.Errorf("SaveUser: ValidateAndNormalizePassword: %w", err)
	}
	if opts.Role == "" {
		opts.Role = models.UserRoleUser
	}
	hashPw, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		l.Err(err).Msg("Failed to generate hashed password")
		return nil, fmt.Errorf("SaveUser: bcrypt.GenerateFromPassword: %w", err)
	}
	u, err := d.q.InsertUser(ctx, repository.InsertUserParams{
		Username: strings.ToLower(opts.Username),
		Password: hashPw,
		Role:     repository.Role(opts.Role),
	})
	if err != nil {
		return nil, fmt.Errorf("SaveUser: InsertUser: %w", err)
	}
	return &models.User{
		ID:       u.ID,
		Username: u.Username,
		Role:     models.UserRole(u.Role),
	}, nil
}
func (d *Psql) SaveApiKey(ctx context.Context, opts db.SaveApiKeyOpts) (*models.ApiKey, error) {
	row, err := d.q.InsertApiKey(ctx, repository.InsertApiKeyParams{
		Key:    opts.Key,
		Label:  opts.Label,
		UserID: opts.UserID,
	})
	if err != nil {
		return nil, fmt.Errorf("SaveApiKey: InsertApiKey: %w", err)
	}
	return &models.ApiKey{
		ID:        row.ID,
		UserID:    row.UserID,
		Key:       row.Key,
		Label:     row.Label,
		CreatedAt: row.CreatedAt.Time,
	}, nil
}

func (d *Psql) UpdateUser(ctx context.Context, opts db.UpdateUserOpts) error {
	l := logger.FromContext(ctx)
	if opts.ID == 0 {
		return errors.New("user id is required")
	}
	tx, err := d.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		l.Err(err).Msg("Failed to begin transaction")
		return fmt.Errorf("UpdateUser: BeginTx: %w", err)
	}
	defer tx.Rollback(ctx)
	qtx := d.q.WithTx(tx)
	if opts.Username != "" {
		err := ValidateUsername(opts.Username)
		if err != nil {
			l.Debug().AnErr("validator_notice", err).Msgf("Username failed validation: %s", opts.Username)
			return fmt.Errorf("UpdateUser: ValidateUsername: %w", err)
		}
		err = qtx.UpdateUserUsername(ctx, repository.UpdateUserUsernameParams{
			ID:       opts.ID,
			Username: strings.ToLower(opts.Username),
		})
		if err != nil {
			return fmt.Errorf("UpdateUser: UpdateUserUsername: %w", err)
		}
	}
	if opts.Password != "" {
		pw, err := ValidateAndNormalizePassword(opts.Password)
		if err != nil {
			l.Debug().AnErr("validator_notice", err).Msgf("Password failed validation")
			return fmt.Errorf("UpdateUser: ValidateAndNormalizePassword: %w", err)
		}
		hashPw, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			l.Err(err).Msg("Failed to generate hashed password")
			return fmt.Errorf("UpdateUser: bcrypt.GenerateFromPassword: %w", err)
		}
		err = qtx.UpdateUserPassword(ctx, repository.UpdateUserPasswordParams{
			ID:       opts.ID,
			Password: hashPw,
		})
		if err != nil {
			return fmt.Errorf("UpdateUser: UpdateUserPassword: %w", err)
		}
	}
	return tx.Commit(ctx)
}

func (d *Psql) GetApiKeysByUserID(ctx context.Context, id int32) ([]models.ApiKey, error) {
	rows, err := d.q.GetAllApiKeysByUserID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("GetApiKeysByUserID: %w", err)
	}
	keys := make([]models.ApiKey, len(rows))
	for i, row := range rows {
		keys[i] = models.ApiKey{
			ID:     row.ID,
			Key:    row.Key,
			Label:  row.Label,
			UserID: row.UserID,
		}
	}
	return keys, nil
}

func (d *Psql) UpdateApiKeyLabel(ctx context.Context, opts db.UpdateApiKeyLabelOpts) error {
	return d.q.UpdateApiKeyLabel(ctx, repository.UpdateApiKeyLabelParams{
		ID:     opts.ID,
		Label:  opts.Label,
		UserID: opts.UserID,
	})
}

func (d *Psql) DeleteApiKey(ctx context.Context, id int32) error {
	return d.q.DeleteApiKey(ctx, id)
}

func (d *Psql) CountUsers(ctx context.Context) (int64, error) {
	return d.q.CountUsers(ctx)
}

const (
	maxUsernameLength = 32
	minUsernameLength = 1
	maxPasswordLength = 128
	minPasswordLength = 8
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func ValidateUsername(username string) error {
	length := utf8.RuneCountInString(username)
	if length < minUsernameLength || length > maxUsernameLength {
		return errors.New("username must be between 1 and 32 characters")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain [a-zA-Z0-9_.-]")
	}
	return nil
}

func ValidateAndNormalizePassword(password string) (string, error) {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength {
		return "", errors.New("password must be at least 8 characters long")
	}
	if length > maxPasswordLength {
		var truncated []rune
		for i, r := range password {
			if i >= maxPasswordLength {
				break
			}
			truncated = append(truncated, r)
		}
		password = string(truncated)
	}
	return password, nil
}
