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
	case strings.HasPrefix(lower, "mysql://") || strings.HasPrefix(lower, "mysql:"):
		return NewMySQLConnector(dsn)
	case strings.HasPrefix(lower, "duckdb://") || strings.HasPrefix(lower, "duckdb:") || isLikelyDuckDBPath(lower):
		return NewDuckDBConnector(dsn)
	case strings.HasPrefix(lower, "sqlite:") || strings.HasPrefix(lower, "file:") || isLikelySQLitePath(lower):
		return NewSQLiteConnector(dsn)
	default:
		return nil, errors.New("unsupported DSN: supported drivers are postgres, mysql, sqlite, duckdb")
	}
}

func isLikelySQLitePath(dsn string) bool {
	return strings.HasSuffix(dsn, ".db") || strings.HasSuffix(dsn, ".sqlite") || strings.HasSuffix(dsn, ".sqlite3") || dsn == ":memory:"
}

func isLikelyDuckDBPath(dsn string) bool {
	return strings.HasSuffix(dsn, ".duckdb") || strings.HasSuffix(dsn, ".ddb")
}
