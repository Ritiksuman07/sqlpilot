package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteConnector struct {
	dsn string
	db  *sql.DB
}

func NewSQLiteConnector(dsn string) (*SQLiteConnector, error) {
	connector := &SQLiteConnector{dsn: normalizeSQLiteDSN(dsn)}
	if err := connector.Connect(context.Background(), connector.dsn); err != nil {
		return nil, err
	}
	return connector, nil
}

func (s *SQLiteConnector) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	s.db = db
	return nil
}

func (s *SQLiteConnector) ListTables(ctx context.Context) ([]Table, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT name, type
		FROM sqlite_master
		WHERE type IN ('table','view')
		  AND name NOT LIKE 'sqlite_%'
		ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var name, typ string
		if err := rows.Scan(&name, &typ); err != nil {
			return nil, err
		}
		rowCount := int64(0)
		if typ == "table" {
			if count, err := s.tableCount(ctx, name); err == nil {
				rowCount = count
			}
		}
		tables = append(tables, Table{Schema: "main", Name: name, Type: typ, Rows: rowCount})
	}
	return tables, rows.Err()
}

func (s *SQLiteConnector) ListColumns(ctx context.Context, table string) ([]Column, error) {
	rows, err := s.db.QueryContext(ctx, fmt.Sprintf("PRAGMA table_info(%s)", table))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var cid int
		var name, dataType string
		var notNull int
		var dflt sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &dataType, &notNull, &dflt, &pk); err != nil {
			return nil, err
		}
		cols = append(cols, Column{Name: name, DataType: dataType, Nullable: notNull == 0})
	}
	return cols, rows.Err()
}

func (s *SQLiteConnector) Execute(ctx context.Context, query string) (*Result, error) {
	start := time.Now()
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	cols, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	values := make([]any, len(cols))
	ptrs := make([]any, len(cols))
	resultRows := make([][]string, 0, 100)
	for i := range values {
		ptrs[i] = &values[i]
	}

	for rows.Next() {
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		row := make([]string, len(cols))
		for i, v := range values {
			row[i] = formatValue(v)
		}
		resultRows = append(resultRows, row)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return &Result{
		Columns: cols,
		Rows:    resultRows,
		Elapsed: fmt.Sprintf("%s", time.Since(start).Truncate(time.Millisecond)),
	}, nil
}

func (s *SQLiteConnector) Close() error {
	if s.db == nil {
		return nil
	}
	return s.db.Close()
}

func (s *SQLiteConnector) tableCount(ctx context.Context, table string) (int64, error) {
	row := s.db.QueryRowContext(ctx, fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
	var count int64
	if err := row.Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

func normalizeSQLiteDSN(dsn string) string {
	trimmed := strings.TrimPrefix(dsn, "sqlite://")
	trimmed = strings.TrimPrefix(trimmed, "sqlite:")
	if trimmed == "" {
		return dsn
	}
	return trimmed
}
