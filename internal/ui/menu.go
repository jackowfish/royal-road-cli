package ui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"royal-road-cli/internal/api"
	"royal-road-cli/internal/config"
)

type MenuState int

const (
	MenuStateMain MenuState = iota
	MenuStateHistory
	MenuStateNewBook
	MenuStateNewChapter
)

// MenuItemData represents a menu option with key, label, description, and optional extra info.
type MenuItemData struct {
	Key      string
	Label    string
	Desc     string
	Enabled  bool
	HasExtra bool
	Extra    string
}

type MenuModel struct {
	state  MenuState
	config *config.Config
	client *api.Client

	// History pagination
	historyPage     int
	historyPageSize int

	// Input fields
	fictionInput textinput.Model
	chapterInput textinput.Model

	// Status
	loading bool
	err     error

	// Results
	selectedEntry *config.ReadingEntry

	// Terminal size
	width  int
	height int

	// Menu navigation
	menuIndex int
}

func NewMenuModel() *MenuModel {
	cfg, _ := config.Load()

	// Apply saved theme
	if cfg.ThemeName != "" {
		SetTheme(GetThemeByName(cfg.ThemeName))
	}

	fictionInput := textinput.New()
	fictionInput.Placeholder = "Enter fiction ID (e.g., 21220)"
	fictionInput.Focus()
	fictionInput.Width = 40

	chapterInput := textinput.New()
	chapterInput.Placeholder = "Enter chapter number (default: 1)"
	chapterInput.Width = 40

	width, height := getTerminalSize()

	return &MenuModel{
		state:           MenuStateMain,
		config:          cfg,
		client:          api.NewClient(),
		historyPage:     1,
		historyPageSize: 10,
		fictionInput:    fictionInput,
		chapterInput:    chapterInput,
		width:           width,
		height:          height,
		menuIndex:       0,
	}
}

func (m *MenuModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m *MenuModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case MenuStateMain:
			return m.handleMainMenu(msg)
		case MenuStateHistory:
			return m.handleHistoryMenu(msg)
		case MenuStateNewBook:
			return m.handleNewBookInput(msg)
		case MenuStateNewChapter:
			return m.handleNewChapterInput(msg)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	}

	var cmd tea.Cmd
	m.fictionInput, cmd = m.fictionInput.Update(msg)
	return m, cmd
}

func (m *MenuModel) getMenuItems() []MenuItemData {
	items := []MenuItemData{}

	// Continue reading option
	if lastEntry := m.config.GetLastReadEntry(); lastEntry != nil {
		progress := fmt.Sprintf("Ch %d/%d", lastEntry.CurrentChapter+1, lastEntry.TotalChapters)
		if lastEntry.ChapterProgress > 0 {
			progress += fmt.Sprintf(", %.0f%%", lastEntry.ChapterProgress*100)
		}
		items = append(items, MenuItemData{
			Key:      "c",
			Label:    "Continue Reading",
			Desc:     lastEntry.FictionTitle + " (" + progress + ")",
			Enabled:  true,
			HasExtra: true,
			Extra:    lastEntry.ChapterTitle,
		})
	}

	items = append(items,
		MenuItemData{Key: "h", Label: "Reading History", Desc: "View your reading history", Enabled: true},
		MenuItemData{Key: "n", Label: "Start New Book", Desc: "Enter a fiction ID to start reading", Enabled: true},
		MenuItemData{Key: "b", Label: "Browse Popular", Desc: "Explore popular fictions", Enabled: true},
		MenuItemData{Key: "s", Label: "Search", Desc: "Search for fictions by title", Enabled: true},
	)

	return items
}

func (m *MenuModel) handleMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	menuItems := m.getMenuItems()

	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "up", "k":
		if m.menuIndex > 0 {
			m.menuIndex--
		}
		return m, nil
	case "down", "j":
		if m.menuIndex < len(menuItems)-1 {
			m.menuIndex++
		}
		return m, nil
	case "enter":
		// Select current item
		if m.menuIndex < len(menuItems) {
			return m.executeMenuItem(menuItems[m.menuIndex].Key)
		}
	case "c":
		return m.executeMenuItem("c")
	case "h":
		return m.executeMenuItem("h")
	case "n":
		return m.executeMenuItem("n")
	case "b":
		return m.executeMenuItem("b")
	case "s":
		return m.executeMenuItem("s")
	}
	return m, nil
}

func (m *MenuModel) executeMenuItem(key string) (tea.Model, tea.Cmd) {
	switch key {
	case "c":
		if lastEntry := m.config.GetLastReadEntry(); lastEntry != nil {
			readerModel := NewReaderModel(lastEntry.FictionID)
			readerModel.SetStartChapter(lastEntry.CurrentChapter)
			return readerModel, readerModel.Init()
		}
	case "h":
		m.state = MenuStateHistory
		m.historyPage = 1
	case "n":
		m.state = MenuStateNewBook
		m.fictionInput.Focus()
	case "b":
		browseModel := NewBrowseModel()
		return browseModel, browseModel.Init()
	case "s":
		searchModel := NewSearchModel()
		return searchModel, searchModel.Init()
	}
	return m, nil
}

