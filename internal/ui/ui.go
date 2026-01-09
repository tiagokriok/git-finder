package ui

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
	"github.com/tiagokriok/Git-Fuzzy/internal/config"
	"github.com/tiagokriok/Git-Fuzzy/internal/git"
	"github.com/tiagokriok/Git-Fuzzy/internal/platform"
	"github.com/tiagokriok/Git-Fuzzy/internal/scanner"
)

var selectedRepository *scanner.Repository

const (
	maxHeight       = 10
	boxPadding      = 2
	searchBoxHeight = 3
	footerHeight    = 3
	searchBoxWidth  = 50
)

type ViewMode int

const (
	ViewModeNormal ViewMode = iota
	ViewModeGitStatus
)

type Model struct {
	repositories     []scanner.Repository
	filtered         []scanner.Repository
	searchInput      string
	selectedIdx      int
	scrollOffset     int
	width            int
	height           int
	err              error
	viewMode         ViewMode
	gitStatusContent string
	gitStatusScroll  int
	config           *config.Config
}

func NewModel(repos []scanner.Repository, cfg *config.Config) Model {
	return Model{
		repositories: repos,
		filtered:     repos,
		selectedIdx:  0,
		viewMode:     ViewModeNormal,
		config:       cfg,
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
	if m.viewMode == ViewModeGitStatus {
		return m.viewGitStatus()
	}
	return m.viewNormal()
}

func (m Model) viewNormal() string {

	searchBoxStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("205")).Padding(0, 1).Width(searchBoxWidth).Align(lipgloss.Left)

	selectedStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Bold(true)

	availableHeight := max(m.height-footerHeight-4, 3)

	searchLabel := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Render("Search:")
	searchInput := m.searchInput
	searchBox := searchBoxStyle.Render(searchInput)

	var reposList string
	if len(m.filtered) == 0 {
		emptyStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Italic(true)
		reposList = emptyStyle.Render("No repositories found")
	} else {
		var lines []string

		itemsToShow := min(len(m.filtered), maxHeight)

		if m.selectedIdx < m.scrollOffset {
			m.scrollOffset = m.selectedIdx
		}
		if m.selectedIdx >= m.scrollOffset+itemsToShow {
			m.scrollOffset = m.selectedIdx - itemsToShow + 1
		}

		for i := 0; i < itemsToShow && i < availableHeight; i++ {
			repoIdx := m.scrollOffset + i
			if repoIdx >= len(m.filtered) {
				break
			}
			repo := m.filtered[repoIdx]
			displayPath := formatRepoPath(repo.Path)
			line := fmt.Sprintf("%s (%s)", repo.Name, displayPath)

			if repoIdx == m.selectedIdx {
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
	footer := footerStyle.Render("↑/↓: nav | Enter: editor | ^O: files | ^T: term | ^B: remote | ^G: status | Esc: exit")

	content := lipgloss.JoinVertical(lipgloss.Center, searchLabel, searchBox, "", reposList, pagination)

	centered := lipgloss.Place(m.width, m.height-footerHeight, lipgloss.Center, lipgloss.Center, content)

	fullView := lipgloss.JoinVertical(lipgloss.Left, centered, footer)

	return fullView
}

func (m Model) viewGitStatus() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	title := titleStyle.Render("📊 Git Status")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("205")).
		Padding(1, 2).
		Width(min(m.width-4, 80))

	// Handle scrolling
	lines := strings.Split(m.gitStatusContent, "\n")
	visibleHeight := min(m.height-10, 20)

	if m.gitStatusScroll >= len(lines)-visibleHeight {
		m.gitStatusScroll = max(0, len(lines)-visibleHeight)
	}

	endIdx := min(m.gitStatusScroll+visibleHeight, len(lines))
	visibleContent := strings.Join(lines[m.gitStatusScroll:endIdx], "\n")

	content := borderStyle.Render(visibleContent)

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Center)
	footer := footerStyle.Render("↑/↓ or j/k: scroll | Esc or q: close")

	fullView := lipgloss.JoinVertical(lipgloss.Center, title, "", content, "", footer)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, fullView)
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

	itemsToShow := min(len(m.filtered), maxHeight)

	if len(m.filtered) <= maxHeight {
		return fmt.Sprintf("(%d results)", len(m.filtered))
	}

	return fmt.Sprintf("Showing %d of %d", itemsToShow, len(m.filtered))
}

