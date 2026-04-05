# SQLPilot — terminal SQL explorer

A beautiful, keyboard‑driven terminal SQL explorer for Postgres, MySQL, SQLite, and DuckDB. One binary. No browser required.

[![Version](https://img.shields.io/github/v/tag/Ritiksuman07/sqlpilot)](https://github.com/Ritiksuman07/sqlpilot/tags)
[![Stars](https://img.shields.io/github/stars/Ritiksuman07/sqlpilot)](https://github.com/Ritiksuman07/sqlpilot/stargazers)
[![Build](https://github.com/Ritiksuman07/sqlpilot/actions/workflows/ci.yml/badge.svg?branch=master)](https://github.com/Ritiksuman07/sqlpilot/actions/workflows/ci.yml)
![Go](https://img.shields.io/badge/Go-1.23%2B-00ADD8-logo=go&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-10B981)
![Platform](https://img.shields.io/badge/Platforms-macOS%20%7C%20Linux%20%7C%20WSL-111827)

Status: v0.5.0 (background autocomplete preload, profile picker, improved history filtering).

## Screenshots & Demo
<img src="assets/sqlpilot-screenshot.png" alt="SQLPilot TUI screenshot showing schema, editor, and results panels" title="SQLPilot TUI screenshot" />

<img src="assets/sqlpilot-demo.gif" alt="SQLPilot demo GIF showing schema navigation, autocomplete, and exports" title="SQLPilot demo GIF" />

## Table of Contents
- Why SQLPilot
- Features
- Install
- Quick Start
- Keybindings
- Profiles
- Export
- Roadmap
- Keywords
- Tags
- Contributing
- License

## Why SQLPilot
- Terminal‑native schema browsing without leaving your workflow
- Keyboard‑first querying with instant results
- Zero‑config feel with profiles + keychain storage
- Single Go binary, no GUI dependencies

## Features
- Three‑panel TUI: schema tree, query editor, results pager
- Multi‑DB: Postgres, MySQL, SQLite, DuckDB
- Autocomplete from live schema (columns preloaded in background)
- SQL formatter (`Ctrl+L`)
- Query history with fuzzy filter and highlighted matches
- CSV/JSON export from results
- Profile wizard + OS keychain password storage
- Fuzzy profile picker when multiple profiles exist

## Install

### Go install
```bash
go install github.com/ritiksuman07/sqlpilot@latest
```

### From source
```bash
git clone https://github.com/Ritiksuman07/sqlpilot.git
cd sqlpilot

go build ./...
./sqlpilot
```

### DuckDB build tag
DuckDB requires the `duckdb` build tag (and CGO enabled).
```bash
go build -tags duckdb ./cmd/sqlpilot
```

## Quick Start

### Postgres
```bash
sqlpilot --dsn "postgres://user:pass@localhost:5432/dbname"
```

### SQLite
```bash
sqlpilot --dsn "/path/to/app.db"
```

### MySQL
```bash
sqlpilot --dsn "mysql://user:pass@localhost:3306/dbname"
```

### DuckDB
```bash
go run -tags duckdb ./cmd/sqlpilot --dsn "/path/to/analytics.duckdb"
```

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

## Profiles
If no DSN is provided, SQLPilot launches a connection wizard and stores passwords in the OS keychain.
Profiles live at `~/.config/sqlpilot/connections.yaml` and can be selected with `--profile`.
If multiple profiles exist, a fuzzy‑search picker appears on launch.

## Export
`Ctrl+E` writes CSV and `Ctrl+J` writes JSON to a timestamped file in the current working directory.

## Roadmap
- v0.6: smarter autocomplete ranking + export path prompt
- v0.7: connection manager UI + schema search
- v1.0: polished UX, docs, and release artifacts

## Keywords
terminal SQL explorer, TUI database client, CLI SQL client, terminal database browser, database explorer, Postgres TUI, MySQL TUI, SQLite TUI, DuckDB CLI, SQL query tool, SQL client, Go TUI, Bubble Tea, command line database tool, schema viewer, terminal data viewer

## Tags
#sql #database #postgres #mysql #sqlite #duckdb #terminal #tui #cli #golang #devtools #opensource #datatools #backend #dataengineering

## Contributing
Issues and PRs are welcome. Please include steps to reproduce and a short rationale for changes.

- Contribution guide: `CONTRIBUTING.md`
- Security policy: `SECURITY.md`
- Code of conduct: `CODE_OF_CONDUCT.md`
- Changelog: `CHANGELOG.md`

## License
MIT
