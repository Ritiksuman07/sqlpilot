package editor

import (
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
		body = m.historyList.View()
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
	return h.entry.Query
}

func (h historyItem) Description() string {
	if h.entry.CreatedAt.IsZero() {
		return ""
	}
	return h.entry.CreatedAt.Format("2006-01-02 15:04:05")
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