func (m *MenuModel) handleHistoryMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.state = MenuStateMain
		return m, nil
	case "left", "h":
		if m.historyPage > 1 {
			m.historyPage--
		}
		return m, nil
	case "right", "l":
		_, totalPages, hasNext, _ := m.config.GetReadingHistoryPage(m.historyPage, m.historyPageSize)
		if hasNext && m.historyPage < totalPages {
			m.historyPage++
		}
		return m, nil
	case "1", "2", "3", "4", "5", "6", "7", "8", "9":
		num, _ := strconv.Atoi(msg.String())
		entries, _, _, _ := m.config.GetReadingHistoryPage(m.historyPage, m.historyPageSize)
		if num > 0 && num <= len(entries) {
			entry := entries[num-1]
			readerModel := NewReaderModel(entry.FictionID)
			readerModel.SetStartChapter(entry.CurrentChapter)
			return readerModel, readerModel.Init()
		}
		return m, nil
	}
	return m, nil
}

func (m *MenuModel) handleNewBookInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.state = MenuStateMain
		m.fictionInput.SetValue("")
		return m, nil
	case "enter":
		if m.fictionInput.Value() != "" {
			m.state = MenuStateNewChapter
			m.chapterInput.Focus()
			m.fictionInput.Blur()
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.fictionInput, cmd = m.fictionInput.Update(msg)
	return m, cmd
}

func (m *MenuModel) handleNewChapterInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "esc":
		m.state = MenuStateNewBook
		m.chapterInput.SetValue("")
		m.chapterInput.Blur()
		m.fictionInput.Focus()
		return m, nil
	case "enter":
		fictionID := m.fictionInput.Value()
		chapterStr := m.chapterInput.Value()

		chapterNum := 1
		if chapterStr != "" {
			if num, err := strconv.Atoi(chapterStr); err == nil && num > 0 {
				chapterNum = num
			}
		}

		readerModel := NewReaderModel(fictionID)
		readerModel.SetStartChapter(chapterNum - 1)
		return readerModel, readerModel.Init()
	}

	var cmd tea.Cmd
	m.chapterInput, cmd = m.chapterInput.Update(msg)
	return m, cmd
}

func (m *MenuModel) View() string {
	switch m.state {
	case MenuStateMain:
		return m.viewMainMenu()
	case MenuStateHistory:
		return m.viewHistoryMenu()
	case MenuStateNewBook:
		return m.viewNewBookInput()
	case MenuStateNewChapter:
		return m.viewNewChapterInput()
	}
	return ""
}

func (m *MenuModel) viewMainMenu() string {
	s := NewStyles()

	// Header
	header := Header("Royal Road CLI", "Main Menu", m.width)

	// Content area
	var content strings.Builder

	// Title
	content.WriteString("\n")
	content.WriteString(s.Title.Render("Welcome to Royal Road CLI"))
	content.WriteString("\n")
	content.WriteString(s.TextMuted.Render("Your terminal-based fiction reader"))
	content.WriteString("\n\n")

	// Menu items
	menuItems := m.getMenuItems()

	for i, item := range menuItems {
		isActive := i == m.menuIndex

		// Key badge
		keyStyle := s.MenuKey.Copy()
		if isActive {
			keyStyle = keyStyle.Background(CurrentTheme.Surface)
		}
		key := keyStyle.Render("[" + item.Key + "]")

		// Label
		labelStyle := s.Text.Copy()
		if isActive {
			labelStyle = labelStyle.
				Foreground(CurrentTheme.Primary).
				Background(CurrentTheme.Surface).
				Bold(true)
		}
		label := labelStyle.Render(" " + item.Label)

		// Description
		descStyle := s.MenuDescription.Copy()
		if isActive {
			descStyle = descStyle.Background(CurrentTheme.Surface)
		}
		desc := descStyle.Render(" - " + item.Desc)

		line := "  " + key + label + desc

		// Pad the entire line if active
		if isActive {
			lineWidth := lipgloss.Width(line)
			padding := m.width - lineWidth - 4
			if padding > 0 {
				line = line + lipgloss.NewStyle().
					Background(CurrentTheme.Surface).
					Render(strings.Repeat(" ", padding))
			}
		}

		content.WriteString(line)
		content.WriteString("\n")

		// Extra info (like chapter title for continue)
		if item.HasExtra && item.Extra != "" {
			extraStyle := s.TextMuted.Copy()
			if isActive {
				extraStyle = extraStyle.Background(CurrentTheme.Surface)
			}
			extra := "       " + extraStyle.Render(item.Extra)
			if isActive {
				extraWidth := lipgloss.Width(extra)
				padding := m.width - extraWidth - 4
				if padding > 0 {
					extra = extra + lipgloss.NewStyle().
						Background(CurrentTheme.Surface).
						Render(strings.Repeat(" ", padding))
				}
			}
			content.WriteString(extra)
			content.WriteString("\n")
		}
	}

	// Footer
	footer := Footer([]KeyBinding{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "enter", Desc: "select"},
		{Key: "q", Desc: "quit"},
	}, m.width)

	// Combine all parts
	contentHeight := m.height - 4 // Subtract header and footer height
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Render(content.String())

	return header + "\n" + contentArea + "\n" + footer
}

