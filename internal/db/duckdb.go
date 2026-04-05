//go:build duckdb

package db

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	_ "github.com/marcboeker/go-duckdb"
)

type DuckDBConnector struct {
	dsn string
	db  *sql.DB
}

func NewDuckDBConnector(dsn string) (*DuckDBConnector, error) {
	connector := &DuckDBConnector{dsn: normalizeDuckDBDSN(dsn)}
	if err := connector.Connect(context.Background(), connector.dsn); err != nil {
		return nil, err
	}
	return connector, nil
}

func (d *DuckDBConnector) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("duckdb", dsn)
	if err != nil {
		return err
	}
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	d.db = db
	return nil
}

func (d *DuckDBConnector) ListTables(ctx context.Context) ([]Table, error) {
	rows, err := d.db.QueryContext(ctx, `
		SELECT table_schema, table_name, table_type
		FROM information_schema.tables
		WHERE table_schema NOT IN ('information_schema', 'pg_catalog')
		ORDER BY table_schema, table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var schema, name, tableType string
		if err := rows.Scan(&schema, &name, &tableType); err != nil {
			return nil, err
		}
		kind := "table"
		if strings.Contains(strings.ToLower(tableType), "view") {
			kind = "view"
		}
		tables = append(tables, Table{Schema: schema, Name: name, Type: kind})
	}
	return tables, rows.Err()
}

func (d *DuckDBConnector) ListColumns(ctx context.Context, table string) ([]Column, error) {
	schema, name := splitQualified(table)
	rows, err := d.db.QueryContext(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = ? AND table_name = ?
		ORDER BY ordinal_position`, schema, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cols []Column
	for rows.Next() {
		var name, dataType, nullable string
		if err := rows.Scan(&name, &dataType, &nullable); err != nil {
			return nil, err
		}
		cols = append(cols, Column{Name: name, DataType: dataType, Nullable: nullable == "YES"})
	}
	return cols, rows.Err()
}

func (d *DuckDBConnector) Execute(ctx context.Context, query string) (*Result, error) {
	start := time.Now()
	rows, err := d.db.QueryContext(ctx, query)
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

func (d *DuckDBConnector) Close() error {
	if d.db == nil {
		return nil
	}
	return d.db.Close()
}

func normalizeDuckDBDSN(dsn string) string {
	trimmed := strings.TrimPrefix(strings.ToLower(dsn), "duckdb://")
	trimmed = strings.TrimPrefix(trimmed, "duckdb:")
	if trimmed == "" {
		return dsn
	}
	return trimmed
}
