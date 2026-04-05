package tui

import (
	"context"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ritiksuman07/sqlpilot/internal/db"
	"github.com/ritiksuman07/sqlpilot/internal/history"
	tmsg "github.com/ritiksuman07/sqlpilot/internal/tui/msg"
	"github.com/ritiksuman07/sqlpilot/internal/tui/editor"
	"github.com/ritiksuman07/sqlpilot/internal/tui/results"
	"github.com/ritiksuman07/sqlpilot/internal/tui/schema"
)

type Focus int

const (
	FocusSchema Focus = iota
	FocusEditor
	FocusResults
)

type Options struct {
	DSN     string
	Profile string
	Version string
}

type Model struct {
	options Options
	focus   Focus
	width   int
	height  int

	connector db.Connector
	history   *history.Store
	status    string
	lastInfo  string

	schema  schema.Model
	editor  editor.Model
	results results.Model
}

func New(options Options) Model {
	return Model{
		options: options,
		focus:   FocusEditor,
		status:  "disconnected",
		schema:  schema.New(),
		editor:  editor.New(),
		results: results.New(),
	}
}

func Run(options Options) error {
	p := tea.NewProgram(New(options), tea.WithAltScreen())
	_, err := p.Run()
	return err
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.schema.Init(),
		m.editor.Init(),
		m.results.Init(),
		connectCmd(m.options),
		openHistoryCmd(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.schema = m.schema.WithSize(m.schemaWidth(), m.panelHeight())
		m.editor = m.editor.WithSize(m.editorWidth(), m.panelHeight())
		m.results = m.results.WithSize(m.resultsWidth(), m.panelHeight())
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q", "q":
			return m, tea.Quit
		case "tab":
			m.focus = (m.focus + 1) % 3
			m = m.applyFocus()
			return m, nil
		case "shift+tab":
			m.focus = (m.focus + 2) % 3
			m = m.applyFocus()
			return m, nil
		}
	case tmsg.Err:
		m.status = "error"
		m.lastInfo = msg.Err.Error()
		m.editor = m.editor.WithError(msg.Err.Error())
		return m, nil
	case tmsg.Tables:
		m.schema = m.schema.WithTables(msg.Tables)
		return m, nil
	case tmsg.SelectTable:
		query := fmt.Sprintf("SELECT * FROM %s LIMIT 100;", qualifiedTable(msg.Schema, msg.Name))
		m.editor = m.editor.WithQuery(query)
		m.focus = FocusEditor
		m = m.applyFocus()
		return m, nil
	case tmsg.ExecuteQuery:
		if m.connector == nil {
			m.status = "error"
			m.lastInfo = "no active connection"
			m.editor = m.editor.WithError("no active connection")
			return m, nil
		}
		m.editor = m.editor.WithError("")
		m.status = "running"
		m.lastInfo = "query executing..."
		return m, tea.Batch(
			executeQueryCmd(m.connector, msg.Query),
			addHistoryCmd(m.history, msg.Query),
		)
	case tmsg.Results:
		if msg.Result != nil {
			m.results = m.results.WithResult(*msg.Result)
			m.status = "connected"
			m.lastInfo = fmt.Sprintf("rows: %d  time: %s", len(msg.Result.Rows), msg.Result.Elapsed)
		}
		return m, nil
	case tmsg.OpenHistory:
		if m.history == nil {
			m.status = "error"
			m.lastInfo = "history store not available"
			m.editor = m.editor.WithError("history store not available")
			return m, nil
		}
		return m, loadHistoryCmd(m.history)
	case tmsg.RequestColumns:
		if m.connector == nil {
			return m, nil
		}
		return m, loadColumnsCmd(m.connector, msg.Schema, msg.Table)
	case dbConnectorMsg:
		m.connector = msg.connector
		m.status = "connected"
		m.lastInfo = "connected"
		return m, loadTablesCmd(m.connector)
	case historyStoreMsg:
		m.history = msg.store
		return m, nil
	}

	var cmds []tea.Cmd
	var cmd tea.Cmd

	m.schema, cmd = m.schema.Update(msg)
	cmds = append(cmds, cmd)
	m.editor, cmd = m.editor.Update(msg)
	cmds = append(cmds, cmd)
	m.results, cmd = m.results.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	left := m.schema.View()
	middle := m.editor.View()
	right := m.results.View()

	status := lipgloss.NewStyle().
		Foreground(lipgloss.Color("241")).
		Padding(0, 1).
		Render(fmt.Sprintf(
			"Focus: %s  Status: %s  DSN: %s  %s",
			m.focusLabel(),
			strings.ToUpper(m.status),
			maskDSN(m.options.DSN),
			m.lastInfo,
		))

	row := lipgloss.JoinHorizontal(lipgloss.Top, left, middle, right)
	return lipgloss.JoinVertical(lipgloss.Left, row, status)
}

