package history

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db *sql.DB
}

type Entry struct {
	ID        int64
	Query     string
	CreatedAt time.Time
}

func Open() (*Store, error) {
	path, err := historyPath()
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}

	if err := migrate(db); err != nil {
		return nil, err
	}

	return &Store{db: db}, nil
}

func (s *Store) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *Store) Add(query string) error {
	if s.db == nil {
		return nil
	}
	_, err := s.db.Exec(`INSERT INTO history (query) VALUES (?)`, query)
	return err
}

func (s *Store) List(limit int) ([]Entry, error) {
	if s.db == nil {
		return nil, nil
	}
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`
		SELECT id, query, created_at
		FROM history
		ORDER BY id DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var entry Entry
		var ts string
		if err := rows.Scan(&entry.ID, &entry.Query, &ts); err != nil {
			return nil, err
		}
		if parsed, err := time.Parse("2006-01-02 15:04:05", ts); err == nil {
			entry.CreatedAt = parsed
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

func historyPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share", "sqlpilot", "history.db"), nil
}

func migrate(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS history (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			query TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
	`)
	return err
}
