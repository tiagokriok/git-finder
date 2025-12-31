package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/tiagokriok/gf/internal/scanner"
)

var selectedRepository *scanner.Repository

const (
	maxHeight       = 10
	boxPadding      = 2
	searchBoxHeight = 3
	footerHeight    = 3
	searchBoxWidth  = 50
)

type Model struct {
	repositories []scanner.Repository
	filtered     []scanner.Repository
	searchInput  string
	selectedIdx  int
	width        int
	height       int
	err          error
}

func NewModel(repos []scanner.Repository) Model {
	return Model{
		repositories: repos,
		filtered:     repos,
		selectedIdx:  0,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	}
	return m, nil
}

func (m Model) View() string {
	// headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))

	searchBoxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(0, 1).Width(searchBoxWidth).Align(lipgloss.Left)

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)

	availableHeight := m.height - footerHeight - 4
	if availableHeight < 3 {
		availableHeight = 3
	}

	// header := headerStyle.Render("🔍 Git Finder")
	searchLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Search:")
	searchInput := m.searchInput
	searchBox := searchBoxStyle.Render(searchInput)

	var reposList string
	if len(m.filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		reposList = emptyStyle.Render("No repositories found")
	} else {
		var lines []string

		itemsToShow := len(m.filtered)
		if itemsToShow > maxHeight {
			itemsToShow = maxHeight
		}

		for i := 0; i < itemsToShow && i < availableHeight; i++ {
			repo := m.filtered[i]
			line := fmt.Sprintf("%s (%s)", repo.Name, repo.Path)

			if i == m.selectedIdx {
				lines = append(lines, selectedStyle.Render("▶ "+line))
			} else {
				lines = append(lines, "  "+line)
			}
		}

		reposList = strings.Join(lines, "\n")
	}
	paginationInfo := m.getPaginationInfo()
	paginationStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
	pagination := paginationStyle.Render(paginationInfo)

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Center)
	footer := footerStyle.Render("↑/↓: navigate | Enter: select | Esc: exit")

	content := lipgloss.JoinVertical(lipgloss.Center, searchLabel, searchBox, "", reposList, pagination)

	centered := lipgloss.Place(m.width, m.height-footerHeight, lipgloss.Center, lipgloss.Center, content)

	fullView := lipgloss.JoinVertical(lipgloss.Left, centered, footer)

	return fullView
}

func (m *Model) updateFiltered() {
	if m.searchInput == "" {
		m.filtered = m.repositories
		return
	}

	names := make([]string, len(m.repositories))
	for i, repo := range m.repositories {
		names[i] = repo.Name
	}

	matches := fuzzy.Find(m.searchInput, names)

	m.filtered = make([]scanner.Repository, len(matches))
	for i, match := range matches {
		m.filtered[i] = m.repositories[match.Index]
	}
}

func (m Model) getPaginationInfo() string {
	if len(m.filtered) == 0 {
		return ""
	}

	itemsToShow := len(m.filtered)
	if itemsToShow > maxHeight {
		itemsToShow = maxHeight
	}

	if len(m.filtered) <= maxHeight {
		return fmt.Sprintf("(%d results)", len(m.filtered))
	}

	return fmt.Sprintf("Showing %d of %d", itemsToShow, len(m.filtered))
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "esc":
		selectedRepository = nil
		return m, tea.Quit
	case "up", "shift+tab":
		if m.selectedIdx > 0 {
			m.selectedIdx--
		}
		return m, nil
	case "down", "tab":
		if m.selectedIdx < len(m.filtered)-1 {
			m.selectedIdx++
		}
		return m, nil
	case "enter":
		if len(m.filtered) > 0 {
			selected := m.filtered[m.selectedIdx]
			selectedRepository = &selected
			return m, tea.Quit
		}
		return m, nil
	case "backspace":
		if len(m.searchInput) > 0 {
			m.searchInput = m.searchInput[:len(m.searchInput)-1]
			m.updateFiltered()
			m.selectedIdx = 0
		}
		return m, nil
	default:
		m.searchInput += msg.String()
		m.updateFiltered()
		m.selectedIdx = 0
		return m, nil
	}
}

func GetSelectedRepository() *scanner.Repository {
	return selectedRepository
}

func Run(repos []scanner.Repository) (*scanner.Repository, error) {
	selectedRepository = nil

	model := NewModel(repos)

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("TUI Error: %w", err)
	}

	return selectedRepository, nil
}
