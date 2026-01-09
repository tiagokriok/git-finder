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
	gitStatusData    *git.StatusData
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
	if m.gitStatusData == nil {
		return m.renderGitStatusError()
	}
	return m.renderDetailedGitStatus()
}

func (m Model) renderGitStatusError() string {
	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("205"))
	title := titleStyle.Render("📊 Git Status")

	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("240")).
		Padding(1, 2).
		Width(min(m.width-4, 80))

	errorStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	errorMsg := errorStyle.Render(m.gitStatusContent)

	content := borderStyle.Render(errorMsg)

	footerStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240")).Align(lipgloss.Center)
	footer := footerStyle.Render("Esc or q: close")

	fullView := lipgloss.JoinVertical(lipgloss.Center, title, "", content, "", footer)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, fullView)
}

func (m Model) renderDetailedGitStatus() string {
	accentColor := lipgloss.Color("205")
	headerColor := lipgloss.Color("46")
	statColor := lipgloss.Color("240")
	borderColor := lipgloss.Color("240")

	data := m.gitStatusData

	// ========== BRANCH HEADER ==========
	branchStyle := lipgloss.NewStyle().
		Foreground(headerColor).
		Bold(true).
		Padding(0, 1)

	branchHeader := branchStyle.Render(fmt.Sprintf("🌿 %s", data.CurrentBranch))

	// ========== TRACKING BRANCH ==========
	var trackingLine string
	if data.TrackingBranch != "" {
		trackingStyle := lipgloss.NewStyle().
			Foreground(statColor).
			Padding(0, 1)
		trackingLine = trackingStyle.Render(fmt.Sprintf("└─ tracking: %s", data.TrackingBranch))
	}

	// ========== STATS SECTION ==========
	statsStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("250"))

	var statLines []string

	// Ahead/Behind
	if data.AheadCount > 0 || data.BehindCount > 0 {
		aheadBehind := ""
		if data.AheadCount > 0 {
			aheadStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("46"))
			aheadBehind += aheadStyle.Render(fmt.Sprintf("⬆ %d", data.AheadCount))
		}
		if data.BehindCount > 0 {
			if aheadBehind != "" {
				aheadBehind += "  "
			}
			behindStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
			aheadBehind += behindStyle.Render(fmt.Sprintf("⬇ %d", data.BehindCount))
		}
		statLines = append(statLines, statsStyle.Render(aheadBehind))
	}

	// File change summary
	changeCount := data.ModifiedCount + data.AddedCount + data.DeletedCount + data.UntrackedCount

	if changeCount > 0 {
		summary := fmt.Sprintf("📊 %d file%s changed", changeCount, m.pluralize(changeCount))

		breakdown := ""
		if data.AddedCount > 0 {
			breakdown += lipgloss.NewStyle().Foreground(lipgloss.Color("46")).Render(fmt.Sprintf("+%d", data.AddedCount)) + " "
		}
		if data.ModifiedCount > 0 {
			breakdown += lipgloss.NewStyle().Foreground(lipgloss.Color("220")).Render(fmt.Sprintf("~%d", data.ModifiedCount)) + " "
		}
		if data.DeletedCount > 0 {
			breakdown += lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render(fmt.Sprintf("-%d", data.DeletedCount))
		}
		if data.UntrackedCount > 0 {
			if breakdown != "" {
				breakdown += " "
			}
			breakdown += lipgloss.NewStyle().Foreground(lipgloss.Color("33")).Render(fmt.Sprintf("?%d", data.UntrackedCount))
		}
		if breakdown != "" {
			summary += " (" + breakdown + ")"
		}

		statLines = append(statLines, statsStyle.Render(summary))
	}

	// Stash count
	if data.StashCount > 0 {
		stashStyle := lipgloss.NewStyle().Padding(0, 1).Foreground(lipgloss.Color("175"))
		stashLine := stashStyle.Render(fmt.Sprintf("📦 %d stashed", data.StashCount))
		statLines = append(statLines, stashLine)
	}

	statsSection := strings.Join(statLines, "\n")

	// ========== DIVIDER ==========
	dividerStyle := lipgloss.NewStyle().Foreground(borderColor)
	divider := dividerStyle.Render(strings.Repeat("─", min(m.width-8, 76)))

	// ========== FILES SECTION ==========
	filesTitle := lipgloss.NewStyle().
		Foreground(accentColor).
		Bold(true).
		Padding(0, 1).
		Render("📝 Files")

	var fileLines []string
	visibleHeight := min(m.height-14, 20)
	maxFiles := min(len(data.Files)-m.gitStatusScroll, visibleHeight)

	if len(data.Files) == 0 {
		cleanStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true).
			Padding(0, 1)
		fileLines = append(fileLines, cleanStyle.Render("✓ Working tree clean"))
	} else {
		statusColors := map[string]lipgloss.Color{
			"M":  lipgloss.Color("220"),  // Yellow for modified
			"A":  lipgloss.Color("46"),   // Green for added
			"D":  lipgloss.Color("196"),  // Red for deleted
			"R":  lipgloss.Color("171"),  // Magenta for renamed
			"C":  lipgloss.Color("51"),   // Cyan for copied
			"??": lipgloss.Color("33"),   // Blue for untracked
		}

		statusSymbols := map[string]string{
			"M":  "✏️ ",
			"A":  "✨",
			"D":  "🗑️ ",
			"R":  "↪️ ",
			"C":  "📋",
			"??": "❓",
		}

		for i := 0; i < maxFiles && i < visibleHeight; i++ {
			file := data.Files[m.gitStatusScroll+i]
			status := file.Status
			color := statusColors[status]
			symbol := statusSymbols[status]
			if symbol == "" {
				symbol = "📄"
			}

			fileStyle := lipgloss.NewStyle().
				Foreground(color).
				Padding(0, 1)

			fileLine := fmt.Sprintf("%s %s  %s", symbol, status, file.Filename)
			fileLines = append(fileLines, fileStyle.Render(fileLine))
		}
	}

	filesContent := strings.Join(fileLines, "\n")

	// ========== SCROLL INDICATOR ==========
	var scrollIndicator string
	if len(data.Files) > visibleHeight {
		scrollStyle := lipgloss.NewStyle().Foreground(borderColor).Italic(true)
		current := min(m.gitStatusScroll+visibleHeight, len(data.Files))
		scrollIndicator = scrollStyle.Render(
			fmt.Sprintf("(showing %d-%d of %d)", m.gitStatusScroll+1, current, len(data.Files)),
		)
	}

	// ========== BORDER STYLING ==========
	borderStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(borderColor).
		Padding(1, 2).
		Width(min(m.width-4, 100))

	// ========== ASSEMBLE CONTENT ==========
	headerSection := lipgloss.JoinVertical(
		lipgloss.Left,
		branchHeader,
		trackingLine,
	)

	contentLines := []string{
		headerSection,
		"",
		statsSection,
		divider,
		filesTitle,
		filesContent,
	}

	if scrollIndicator != "" {
		contentLines = append(contentLines, "", scrollIndicator)
	}

	content := lipgloss.JoinVertical(lipgloss.Left, contentLines...)
	mainContent := borderStyle.Render(content)

	// ========== FOOTER ==========
	footerStyle := lipgloss.NewStyle().
		Foreground(borderColor).
		Align(lipgloss.Center).
		Padding(0, 1)

	footer := footerStyle.Render("↑/↓ j/k: scroll • q/Esc: close")

	// ========== COMBINE ==========
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(accentColor).
		Padding(0, 1)

	title := titleStyle.Render("📊 Git Status")

	fullView := lipgloss.JoinVertical(
		lipgloss.Center,
		title,
		"",
		mainContent,
		"",
		footer,
	)

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, fullView)
}

func (m *Model) pluralize(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
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
	statusData, err := git.GetDetailedStatus(repoPath)
	if err != nil {
		// Show error in modal instead of failing
		m.gitStatusContent = fmt.Sprintf("⚠ Error fetching git status:\n\n%s", err.Error())
		m.gitStatusData = nil
	} else {
		m.gitStatusData = statusData
		m.gitStatusContent = ""
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
