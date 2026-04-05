package editor

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/ritiksuman07/sqlpilot/internal/history"
	"github.com/ritiksuman07/sqlpilot/internal/tui/msg"
)

type Model struct {
	area        textarea.Model
	historyList list.Model
	showHistory bool
	lastError   string
	preview     string
	focus       bool
	width       int
	height      int
}

func New() Model {
	area := textarea.New()
	area.Placeholder = "SELECT * FROM table LIMIT 100;"
	area.Focus()
	area.Prompt = ""
	area.ShowLineNumbers = true
	area.SetValue("SELECT * FROM users LIMIT 100;")
	hlist := list.New([]list.Item{}, list.NewDefaultDelegate(), 20, 5)
	hlist.Title = "History"
	hlist.SetShowStatusBar(false)
	hlist.SetFilteringEnabled(true)
	hlist.SetShowHelp(false)
	return Model{area: area, historyList: hlist, focus: true}
}

func (m Model) Init() tea.Cmd {
	return textarea.Blink
}

func (m Model) Update(message tea.Msg) (Model, tea.Cmd) {
	switch typed := message.(type) {
	case msg.HistoryItems:
		items := make([]list.Item, 0, len(typed.Entries))
		for _, entry := range typed.Entries {
			items = append(items, historyItem{entry: entry})
		}
		m.historyList.SetItems(items)
		m.preview = ""
		return m, nil
	}

	if key, ok := message.(tea.KeyMsg); ok {
		switch key.String() {
		case "f5", "ctrl+enter":
			query := strings.TrimSpace(m.area.Value())
			if query == "" {
				return m, nil
			}
			return m, func() tea.Msg {
				return msg.ExecuteQuery{Query: query}
			}
		case "ctrl+h":
			m.showHistory = true
			return m, func() tea.Msg {
				return msg.OpenHistory{}
			}
		case "esc":
			if m.showHistory {
				m.showHistory = false
				return m, nil
			}
		case "enter":
			if m.showHistory {
				if selected, ok := m.historyList.SelectedItem().(historyItem); ok {
					m.area.SetValue(selected.entry.Query)
					m.showHistory = false
				}
				return m, nil
			}
		}
	}

	var cmd tea.Cmd
	if m.showHistory {
		m.historyList, cmd = m.historyList.Update(message)
		m.preview = previewForHistory(m.historyList.SelectedItem(), max(4, m.height/2-2))
		return m, cmd
	}
	m.area, cmd = m.area.Update(message)
	return m, cmd
}

func (m Model) View() string {
	style := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(m.borderColor())
	body := m.area.View()
	if m.showHistory {
		listHeight := max(6, m.height/2)
		previewHeight := max(4, m.height-2-listHeight)
		m.historyList.SetSize(m.width-2, listHeight)
		preview := renderPreview(m.preview, previewHeight, m.width-2)
		body = lipgloss.JoinVertical(lipgloss.Left, m.historyList.View(), preview)
	}
	if m.lastError != "" {
		errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("160"))
		body = lipgloss.JoinVertical(lipgloss.Left, body, errorStyle.Render(m.lastError))
	}
	return style.Width(m.width).Height(m.height).Render(body)
}

func (m Model) WithFocus(focus bool) Model {
	m.focus = focus
	if focus {
		m.area.Focus()
	} else {
		m.area.Blur()
	}
	return m
}

func (m Model) WithSize(width, height int) Model {
	m.width = width
	m.height = height
	bodyHeight := height - 2
	if m.lastError != "" {
		bodyHeight = max(1, bodyHeight-1)
	}
	m.area.SetWidth(width - 2)
	m.area.SetHeight(bodyHeight)
	m.historyList.SetSize(width-2, bodyHeight)
	return m
}

func (m Model) WithQuery(query string) Model {
	m.area.SetValue(query)
	return m
}

func (m Model) WithError(err string) Model {
	m.lastError = err
	if m.width > 0 && m.height > 0 {
		return m.WithSize(m.width, m.height)
	}
	return m
}

func (m Model) borderColor() lipgloss.Color {
	if m.focus {
		return lipgloss.Color("39")
	}
	return lipgloss.Color("238")
}

type historyItem struct {
	entry history.Entry
}

func (h historyItem) Title() string {
	return truncateLine(normalizeQuery(h.entry.Query), 80)
}

func (h historyItem) Description() string {
	if h.entry.CreatedAt.IsZero() {
		return ""
	}
	lines := 1 + strings.Count(h.entry.Query, "\n")
	return h.entry.CreatedAt.Format("2006-01-02 15:04:05") + " · " + fmt.Sprintf("%d lines", lines)
}

func (h historyItem) FilterValue() string {
	return h.entry.Query
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func normalizeQuery(query string) string {
	query = strings.ReplaceAll(query, "\n", " ")
	query = strings.ReplaceAll(query, "\t", " ")
	query = strings.TrimSpace(query)
	return strings.Join(strings.Fields(query), " ")
}

func truncateLine(value string, limit int) string {
	if limit <= 0 {
		return value
	}
	if len(value) <= limit {
		return value
	}
	return value[:limit-1] + "…"
}

func previewForHistory(item list.Item, maxLines int) string {
	hist, ok := item.(historyItem)
	if !ok {
		return ""
	}
	if maxLines <= 0 {
		maxLines = 4
	}
	query := strings.TrimSpace(hist.entry.Query)
	if query == "" {
		return ""
	}
	highlighted := highlightSQL(query)
	lines := strings.Split(highlighted, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
		lines = append(lines, "…")
	}
	return strings.Join(lines, "\n")
}

func renderPreview(content string, height, width int) string {
	if content == "" {
		content = "History preview"
	}
	box := lipgloss.NewStyle().
		Border(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("238")).
		Padding(0, 1)
	return box.Width(width).Height(height).Render(content)
}
