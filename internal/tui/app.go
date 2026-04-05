package tui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ritiksuman07/sqlpilot/internal/config"
	"github.com/ritiksuman07/sqlpilot/internal/db"
	"github.com/ritiksuman07/sqlpilot/internal/export"
	"github.com/ritiksuman07/sqlpilot/internal/history"
	"github.com/ritiksuman07/sqlpilot/internal/tui/connect"
	"github.com/ritiksuman07/sqlpilot/internal/tui/editor"
	tmsg "github.com/ritiksuman07/sqlpilot/internal/tui/msg"
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

	connector         db.Connector
	history           *history.Store
	status            string
	lastInfo          string
	showHelp          bool
	lastResult        *db.Result
	words             map[string]struct{}
	showProfilePicker bool
	profileList       list.Model
	profiles          []config.Profile

	schema  schema.Model
	editor  editor.Model
	results results.Model
}

func New(options Options) Model {
	profileList := list.New([]list.Item{}, list.NewDefaultDelegate(), 40, 10)
	profileList.Title = "Select Profile"
	profileList.SetShowStatusBar(false)
	profileList.SetFilteringEnabled(true)
	profileList.SetShowHelp(false)
	return Model{
		options:     options,
		focus:       FocusEditor,
		status:      "disconnected",
		schema:      schema.New(),
		editor:      editor.New(),
		results:     results.New(),
		words:       map[string]struct{}{},
		profileList: profileList,
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
	if m.showProfilePicker {
		if key, ok := msg.(tea.KeyMsg); ok {
			switch key.String() {
			case "esc":
				m.showProfilePicker = false
				return m, nil
			case "enter":
				if selected, ok := m.profileList.SelectedItem().(profileItem); ok {
					m.showProfilePicker = false
					return m, connectProfileCmd(selected.profile)
				}
			}
		}
		var cmd tea.Cmd
		m.profileList, cmd = m.profileList.Update(msg)
		return m, cmd
	}

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.schema = m.schema.WithSize(m.schemaWidth(), m.panelHeight())
		m.editor = m.editor.WithSize(m.editorWidth(), m.panelHeight())
		m.results = m.results.WithSize(m.resultsWidth(), m.panelHeight())
		m.profileList.SetSize(min(72, m.width-6), min(16, m.height-6))
		return m, nil
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "ctrl+q", "q":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
			return m, tea.Quit
		case "?", "f1":
			m.showHelp = !m.showHelp
			return m, nil
		case "esc":
			if m.showHelp {
				m.showHelp = false
				return m, nil
			}
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
		m.addWordsFromTables(msg.Tables)
		return m, tea.Batch(
			m.pushAutocomplete(),
			preloadColumnsCmd(m.connector, msg.Tables),
		)
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
			m.lastResult = msg.Result
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
	case tmsg.ColumnsLoaded:
		m.schema = m.schema.WithColumns(msg.Schema, msg.Table, msg.Columns)
		m.addWordsFromColumns(msg.Schema, msg.Table, msg.Columns)
		return m, m.pushAutocomplete()
	case tmsg.PreloadWords:
		for _, word := range msg.Words {
			m.words[word] = struct{}{}
		}
		return m, m.pushAutocomplete()
	case tmsg.ExportRequest:
		if m.lastResult == nil {
			m.status = "error"
			m.lastInfo = "no results to export"
			return m, nil
		}
		return m, exportCmd(msg.Format, m.lastResult)
	case tmsg.ExportDone:
		if msg.Err != nil {
			m.status = "error"
			m.lastInfo = msg.Err.Error()
			return m, nil
		}
		m.status = "connected"
		m.lastInfo = fmt.Sprintf("exported %s to %s", msg.Format, msg.Path)
		return m, nil
	case dbConnectorMsg:
		m.connector = msg.connector
		m.status = "connected"
		m.lastInfo = "connected"
		return m, loadTablesCmd(m.connector)
	case historyStoreMsg:
		m.history = msg.store
		return m, nil
	case tmsg.ProfileList:
		m.profiles = msg.Profiles
		items := make([]list.Item, 0, len(msg.Profiles))
		for _, profile := range msg.Profiles {
			items = append(items, profileItem{profile: profile})
		}
		m.profileList.SetItems(items)
		m.showProfilePicker = true
		m.status = "select profile"
		m.lastInfo = "choose a connection profile"
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
	base := lipgloss.JoinVertical(lipgloss.Left, row, status)
	if m.showHelp {
		return overlayHelp(base, m.width, m.height)
	}
	if m.showProfilePicker {
		return overlayProfilePicker(base, m.profileList, m.width, m.height)
	}
	return base
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
		if strings.TrimSpace(options.DSN) != "" {
			conn, err := db.Open(options.DSN)
			if err != nil {
				return tmsg.Err{Err: err}
			}
			return dbConnectorMsg{connector: conn}
		}

		cfg, err := config.Load()
		if err != nil {
			return tmsg.Err{Err: err}
		}

		profileName := strings.TrimSpace(options.Profile)
		if profileName == "" && len(cfg.Profiles) == 0 {
			profile, password, err := connect.RunWizard()
			if err != nil {
				return tmsg.Err{Err: err}
			}
			dsn := config.BuildDSN(profile, password)
			conn, err := db.Open(dsn)
			if err != nil {
				return tmsg.Err{Err: err}
			}
			cfg.Profiles = append(cfg.Profiles, profile)
			if err := config.Save(cfg); err != nil {
				return tmsg.Err{Err: err}
			}
			if err := config.SavePassword(profile.Name, password); err != nil {
				return tmsg.Err{Err: err}
			}
			return dbConnectorMsg{connector: conn}
		}

		if profileName == "" && len(cfg.Profiles) > 1 {
			return tmsg.ProfileList{Profiles: cfg.Profiles}
		}

		if profileName == "" && len(cfg.Profiles) == 1 {
			profileName = cfg.Profiles[0].Name
		}

		profile, ok := cfg.FindProfile(profileName)
		if !ok {
			return tmsg.Err{Err: fmt.Errorf("profile not found: %s", profileName)}
		}
		password, _ := config.LoadPassword(profile.Name)
		dsn := config.BuildDSN(profile, password)
		if dsn == "" {
			return tmsg.Err{Err: fmt.Errorf("invalid profile: %s", profileName)}
		}
		conn, err := db.Open(dsn)
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

func preloadColumnsCmd(connector db.Connector, tables []db.Table) tea.Cmd {
	return func() tea.Msg {
		if connector == nil {
			return nil
		}
		words := map[string]struct{}{}
		for _, table := range tables {
			cols, err := connector.ListColumns(context.Background(), qualifiedTable(table.Schema, table.Name))
			if err != nil {
				continue
			}
			for _, col := range cols {
				words[col.Name] = struct{}{}
				words[fmt.Sprintf("%s.%s", table.Name, col.Name)] = struct{}{}
				if table.Schema != "" {
					words[fmt.Sprintf("%s.%s.%s", table.Schema, table.Name, col.Name)] = struct{}{}
				}
			}
		}
		if len(words) == 0 {
			return nil
		}
		out := make([]string, 0, len(words))
		for word := range words {
			out = append(out, word)
		}
		sort.Strings(out)
		return tmsg.PreloadWords{Words: out}
	}
}

func exportCmd(format string, result *db.Result) tea.Cmd {
	return func() tea.Msg {
		if result == nil {
			return tmsg.ExportDone{Format: format, Err: fmt.Errorf("no results to export")}
		}
		ext := "csv"
		if format == "json" {
			ext = "json"
		}
		filename := fmt.Sprintf("sqlpilot_export_%s.%s", time.Now().Format("20060102_150405"), ext)
		path := filepath.Join(".", filename)
		file, err := os.Create(path)
		if err != nil {
			return tmsg.ExportDone{Format: format, Err: err}
		}
		defer file.Close()
		if format == "json" {
			if err := export.WriteJSON(file, result.Columns, result.Rows); err != nil {
				return tmsg.ExportDone{Format: format, Err: err}
			}
		} else {
			if err := export.WriteCSV(file, result.Columns, result.Rows); err != nil {
				return tmsg.ExportDone{Format: format, Err: err}
			}
		}
		return tmsg.ExportDone{Format: format, Path: path}
	}
}

func (m *Model) addWordsFromTables(tables []db.Table) {
	for _, table := range tables {
		m.words[table.Name] = struct{}{}
		if table.Schema != "" {
			m.words[fmt.Sprintf("%s.%s", table.Schema, table.Name)] = struct{}{}
		}
	}
}

func (m *Model) addWordsFromColumns(schemaName, tableName string, cols []db.Column) {
	for _, col := range cols {
		m.words[col.Name] = struct{}{}
		if tableName != "" {
			m.words[fmt.Sprintf("%s.%s", tableName, col.Name)] = struct{}{}
		}
		if schemaName != "" && tableName != "" {
			m.words[fmt.Sprintf("%s.%s.%s", schemaName, tableName, col.Name)] = struct{}{}
		}
	}
}

func (m *Model) pushAutocomplete() tea.Cmd {
	if len(m.words) == 0 {
		return nil
	}
	words := make([]string, 0, len(m.words))
	for word := range m.words {
		words = append(words, word)
	}
	sort.Strings(words)
	return func() tea.Msg {
		return tmsg.AutocompleteUpdate{Words: words}
	}
}

func connectProfileCmd(profile config.Profile) tea.Cmd {
	return func() tea.Msg {
		password, _ := config.LoadPassword(profile.Name)
		dsn := config.BuildDSN(profile, password)
		if dsn == "" {
			return tmsg.Err{Err: fmt.Errorf("invalid profile: %s", profile.Name)}
		}
		conn, err := db.Open(dsn)
		if err != nil {
			return tmsg.Err{Err: err}
		}
		return dbConnectorMsg{connector: conn}
	}
}

type profileItem struct {
	profile config.Profile
}

func (p profileItem) Title() string {
	return p.profile.Name
}

func (p profileItem) Description() string {
	if p.profile.Driver == "sqlite" || p.profile.Driver == "duckdb" {
		return fmt.Sprintf("%s · %s", p.profile.Driver, p.profile.Path)
	}
	return fmt.Sprintf("%s · %s:%d/%s", p.profile.Driver, p.profile.Host, p.profile.Port, p.profile.Database)
}

func (p profileItem) FilterValue() string {
	return p.profile.Name + " " + p.profile.Host + " " + p.profile.Database + " " + p.profile.Path
}

func overlayProfilePicker(base string, list list.Model, width, height int) string {
	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(min(72, width-4))
	content := box.Render(list.View())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, content)
}

func qualifiedTable(schemaName, tableName string) string {
	if schemaName == "" || schemaName == "main" {
		return tableName
	}
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}

func overlayHelp(base string, width, height int) string {
	help := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("39")).
		Padding(1, 2).
		Width(min(72, width-4)).
		Render(helpText())
	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, help)
}

func helpText() string {
	return strings.Join([]string{
		"SQLPilot Help",
		"",
		"Global",
		"  Tab / Shift+Tab    Cycle focus",
		"  ? or F1            Toggle help",
		"  q / Ctrl+Q         Quit",
		"",
		"Editor",
		"  F5 / Ctrl+Enter    Run query",
		"  Ctrl+Space         Autocomplete",
		"  Ctrl+L             Format SQL",
		"  Ctrl+H             Open history",
		"",
		"Schema",
		"  Enter              SELECT * FROM table",
		"  Right / Space      Expand columns",
		"  Left               Collapse columns",
		"",
		"Results",
		"  Ctrl+E             Export CSV",
		"  Ctrl+J             Export JSON",
	}, "\n")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
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