func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Git status overlay navigation
	if m.viewMode == ViewModeGitStatus {
		switch msg.String() {
		case "esc", "q":
			m.viewMode = ViewModeNormal
			m.gitStatusContent = ""
			m.gitStatusScroll = 0
			return m, nil
		case "up", "k":
			if m.gitStatusScroll > 0 {
				m.gitStatusScroll--
			}
			return m, nil
		case "down", "j":
			m.gitStatusScroll++
			return m, nil
		}
		return m, nil
	}

	// Normal mode keyboard handling
	switch msg.String() {
	case "ctrl+c", "esc":
		selectedRepository = nil
		return m, tea.Quit

	case "ctrl+o": // Open file manager
		if len(m.filtered) > 0 {
			selected := m.filtered[m.selectedIdx]
			m.openFileManager(selected.Path)
		}
		return m, nil

	case "ctrl+t": // Open terminal
		if len(m.filtered) > 0 {
			selected := m.filtered[m.selectedIdx]
			m.openTerminal(selected.Path)
		}
		return m, nil

	case "ctrl+b": // Open in browser
		if len(m.filtered) > 0 {
			selected := m.filtered[m.selectedIdx]
			m.openInBrowser(selected.Path)
		}
		return m, nil

	case "ctrl+g": // Show git status
		if len(m.filtered) > 0 {
			selected := m.filtered[m.selectedIdx]
			m.showGitStatus(selected.Path)
		}
		return m, nil

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
			m.scrollOffset = 0
		}
		return m, nil

	default:
		m.searchInput += msg.String()
		m.updateFiltered()
		m.selectedIdx = 0
		m.scrollOffset = 0
		return m, nil
	}
}

func formatRepoPath(fullPath string) string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fullPath
	}

	if strings.HasPrefix(fullPath, homeDir) {
		return strings.TrimPrefix(fullPath, homeDir+"/")
	}
	return fullPath
}

func (m *Model) openFileManager(repoPath string) {
	cmd := m.config.GetFileManager()
	if cmd == "" {
		return
	}

	// Fire and forget
	go func() {
		exec.Command(cmd, repoPath).Start()
	}()
}

func (m *Model) openTerminal(repoPath string) {
	cmd := m.config.GetTerminal()
	if cmd == "" {
		return
	}

	// Fire and forget
	go func() {
		parts := strings.Fields(cmd)
		if len(parts) > 0 {
			c := exec.Command(parts[0], append(parts[1:], repoPath)...)
			c.Dir = repoPath
			c.Start()
		}
	}()
}

func (m *Model) openInBrowser(repoPath string) {
	remoteURL, err := git.GetRemoteURL(repoPath)
	if err != nil {
		// Silently fail - no remote configured
		return
	}

	httpsURL, err := git.ConvertToHTTPS(remoteURL)
	if err != nil {
		return
	}

	platform.OpenInBrowser(httpsURL)
}

func (m *Model) showGitStatus(repoPath string) {
	status, err := git.GetStatus(repoPath)
	if err != nil {
		// Show error in modal instead of failing
		m.gitStatusContent = fmt.Sprintf("⚠ Error fetching git status:\n\n%s", err.Error())
	} else {
		m.gitStatusContent = status
	}
	m.viewMode = ViewModeGitStatus
	m.gitStatusScroll = 0
}

func GetSelectedRepository() *scanner.Repository {
	return selectedRepository
}

func Run(repos []scanner.Repository, cfg *config.Config) (*scanner.Repository, error) {
	selectedRepository = nil

	model := NewModel(repos, cfg)

	p := tea.NewProgram(model)

	if _, err := p.Run(); err != nil {
		return nil, fmt.Errorf("TUI Error: %w", err)
	}

	return selectedRepository, nil
}
