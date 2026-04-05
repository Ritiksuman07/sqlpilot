# SQLPilot

A beautiful, keyboard-driven terminal SQL explorer for Postgres, MySQL, SQLite, and DuckDB. One binary. No browser required.

Status: v0.5.0 (background autocomplete preload, profile picker, improved history filtering).

## Why SQLPilot
- Zero-config, terminal-native database exploration
- Fast schema browsing with lazy expansion
- Query with instant, scrollable results
- Keyboard-first workflow (no mouse, no browser)

## Goals (v1.0)
- Single binary installable via brew/apt/go install
- Three-panel layout (schema tree / query editor / results pager)
- Zero-config connection workflow
- Fast start and smooth terminal UX

## Roadmap (from PRD)
- v0.1: Core loop (SQLite + Postgres, schema tree, editor, results, keyboard nav)
- v0.2: UX polish (history, error display, richer status bar, schema expansion)
- v0.3: Multi-DB (MySQL + DuckDB, exports, connection profiles)
- v0.4: Autocomplete + formatting

## Run (dev)
```bash
cd sqlpilot

go run ./cmd/sqlpilot --dsn "postgres://user:pass@localhost:5432/dbname"
```

SQLite example:
```bash
go run ./cmd/sqlpilot --dsn "/path/to/app.db"
```

MySQL example:
```bash
go run ./cmd/sqlpilot --dsn "mysql://user:pass@localhost:3306/dbname"
```

DuckDB example:
```bash
go run -tags duckdb ./cmd/sqlpilot --dsn "/path/to/analytics.duckdb"
```

Note: DuckDB requires building with the `duckdb` build tag (and CGO enabled).

## Keybindings
- `Tab` / `Shift+Tab`: cycle focus between panels
- `F5` or `Ctrl+Enter`: run query
- `Ctrl+Space`: autocomplete from schema
- `Ctrl+L`: format SQL
- `Ctrl+H`: open query history picker
- `?` or `F1`: help overlay
- `Enter` on table: fill editor with `SELECT * FROM table LIMIT 100`
- `Right` / `Space`: expand table columns
- `Left`: collapse table columns
- `Ctrl+E`: export CSV
- `Ctrl+J`: export JSON
- `q` or `Ctrl+Q`: quit

## History
`Ctrl+H` opens the history picker with a highlighted preview of the selected query.

## Profiles
If no DSN is provided, SQLPilot prompts for a connection profile and stores passwords in the OS keychain.
Profiles live at `~/.config/sqlpilot/connections.yaml` and can be selected with `--profile`.
If multiple profiles exist, a fuzzy-search picker appears on launch.

## Export
`Ctrl+E` writes CSV and `Ctrl+J` writes JSON to a timestamped file in the current working directory.

## Layout
- `cmd/sqlpilot` CLI entry
- `internal/tui` TUI app + panels
- `internal/db` connector interface
- `internal/config` connection profile store
- `internal/history` query history store
- `internal/export` CSV/JSON export

## Notes
This is a Go + Charmbracelet (Bubble Tea) project. v0.3 adds MySQL/DuckDB plus history preview and help overlay.

PRD snapshot: `docs/PRD.txt`
