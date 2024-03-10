package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"url_shortner/internal/models"
	"url_shortner/internal/storage"
)

type sqliteStorage struct {
	db *sql.DB
}

func NewSqliteStorage(storagePath string) (*sqliteStorage, error) {
	const op = "storage.sqlite.New"

	db, err := sql.Open("sqlite3", storagePath)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)

	}
	stmt, err := db.Prepare(`
		CREATE TABLE IF NOT EXISTS url(
			id INTEGER PRIMARY KEY,
			alias TEXT NOT NULL UNIQUE,
			url TEXT NOT NULL);
		CREATE INDEX IF NOT EXISTS idx_alias ON url(alias);
		`)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &sqliteStorage{db: db}, nil

}

func (s *sqliteStorage) SaveURL(urlToSave string, alias string) (int64, error) {
	const op = "storage.sqlite.SaveURL"
	stmt, err := s.db.Prepare("INSERT INTO url(url, alias) VALUES (?, ?)")
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && errors.Is(sqliteErr.ExtendedCode, sqlite3.ErrConstraintUnique) {
			return 0, fmt.Errorf("%s: %w", op, storage.ErrURLExists)
		}
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	res, err := stmt.Exec(urlToSave, alias)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return id, nil

}

func (s *sqliteStorage) GetURL(Id int64) (models.URL, error) {
	const op = "storage.sqlite.GetURL"
	var url models.URL
	stmt, err := s.db.Prepare("SELECT * FROM url WHERE id = ?")
	if err != nil {
		return models.URL{}, fmt.Errorf("%s: %w", op, err)
	}

	err = stmt.QueryRow(Id).Scan(&url.Id, &url.Alias, &url.Url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.URL{}, fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return models.URL{}, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil

}

func (s *sqliteStorage) GetURLByAlias(Alias string) (models.URL, error) {
	const op = "storage.sqlite.GetURLByAlias"
	var url models.URL
	stmt, err := s.db.Prepare("SELECT * FROM url WHERE alias = ?")
	if err != nil {
		return models.URL{}, fmt.Errorf("%s: %w", op, err)
	}

	err = stmt.QueryRow(Alias).Scan(&url.Id, &url.Alias, &url.Url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.URL{}, fmt.Errorf("%s: %w", op, storage.ErrURLNotFound)
		}
		return models.URL{}, fmt.Errorf("%s: %w", op, err)
	}

	return url, nil

}