func (m *MenuModel) viewHistoryMenu() string {
	s := NewStyles()

	// Header
	header := Header("Royal Road CLI", "Reading History", m.width)

	entries, totalPages, hasNext, hasPrev := m.config.GetReadingHistoryPage(m.historyPage, m.historyPageSize)

	var content strings.Builder
	content.WriteString("\n")

	if len(entries) == 0 {
		content.WriteString(s.TextMuted.Render("  No reading history found."))
		content.WriteString("\n\n")
		content.WriteString(s.TextMuted.Render("  Start reading a book to build your history!"))
	} else {
		for i, entry := range entries {
			num := i + 1

			// Number badge
			numStyle := s.MenuKey
			numStr := numStyle.Render(fmt.Sprintf("[%d]", num))

			// Title
			titleStyle := s.TextBold
			title := titleStyle.Render(" " + entry.FictionTitle)

			// Progress
			progress := fmt.Sprintf("(%d/%d", entry.CurrentChapter+1, entry.TotalChapters)
			if entry.ChapterProgress > 0 {
				progress += fmt.Sprintf(", %.0f%%", entry.ChapterProgress*100)
			}
			progress += ")"
			progressStr := s.TextMuted.Render(" " + progress)

			content.WriteString("  " + numStr + title + progressStr + "\n")

			// Author and chapter
			author := s.TextMuted.Render("      by " + entry.Author)
			chapter := s.TextMuted.Render(" • " + entry.ChapterTitle)
			content.WriteString(author + chapter + "\n")

			// Last read
			lastRead := s.TextMuted.Render("      Last read: " + entry.LastRead)
			content.WriteString(lastRead + "\n\n")
		}

		// Pagination
		pageInfo := s.TextMuted.Render(fmt.Sprintf("Page %d/%d", m.historyPage, totalPages))
		content.WriteString("  " + pageInfo + "\n")
	}

	// Footer bindings
	bindings := []KeyBinding{
		{Key: "1-9", Desc: "select"},
		{Key: "esc", Desc: "back"},
	}
	if hasPrev {
		bindings = append(bindings, KeyBinding{Key: "←", Desc: "prev page"})
	}
	if hasNext {
		bindings = append(bindings, KeyBinding{Key: "→", Desc: "next page"})
	}

	footer := Footer(bindings, m.width)

	contentHeight := m.height - 4
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Render(content.String())

	return header + "\n" + contentArea + "\n" + footer
}

func (m *MenuModel) viewNewBookInput() string {
	s := NewStyles()

	// Header
	header := Header("Royal Road CLI", "Start New Book", m.width)

	var content strings.Builder
	content.WriteString("\n")
	content.WriteString(s.Title.Render("  Start a New Book"))
	content.WriteString("\n\n")

	content.WriteString(s.Text.Render("  Enter the fiction ID from the Royal Road URL:"))
	content.WriteString("\n")
	content.WriteString(s.TextMuted.Render("  (e.g., royalroad.com/fiction/21220/... → ID is 21220)"))
	content.WriteString("\n\n")

	// Styled input
	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Primary).
		Padding(0, 1).
		Width(50)

	content.WriteString("  " + inputStyle.Render(m.fictionInput.View()))

	footer := Footer([]KeyBinding{
		{Key: "enter", Desc: "continue"},
		{Key: "esc", Desc: "back"},
	}, m.width)

	contentHeight := m.height - 4
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Render(content.String())

	return header + "\n" + contentArea + "\n" + footer
}

func (m *MenuModel) viewNewChapterInput() string {
	s := NewStyles()

	// Header
	header := Header("Royal Road CLI", "Start New Book", m.width)

	var content strings.Builder
	content.WriteString("\n")
	content.WriteString(s.Title.Render("  Start a New Book"))
	content.WriteString("\n\n")

	content.WriteString(s.Text.Render("  Fiction ID: "))
	content.WriteString(s.Highlight.Render(m.fictionInput.Value()))
	content.WriteString("\n\n")

	content.WriteString(s.Text.Render("  Starting chapter (optional, defaults to 1):"))
	content.WriteString("\n\n")

	inputStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Primary).
		Padding(0, 1).
		Width(50)

	content.WriteString("  " + inputStyle.Render(m.chapterInput.View()))

	footer := Footer([]KeyBinding{
		{Key: "enter", Desc: "start reading"},
		{Key: "esc", Desc: "back"},
	}, m.width)

	contentHeight := m.height - 4
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Render(content.String())

	return header + "\n" + contentArea + "\n" + footer
}
