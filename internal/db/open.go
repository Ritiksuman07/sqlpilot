package db

import (
	"errors"
	"strings"
)

func Open(dsn string) (Connector, error) {
	dsn = strings.TrimSpace(dsn)
	if dsn == "" {
		return nil, errors.New("empty DSN")
	}

	lower := strings.ToLower(dsn)
	switch {
	case strings.HasPrefix(lower, "postgres://") || strings.HasPrefix(lower, "postgresql://"):
		return NewPostgresConnector(dsn)
	case strings.HasPrefix(lower, "sqlite:") || strings.HasPrefix(lower, "file:") || isLikelySQLitePath(lower):
		return NewSQLiteConnector(dsn)
	default:
		return nil, errors.New("unsupported DSN: only postgres and sqlite are supported in v0.1")
	}
}

func isLikelySQLitePath(dsn string) bool {
	return strings.HasSuffix(dsn, ".db") || strings.HasSuffix(dsn, ".sqlite") || strings.HasSuffix(dsn, ".sqlite3") || dsn == ":memory:"
}
