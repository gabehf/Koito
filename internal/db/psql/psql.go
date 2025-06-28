// package psql implements the db.DB interface using psx and a sql generated repository
package psql

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/gabehf/koito/db/migrations"
	"github.com/gabehf/koito/internal/cfg"
	"github.com/gabehf/koito/internal/db"
	"github.com/gabehf/koito/internal/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
)

const (
	DefaultItemsPerPage = 20
)

type Psql struct {
	q    *repository.Queries
	conn *pgxpool.Pool
}

func New() (*Psql, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	config, err := pgxpool.ParseConfig(cfg.DatabaseUrl())
	if err != nil {
		return nil, fmt.Errorf("psql.New: failed to parse pgx config: %w", err)
	}

	config.ConnConfig.ConnectTimeout = 15 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("psql.New: failed to create pgx pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("psql.New: database not reachable: %w", err)
	}

	sqlDB, err := sql.Open("pgx", cfg.DatabaseUrl())
	if err != nil {
		return nil, fmt.Errorf("psql.New: failed to open db for migrations: %w", err)
	}

	goose.SetBaseFS(migrations.Files)

	if err := goose.Up(sqlDB, "."); err != nil {
		return nil, fmt.Errorf("psql.New: goose failed: %w", err)
	}
	_ = sqlDB.Close()

	return &Psql{
		q:    repository.New(pool),
		conn: pool,
	}, nil
}

// Not part of the DB interface this package implements. Only used for testing.
func (d *Psql) Exec(ctx context.Context, query string, args ...any) error {
	_, err := d.conn.Exec(ctx, query, args...)
	return err
}

// Not part of the DB interface this package implements. Only used for testing.
func (d *Psql) RowExists(ctx context.Context, query string, args ...any) (bool, error) {
	var exists bool
	err := d.conn.QueryRow(ctx, query, args...).Scan(&exists)
	return exists, err
}

func (p *Psql) Count(ctx context.Context, query string, args ...any) (count int, err error) {
	err = p.conn.QueryRow(ctx, query, args...).Scan(&count)
	return
}

// Exposes p.conn.QueryRow. Only used for testing. Not part of the DB interface this package implements.
func (p *Psql) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	return p.conn.QueryRow(ctx, query, args...)
}

func (d *Psql) Close(ctx context.Context) {
	d.conn.Close()
}

func (d *Psql) Ping(ctx context.Context) error {
	return d.conn.Ping(ctx)
}

func stepToInterval(p db.StepInterval) pgtype.Interval {
	var interval pgtype.Interval
	switch p {
	case db.StepDay:
		interval.Days = 1
	case db.StepWeek:
		interval.Days = 7
	case db.StepMonth:
		interval.Months = 1
	case db.StepYear:
		interval.Months = 12
	}
	interval.Valid = true
	return interval
}
