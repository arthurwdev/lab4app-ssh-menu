// Package ui implements the interactive bubbletea menu that lists SSH hosts
// and lets the user pick one to connect to.
package ui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/arthurwdev/lab4app-ssh-menu/internal/sshconfig"
)

// Lines reserved outside the scrollable host list: title line, divider,
// column header, blank spacer before the footer, and the footer itself.
const reservedLines = 5

// defaultWidth is used until the terminal reports its real size via the
// first WindowSizeMsg, so the very first frame isn't rendered blank.
const defaultWidth = 80

// Model is the bubbletea model driving the host selection menu.
type Model struct {
	hosts    []sshconfig.Host
	version  string
	hostname string
	username string
	styles   Styles

	columns     columns
	cursor      int
	offset      int
	visibleRows int
	width       int

	selected string
	quitting bool
}

// New builds the initial model for the given hosts and header metadata.
func New(hosts []sshconfig.Host, version, hostname, username string) Model {
	m := Model{
		hosts:    hosts,
		version:  version,
		hostname: hostname,
		username: username,
		styles:   NewStyles(),
		width:    defaultWidth,
	}
	m.columns = computeColumns(hosts, m.width)
	m.visibleRows = max(len(hosts), 1)
	return m
}

// Selected returns the alias the user chose, or "" if they cancelled.
func (m Model) Selected() string {
	return m.selected
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.columns = computeColumns(m.hosts, m.width)
		m.visibleRows = max(msg.Height-reservedLines, 1)
		m.clampOffset()
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case "enter":
			if len(m.hosts) > 0 {
				m.selected = m.hosts[m.cursor].Alias
			}
			m.quitting = true
			return m, tea.Quit
		case "up", "k":
			if len(m.hosts) > 0 {
				if m.cursor > 0 {
					m.cursor--
				} else {
					m.cursor = len(m.hosts) - 1
				}
			}
			m.clampOffset()
		case "down", "j":
			if len(m.hosts) > 0 {
				if m.cursor < len(m.hosts)-1 {
					m.cursor++
				} else {
					m.cursor = 0
				}
			}
			m.clampOffset()
		}
	}
	return m, nil
}

func (m *Model) clampOffset() {
	if m.visibleRows <= 0 {
		return
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+m.visibleRows {
		m.offset = m.cursor - m.visibleRows + 1
	}
	maxOffset := max(len(m.hosts)-m.visibleRows, 0)
	m.offset = min(max(m.offset, 0), maxOffset)
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if len(m.hosts) == 0 {
		return m.styles.Muted.Render(" No hosts found in the SSH config.\n")
	}

	var b strings.Builder
	b.WriteString(renderHeader(m.version, m.username, m.hostname, m.width, m.styles))
	b.WriteString("\n")
	b.WriteString(renderHeaderRow(m.columns, m.styles))
	b.WriteString("\n")

	end := min(m.offset+m.visibleRows, len(m.hosts))
	for i := m.offset; i < end; i++ {
		b.WriteString(renderRow(m.hosts[i], m.columns, i == m.cursor, m.styles))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(renderFooter(m.styles))
	return b.String()
}
