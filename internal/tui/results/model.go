package results

import (
	"unicode/utf8"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ritiksuman07/sqlpilot/internal/db"
)

type Model struct {
	table  table.Model
	focus  bool
	width  int
	height int
}

func New() Model {
	columns := []table.Column{
		{Title: "id", Width: 4},
		{Title: "email", Width: 24},
		{Title: "created_at", Width: 20},
	}
	rows := []table.Row{
		{"1", "alex@example.com", "2026-04-05 10:21"},
		{"2", "maya@example.com", "2026-04-05 10:22"},
		{"3", "jordan@example.com", "2026-04-05 10:23"},
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
	)
	styles := table.DefaultStyles()
	styles.Header = styles.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("240")).
		BorderBottom(true).
		Bold(true)
	styles.Selected = styles.Selected.
		Foreground(lipgloss.Color("229")).
		Background(lipgloss.Color("57")).
		Bold(false)
	t.SetStyles(styles)

	return Model{table: t, focus: true}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (Model, tea.Cmd) {
	var cmd tea.Cmd
	m.table, cmd = m.table.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor())
	return style.Width(m.width).Height(m.height).Render(m.table.View())
}

func (m Model) WithFocus(focus bool) Model {
	m.focus = focus
	return m
}

func (m Model) WithSize(width, height int) Model {
	m.width = width
	m.height = height
	m.table.SetWidth(width - 2)
	m.table.SetHeight(height - 2)
	return m
}

func (m Model) WithResult(result db.Result) Model {
	columns := make([]table.Column, 0, len(result.Columns))
	for _, col := range result.Columns {
		columns = append(columns, table.Column{Title: col, Width: 12})
	}

	rows := make([]table.Row, 0, len(result.Rows))
	for _, row := range result.Rows {
		rows = append(rows, table.Row(row))
	}

	widths := computeWidths(result.Columns, result.Rows)
	for i := range columns {
		if i < len(widths) {
			columns[i].Width = widths[i]
		}
	}

	m.table.SetColumns(columns)
	m.table.SetRows(rows)
	return m
}

func computeWidths(columns []string, rows [][]string) []int {
	widths := make([]int, len(columns))
	for i, col := range columns {
		widths[i] = clamp(utf8.RuneCountInString(col), 6, 40)
	}
	for _, row := range rows {
		for i, value := range row {
			if i >= len(widths) {
				continue
			}
			widths[i] = clamp(max(widths[i], utf8.RuneCountInString(value)), 6, 40)
		}
	}
	return widths
}

func clamp(value, minValue, maxValue int) int {
	if value < minValue {
		return minValue
	}
	if value > maxValue {
		return maxValue
	}
	return value
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) borderColor() lipgloss.Color {
	if m.focus {
		return lipgloss.Color("39")
	}
	return lipgloss.Color("238")
}
