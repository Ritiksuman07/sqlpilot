//go:build !duckdb

package db

import (
	"context"
	"errors"
)

type DuckDBConnector struct{}

func NewDuckDBConnector(dsn string) (*DuckDBConnector, error) {
	return nil, errors.New("duckdb connector requires build tag: -tags duckdb")
}

func (d *DuckDBConnector) Connect(ctx context.Context, dsn string) error {
	return errors.New("duckdb connector requires build tag: -tags duckdb")
}

func (d *DuckDBConnector) ListTables(ctx context.Context) ([]Table, error) {
	return nil, errors.New("duckdb connector requires build tag: -tags duckdb")
}

func (d *DuckDBConnector) ListColumns(ctx context.Context, table string) ([]Column, error) {
	return nil, errors.New("duckdb connector requires build tag: -tags duckdb")
}

func (d *DuckDBConnector) Execute(ctx context.Context, query string) (*Result, error) {
	return nil, errors.New("duckdb connector requires build tag: -tags duckdb")
}

func (d *DuckDBConnector) Close() error {
	return nil
}
