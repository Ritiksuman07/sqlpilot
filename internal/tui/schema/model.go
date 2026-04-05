package schema

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ritiksuman07/sqlpilot/internal/db"
	"github.com/ritiksuman07/sqlpilot/internal/tui/msg"
)

type Model struct {
	list     list.Model
	focus    bool
	width    int
	height   int
	tables   []db.Table
	columns  map[string][]db.Column
	expanded map[string]bool
}

type item struct {
	schema string
	name   string
	kind   string
	rows   int64
	indent int
}

func (i item) Title() string {
	if i.schema == "" || i.schema == "main" {
		return indent(i.indent) + i.name
	}
	return indent(i.indent) + fmt.Sprintf("%s.%s", i.schema, i.name)
}
func (i item) Description() string { return i.kind }
func (i item) FilterValue() string { return i.name }

func New() Model {
	items := []list.Item{}
	l := list.New(items, list.NewDefaultDelegate(), 24, 10)
	l.Title = "Schema"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)
	return Model{
		list:     l,
		columns: map[string][]db.Column{},
		expanded: map[string]bool{},
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	if loaded, ok := message.(msg.ColumnsLoaded); ok {
		return m.WithColumns(loaded.Schema, loaded.Table, loaded.Columns), nil
	}
	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "enter":
			if selected, ok := m.list.SelectedItem().(item); ok {
				return m, func() tea.Msg {
					return msg.SelectTable{Schema: selected.schema, Name: selected.name}
				}
			}
		case "right", "l", " ":
			if selected, ok := m.list.SelectedItem().(item); ok && selected.kind == "table" {
				key := tableKey(selected.schema, selected.name)
				m.expanded[key] = !m.expanded[key]
				if m.expanded[key] && len(m.columns[key]) == 0 {
					return m, func() tea.Msg {
						return msg.RequestColumns{Schema: selected.schema, Table: selected.name}
					}
				}
				m.rebuild()
				return m, nil
			}
		case "left", "h":
			if selected, ok := m.list.SelectedItem().(item); ok && selected.kind == "table" {
				key := tableKey(selected.schema, selected.name)
				if m.expanded[key] {
					m.expanded[key] = false
					m.rebuild()
					return m, nil
				}
			}
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(message)
	return m, cmd
}

func (m Model) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor())
	return style.Width(m.width).Height(m.height).Render(m.list.View())
}

func (m Model) WithFocus(focus bool) Model {
	m.focus = focus
	return m
}

func (m Model) WithSize(width, height int) Model {
	m.width = width
	m.height = height
	m.list.SetSize(width-2, height-2)
	return m
}

func (m Model) WithTables(tables []db.Table) Model {
	m.tables = tables
	m.rebuild()
	return m
}

func (m Model) WithColumns(schemaName, tableName string, cols []db.Column) Model {
	key := tableKey(schemaName, tableName)
	m.columns[key] = cols
	m.expanded[key] = true
	m.rebuild()
	return m
}

func (m *Model) rebuild() {
	items := make([]list.Item, 0, len(m.tables))
	for _, table := range m.tables {
		kind := table.Type
		if kind == "" {
			kind = "table"
		}
		itemRow := item{
			schema: table.Schema,
			name:   table.Name,
			kind:   tableLabel(kind, table.Rows),
			rows:   table.Rows,
		}
		items = append(items, itemRow)

		key := tableKey(table.Schema, table.Name)
		if m.expanded[key] {
			for _, col := range m.columns[key] {
				colLabel := fmt.Sprintf("%s %s", col.Name, col.DataType)
				if col.Nullable {
					colLabel = fmt.Sprintf("%s nullable", colLabel)
				}
				items = append(items, item{
					schema: table.Schema,
					name:   colLabel,
					kind:   "column",
					indent: 2,
				})
			}
		}
	}
	m.list.SetItems(items)
}

func tableKey(schemaName, tableName string) string {
	return fmt.Sprintf("%s.%s", schemaName, tableName)
}

func indent(level int) string {
	return strings.Repeat(" ", level)
}

func tableLabel(kind string, rows int64) string {
	if rows <= 0 {
		return kind
	}
	return fmt.Sprintf("%s · ~%d rows", kind, rows)
}

func (m Model) borderColor() lipgloss.Color {
	if m.focus {
		return lipgloss.Color("39")
	}
	return lipgloss.Color("238")
}