func (m Model) focusLabel() string {
	switch m.focus {
	case FocusSchema:
		return "Schema"
	case FocusEditor:
		return "Editor"
	case FocusResults:
		return "Results"
	default:
		return "Unknown"
	}
}

func (m Model) applyFocus() Model {
	m.schema = m.schema.WithFocus(m.focus == FocusSchema)
	m.editor = m.editor.WithFocus(m.focus == FocusEditor)
	m.results = m.results.WithFocus(m.focus == FocusResults)
	return m
}

func (m Model) panelHeight() int {
	if m.height <= 1 {
		return m.height
	}
	return m.height - 1
}

func (m Model) schemaWidth() int {
	return max(24, m.width/4)
}

func (m Model) editorWidth() int {
	return max(40, m.width/3)
}

func (m Model) resultsWidth() int {
	return max(40, m.width-m.schemaWidth()-m.editorWidth())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

type dbConnectorMsg struct {
	connector db.Connector
}

type historyStoreMsg struct {
	store *history.Store
}

func connectCmd(options Options) tea.Cmd {
	return func() tea.Msg {
		if strings.TrimSpace(options.DSN) == "" {
			return tmsg.Err{Err: fmt.Errorf("missing DSN: run with --dsn")}
		}
		conn, err := db.Open(options.DSN)
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return dbConnectorMsg{connector: conn}
	}
}

func openHistoryCmd() tea.Cmd {
	return func() tea.Msg {
		store, err := history.Open()
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return historyStoreMsg{store: store}
	}
}

func loadTablesCmd(connector db.Connector) tea.Cmd {
	return func() tea.Msg {
		tables, err := connector.ListTables(context.Background())
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return tmsg.Tables{Tables: tables}
	}
}

func executeQueryCmd(connector db.Connector, query string) tea.Cmd {
	return func() tea.Msg {
		result, err := connector.Execute(context.Background(), query)
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return tmsg.Results{Result: result}
	}
}

func loadHistoryCmd(store *history.Store) tea.Cmd {
	return func() tea.Msg {
		entries, err := store.List(200)
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return tmsg.HistoryItems{Entries: entries}
	}
}

func addHistoryCmd(store *history.Store, query string) tea.Cmd {
	return func() tea.Msg {
		if store == nil {
			return nil
		}
		if err := store.Add(query); err != nil {
			return tmsg.Err{Err: err}
		}
		return nil
	}
}

func loadColumnsCmd(connector db.Connector, schemaName, tableName string) tea.Cmd {
	return func() tea.Msg {
		cols, err := connector.ListColumns(context.Background(), qualifiedTable(schemaName, tableName))
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return tmsg.ColumnsLoaded{Schema: schemaName, Table: tableName, Columns: cols}
	}
}

func qualifiedTable(schemaName, tableName string) string {
	if schemaName == "" || schemaName == "main" {
		return tableName
	}
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}

func maskDSN(dsn string) string {
	if dsn == "" {
		return "-"
	}
	parts := strings.SplitN(dsn, "@", 2)
	if len(parts) != 2 {
		return dsn
	}
	userinfo := parts[0]
	host := parts[1]
	if strings.Contains(userinfo, "://") {
		protoSplit := strings.SplitN(userinfo, "://", 2)
		creds := protoSplit[1]
		if strings.Contains(creds, ":") {
			credsParts := strings.SplitN(creds, ":", 2)
			return fmt.Sprintf("%s://%s:****@%s", protoSplit[0], credsParts[0], host)
		}
	}
	return dsn
}
