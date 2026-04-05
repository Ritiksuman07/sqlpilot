package msg

import "github.com/ritiksuman07/sqlpilot/internal/db"
import "github.com/ritiksuman07/sqlpilot/internal/history"

type Err struct {
	Err error
}

type Tables struct {
	Tables []db.Table
}

type Results struct {
	Result *db.Result
}

type SelectTable struct {
	Schema string
	Name   string
}

type ExecuteQuery struct {
	Query string
}

type OpenHistory struct{}

type HistoryItems struct {
	Entries []history.Entry
}

type ColumnsLoaded struct {
	Schema  string
	Table   string
	Columns []db.Column
}

type RequestColumns struct {
	Schema string
	Table  string
}

type AutocompleteUpdate struct {
	Words []string
}

type ExportRequest struct {
	Format string
}

type ExportDone struct {
	Path   string
	Format string
	Err    error
}
