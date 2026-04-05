package db

import (
	"context"
	"database/sql"
	"fmt"
	"net/url"
	"strings"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

type MySQLConnector struct {
	dsn string
	db  *sql.DB
}

func NewMySQLConnector(dsn string) (*MySQLConnector, error) {
	connector := &MySQLConnector{dsn: normalizeMySQLDSN(dsn)}
	if err := connector.Connect(context.Background(), connector.dsn); err != nil {
		return nil, err
	}
	return connector, nil
}

func (m *MySQLConnector) Connect(ctx context.Context, dsn string) error {
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return err
	}
	db.SetConnMaxLifetime(5 * time.Minute)
	db.SetMaxOpenConns(4)
	db.SetMaxIdleConns(4)
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	m.db = db
	return nil
}

func (m *MySQLConnector) ListTables(ctx context.Context) ([]Table, error) {
	rows, err := m.db.QueryContext(ctx, `
		SELECT table_schema, table_name, table_type, table_rows
		FROM information_schema.tables
		WHERE table_schema NOT IN ('information_schema', 'mysql', 'performance_schema', 'sys')
		ORDER BY table_schema, table_name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tables []Table
	for rows.Next() {
		var schema, name, tableType string
		var rowsEstimate sql.NullInt64
		if err := rows.Scan(&schema, &name, &tableType, &rowsEstimate); err != nil {
			return nil, err
		}
		kind := "table"
		if strings.Contains(strings.ToLower(tableType), "view") {
			kind = "view"
		}
		rowCount := int64(0)
		if rowsEstimate.Valid {
			rowCount = rowsEstimate.Int64
		}
		tables = append(tables, Table{Schema: schema, Name: name, Type: kind, Rows: rowCount})
	}
	return tables, rows.Err()
}

func (m *MySQLConnector) ListColumns(ctx context.Context, table string) ([]Column, error) {
	schema, name := splitQualified(table)
	rows, err := m.db.QueryContext(ctx, `
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

func (m *MySQLConnector) Execute(ctx context.Context, query string) (*Result, error) {
	start := time.Now()
	rows, err := m.db.QueryContext(ctx, query)
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

func (m *MySQLConnector) Close() error {
	if m.db == nil {
		return nil
	}
	return m.db.Close()
}

func normalizeMySQLDSN(dsn string) string {
	if strings.HasPrefix(strings.ToLower(dsn), "mysql://") {
		if parsed, err := url.Parse(dsn); err == nil {
			user := parsed.User.Username()
			pass, _ := parsed.User.Password()
			host := parsed.Host
			dbname := strings.TrimPrefix(parsed.Path, "/")
			query := parsed.RawQuery
			auth := user
			if pass != "" {
				auth = fmt.Sprintf("%s:%s", user, pass)
			}
			if host == "" {
				host = "127.0.0.1:3306"
			}
			dsn = fmt.Sprintf("%s@tcp(%s)/%s", auth, host, dbname)
			if query != "" {
				dsn = dsn + "?" + query
			}
		}
	}
	if !strings.Contains(dsn, "parseTime=") {
		sep := "?"
		if strings.Contains(dsn, "?") {
			sep = "&"
		}
		dsn = dsn + sep + "parseTime=true"
	}
	return dsn
}
