package postgresql

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"url_shortner/internal/models"
	"url_shortner/internal/storage"
)

type PostgresqlStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresqlStorage(dsn string) (*PostgresqlStorage, error) {
	const op = "storage.postgres.New"

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS url(
			id SERIAL PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
	`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &PostgresqlStorage{pool: pool}, nil
}

func (s *PostgresqlStorage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.postgres.SaveURL"
	ctx := context.Background()

	var id int64
	err := s.pool.QueryRow(ctx, "INSERT INTO url(url, alias) VALUES ($1, $2) RETURNING id", urlToSave, alias).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // Код ошибки уникальности в PostgreSQL
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}
	defer s.pool.Close()

	return id, err
}

func (s *PostgresqlStorage) GetURL(Id int64) (models.URL, error) {
	const op = "storage.sqlite.GetURL"
	ctx := context.Background()
	var url models.URL
	err := s.pool.QueryRow(ctx, "SELECT * FROM url WHERE id = $1", Id).Scan(&url.Id, &url.Alias, &url.Url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.URL{}, fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return models.URL{}, fmt.Errorf("%s: %w", op, err)
	}
	defer s.pool.Close()

	return url, nil

}

func (s *PostgresqlStorage) GetURLByAlias(Alias string) (models.URL, error) {
	const op = "storage.sqlite.GetURLByAlias"
	ctx := context.Background()
	var url models.URL
	err := s.pool.QueryRow(ctx, "SELECT * FROM url WHERE alias = $1", Alias).Scan(&url.Id, &url.Alias, &url.Url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.URL{}, fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return models.URL{}, fmt.Errorf("%s: %w", op, err)
	}
	defer s.pool.Close()

	return url, nil

}
