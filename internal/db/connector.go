package db

import "context"

type Connector interface {
	Connect(ctx context.Context, dsn string) error
	ListTables(ctx context.Context) ([]Table, error)
	ListColumns(ctx context.Context, table string) ([]Column, error)
	Execute(ctx context.Context, query string) (*Result, error)
	Close() error
}

type Table struct {
	Schema string
	Name   string
	Type   string
	Rows   int64
}

type Column struct {
	Name     string
	DataType string
	Nullable bool
}

type Result struct {
	Columns []string
	Rows    [][]string
	Elapsed string
}
