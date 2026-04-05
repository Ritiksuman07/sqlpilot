# SQLPilot

A beautiful, keyboard-driven terminal SQL explorer for Postgres, MySQL, SQLite, and DuckDB. One binary. No browser required.

Status: v0.2.0 (core loop + history + schema expansion). MySQL/DuckDB are planned in v0.3.

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

## Keybindings
- `Tab` / `Shift+Tab`: cycle focus between panels
- `F5` or `Ctrl+Enter`: run query
- `Ctrl+H`: open query history picker
- `Enter` on table: fill editor with `SELECT * FROM table LIMIT 100`
- `Right` / `Space`: expand table columns
- `Left`: collapse table columns
- `q` or `Ctrl+Q`: quit

## Layout
- `cmd/sqlpilot` CLI entry
- `internal/tui` TUI app + panels
- `internal/db` connector interface
- `internal/config` connection profile store
- `internal/history` query history store
- `internal/export` CSV/JSON export

## Notes
This is a Go + Charmbracelet (Bubble Tea) project. v0.2 adds history, schema expansion, and richer status feedback.

PRD snapshot: `docs/PRD.txt`
