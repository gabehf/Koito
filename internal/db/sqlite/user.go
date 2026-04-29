package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/models"
	"golang.org/x/crypto/bcrypt"
)

const (
	maxUsernameLength = 32
	minUsernameLength = 1
	maxPasswordLength = 128
	minPasswordLength = 8
)

var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_.-]+$`)

func validateUsername(username string) error {
	length := utf8.RuneCountInString(username)
	if length < minUsernameLength || length > maxUsernameLength {
		return errors.New("username must be between 1 and 32 characters")
	}
	if !usernameRegex.MatchString(username) {
		return errors.New("username can only contain [a-zA-Z0-9_.-]")
	}
	return nil
}

func validateAndNormalizePassword(password string) (string, error) {
	length := utf8.RuneCountInString(password)
	if length < minPasswordLength {
		return "", errors.New("password must be at least 8 characters long")
	}
	if length > maxPasswordLength {
		runes := []rune(password)
		password = string(runes[:maxPasswordLength])
	}
	return password, nil
}

func (s *Sqlite) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var u models.User
	var role string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, role, password FROM users WHERE username = ?`,
		strings.ToLower(username)).Scan(&u.ID, &u.Username, &role, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("GetUserByUsername: %w", err)
	}
	u.Role = models.UserRole(role)
	return &u, nil
}

func (s *Sqlite) GetAdminUser(ctx context.Context) (*models.User, error) {
	var u models.User
	var role string
	err := s.db.QueryRowContext(ctx,
		`SELECT id, username, role, password FROM users WHERE role = 'admin' LIMIT 1`,
	).Scan(&u.ID, &u.Username, &role, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("GetAdminUser: %w", err)
	}
	u.Role = models.UserRole(role)
	return &u, nil
}

func (s *Sqlite) GetUserByApiKey(ctx context.Context, key string) (*models.User, error) {
	var u models.User
	var role string
	err := s.db.QueryRowContext(ctx, `
		SELECT u.id, u.username, u.role, u.password
		FROM users u JOIN api_keys ak ON u.id = ak.user_id
		WHERE ak.key = ? LIMIT 1`, key).Scan(&u.ID, &u.Username, &role, &u.Password)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("GetUserByApiKey: %w", err)
	}
	u.Role = models.UserRole(role)
	return &u, nil
}

func (s *Sqlite) SaveUser(ctx context.Context, opts db.SaveUserOpts) (*models.User, error) {
	if err := validateUsername(opts.Username); err != nil {
		return nil, fmt.Errorf("SaveUser: %w", err)
	}
	pw, err := validateAndNormalizePassword(opts.Password)
	if err != nil {
		return nil, fmt.Errorf("SaveUser: %w", err)
	}
	if opts.Role == "" {
		opts.Role = models.UserRoleUser
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("SaveUser: bcrypt: %w", err)
	}
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO users (username, role, password) VALUES (?,?,?)`,
		strings.ToLower(opts.Username), string(opts.Role), hash,
	)
	if err != nil {
		return nil, fmt.Errorf("SaveUser: insert: %w", err)
	}
	id64, _ := res.LastInsertId()
	return &models.User{
		ID:       int32(id64),
		Username: strings.ToLower(opts.Username),
		Role:     opts.Role,
	}, nil
}

func (s *Sqlite) UpdateUser(ctx context.Context, opts db.UpdateUserOpts) error {
	if opts.ID == 0 {
		return errors.New("UpdateUser: user id is required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("UpdateUser: BeginTx: %w", err)
	}
	defer tx.Rollback()

	if opts.Username != "" {
		if err := validateUsername(opts.Username); err != nil {
			return fmt.Errorf("UpdateUser: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET username = ? WHERE id = ?`,
			strings.ToLower(opts.Username), opts.ID); err != nil {
			return fmt.Errorf("UpdateUser: username: %w", err)
		}
	}
	if opts.Password != "" {
		pw, err := validateAndNormalizePassword(opts.Password)
		if err != nil {
			return fmt.Errorf("UpdateUser: %w", err)
		}
		hash, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("UpdateUser: bcrypt: %w", err)
		}
		if _, err := tx.ExecContext(ctx,
			`UPDATE users SET password = ? WHERE id = ?`, hash, opts.ID); err != nil {
			return fmt.Errorf("UpdateUser: password: %w", err)
		}
	}
	return tx.Commit()
}

func (s *Sqlite) SaveApiKey(ctx context.Context, opts db.SaveApiKeyOpts) (*models.ApiKey, error) {
	now := time.Now().Unix()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO api_keys (key, user_id, created_at, label) VALUES (?,?,?,?)`,
		opts.Key, opts.UserID, now, opts.Label,
	)
	if err != nil {
		return nil, fmt.Errorf("SaveApiKey: %w", err)
	}
	id64, _ := res.LastInsertId()
	return &models.ApiKey{
		ID:        int32(id64),
		Key:       opts.Key,
		UserID:    opts.UserID,
		Label:     opts.Label,
		CreatedAt: time.Unix(now, 0).UTC(),
	}, nil
}

func (s *Sqlite) GetApiKeysByUserID(ctx context.Context, id int32) ([]models.ApiKey, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, key, user_id, label, created_at FROM api_keys WHERE user_id = ?`, id)
	if err != nil {
		return nil, fmt.Errorf("GetApiKeysByUserID: %w", err)
	}
	defer rows.Close()
	var keys []models.ApiKey
	for rows.Next() {
		var k models.ApiKey
		var createdAt int64
		if err := rows.Scan(&k.ID, &k.Key, &k.UserID, &k.Label, &createdAt); err != nil {
			return nil, err
		}
		k.CreatedAt = time.Unix(createdAt, 0).UTC()
		keys = append(keys, k)
	}
	return keys, rows.Err()
}

func (s *Sqlite) UpdateApiKeyLabel(ctx context.Context, opts db.UpdateApiKeyLabelOpts) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE api_keys SET label = ? WHERE id = ? AND user_id = ?`,
		opts.Label, opts.ID, opts.UserID)
	return err
}

func (s *Sqlite) DeleteApiKey(ctx context.Context, id int32) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM api_keys WHERE id = ?`, id)
	return err
}

func (s *Sqlite) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM users`).Scan(&count)
	return count, err
}
