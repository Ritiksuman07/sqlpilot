package db

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

type PostgresConnector struct {
	dsn string
	db  *sql.DB
}

func NewPostgresConnector(dsn string) (*PostgresConnector, error) {
	connector := &PostgresConnector{dsn: dsn}
	if err := connector.Connect(context.Background(), dsn); err != nil {
		return nil, err
	}
	return connector, nil
}

func (p *PostgresConnector) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	p.db = db
	return nil
}

func (p *PostgresConnector) ListTables(ctx context.Context) ([]Table, error) {
	rows, err := p.db.QueryContext(ctx, `
		SELECT n.nspname AS schema_name,
			   c.relname AS table_name,
			   CASE c.relkind WHEN 'v' THEN 'VIEW' ELSE 'BASE TABLE' END AS table_type,
			   c.reltuples::bigint AS row_estimate
		FROM pg_class c
		JOIN pg_namespace n ON n.oid = c.relnamespace
		WHERE n.nspname NOT IN ('pg_catalog', 'information_schema')
		  AND c.relkind IN ('r','v')
		ORDER BY n.nspname, c.relname`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var schema, name, tableType string
		var rowsEstimate int64
		if err := rows.Scan(&schema, &name, &tableType, &rowsEstimate); err != nil {
			return nil, err
		}
		kind := "table"
		if tableType == "VIEW" {
			kind = "view"
		}
		tables = append(tables, Table{Schema: schema, Name: name, Type: kind, Rows: rowsEstimate})
	}
	return tables, rows.Err()
}

func (p *PostgresConnector) ListColumns(ctx context.Context, table string) ([]Column, error) {
	schema, name := splitQualified(table)
	rows, err := p.db.QueryContext(ctx, `
		SELECT column_name, data_type, is_nullable
		FROM information_schema.columns
		WHERE table_schema = $1 AND table_name = $2
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

func (p *PostgresConnector) Execute(ctx context.Context, query string) (*Result, error) {
	start := time.Now()
	rows, err := p.db.QueryContext(ctx, query)
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

func (p *PostgresConnector) Close() error {
	if p.db == nil {
		return nil
	}
	return p.db.Close()
}
