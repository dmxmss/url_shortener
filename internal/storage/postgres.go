package storage

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrConflict = errors.New("short code already exists")
	ErrNotFound = errors.New("short code not found")
)

type Store struct {
	pool *pgxpool.Pool
}

func New(ctx context.Context, dsn string) (*Store, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 8
	cfg.MinConns = 1
	cfg.MaxConnLifetime = 30 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return &Store{pool: pool}, nil
}

func (s *Store) CreateLink(ctx context.Context, code, longURL string) error {
	_, err := s.pool.Exec(ctx, `insert into links(short_code, long_url) values($1, $2)`, code, longURL)
	if err == nil {
		return nil
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

func (s *Store) GetURL(ctx context.Context, code string) (string, error) {
	var longURL string
	err := s.pool.QueryRow(ctx, `select long_url from links where short_code = $1`, code).Scan(&longURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", ErrNotFound
	}
	return longURL, err
}

func (s *Store) RecordRedirect(ctx context.Context, code string, at time.Time) error {
	_, err := s.pool.Exec(ctx, `
		update links
		set redirect_count = redirect_count + 1, last_accessed_at = $2
		where short_code = $1
	`, code, at)
	return err
}

func (s *Store) Ping(ctx context.Context) error {
	return s.pool.Ping(ctx)
}

func (s *Store) Close() {
	s.pool.Close()
}
