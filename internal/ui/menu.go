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

type MenuModel struct {
	state       MenuState
	config      *config.Config
	client      *api.Client
	
	// History pagination
	historyPage     int
	historyPageSize int
	
	// Input fields
	fictionInput  textinput.Model
	chapterInput  textinput.Model
	
	// Status
	loading bool
	err     error
	
	// Results
	selectedEntry *config.ReadingEntry
}

func NewMenuModel() *MenuModel {
	cfg, _ := config.Load()
	
	fictionInput := textinput.New()
	fictionInput.Placeholder = "Enter fiction ID (e.g., 21220)"
	fictionInput.Focus()
	fictionInput.Width = 30
	
	chapterInput := textinput.New()
	chapterInput.Placeholder = "Enter chapter number (default: 1)"
	chapterInput.Width = 30
	
	return &MenuModel{
		state:           MenuStateMain,
		config:          cfg,
		client:          api.NewClient(),
		historyPage:     1,
		historyPageSize: 10,
		fictionInput:    fictionInput,
		chapterInput:    chapterInput,
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
		return m, nil
	}
	
	var cmd tea.Cmd
	m.fictionInput, cmd = m.fictionInput.Update(msg)
	return m, cmd
}

func (m *MenuModel) handleMainMenu(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c", "q":
		return m, tea.Quit
	case "c":
		// Continue last read
		if lastEntry := m.config.GetLastReadEntry(); lastEntry != nil {
			readerModel := NewReaderModel(lastEntry.FictionID)
			readerModel.SetStartChapter(lastEntry.CurrentChapter)
			return readerModel, readerModel.Init()
		}
	case "h":
		// Show history
		m.state = MenuStateHistory
		m.historyPage = 1
		return m, nil
	case "n":
		// New book
		m.state = MenuStateNewBook
		m.fictionInput.Focus()
		return m, nil
	case "b":
		// Browse popular
		browseModel := NewBrowseModel()
		return browseModel, browseModel.Init()
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
		// Select entry by number
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
		
		chapterNum := 1 // Default to chapter 1
		if chapterStr != "" {
			if num, err := strconv.Atoi(chapterStr); err == nil && num > 0 {
				chapterNum = num
			}
		}
		
		readerModel := NewReaderModel(fictionID)
		readerModel.SetStartChapter(chapterNum - 1) // Convert to 0-based index
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
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Padding(1, 0).
		Render("📚 Royal Road CLI")
	
	var options strings.Builder
	
	// Continue option
	if lastEntry := m.config.GetLastReadEntry(); lastEntry != nil {
		continueStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("120")).
			Bold(true)
		
		chapterProgress := fmt.Sprintf("(%d/%d", lastEntry.CurrentChapter+1, lastEntry.TotalChapters)
		if lastEntry.ChapterProgress > 0 {
			chapterProgress += fmt.Sprintf(", %.0f%% through chapter)", lastEntry.ChapterProgress*100)
		} else {
			chapterProgress += ")"
		}
		
		options.WriteString(continueStyle.Render(fmt.Sprintf("  [c] Continue: %s %s\n", lastEntry.FictionTitle, chapterProgress)))
		options.WriteString(fmt.Sprintf("      Chapter: %s\n\n", lastEntry.ChapterTitle))
	}
	
	// Other options
	options.WriteString("  [h] Reading History\n")
	options.WriteString("  [n] Start New Book\n") 
	options.WriteString("  [b] Browse Popular Fictions\n")
	options.WriteString("  [q] Quit\n")
	
	return fmt.Sprintf("%s\n\n%s", title, options.String())
}

func (m *MenuModel) viewHistoryMenu() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Render("📖 Reading History")
	
	entries, totalPages, hasNext, hasPrev := m.config.GetReadingHistoryPage(m.historyPage, m.historyPageSize)
	
	if len(entries) == 0 {
		return fmt.Sprintf("%s\n\nNo reading history found.\n\nPress [esc] to go back", title)
	}
	
	var content strings.Builder
	content.WriteString(fmt.Sprintf("%s\n\n", title))
	
	for i, entry := range entries {
		num := i + 1
		progress := fmt.Sprintf("(%d/%d", entry.CurrentChapter+1, entry.TotalChapters)
		if entry.ChapterProgress > 0 {
			progress += fmt.Sprintf(", %.0f%% through chapter)", entry.ChapterProgress*100)
		} else {
			progress += ")"
		}
		
		entryStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("150"))
		titleStyle := lipgloss.NewStyle().Bold(true)
		
		content.WriteString(fmt.Sprintf("  [%d] %s %s\n", num, titleStyle.Render(entry.FictionTitle), progress))
		content.WriteString(fmt.Sprintf("      %s • Chapter: %s\n", 
			entryStyle.Render("by "+entry.Author), entry.ChapterTitle))
		content.WriteString(fmt.Sprintf("      Last read: %s\n\n", entry.LastRead))
	}
	
	// Pagination info
	pageInfo := fmt.Sprintf("Page %d/%d", m.historyPage, totalPages)
	if hasPrev || hasNext {
		nav := ""
		if hasPrev {
			nav += "[←/h] prev"
		}
		if hasPrev && hasNext {
			nav += " • "
		}
		if hasNext {
			nav += "[→/l] next"
		}
		pageInfo += " • " + nav
	}
	
	content.WriteString(fmt.Sprintf("%s\n", pageInfo))
	content.WriteString("Press number to continue reading • [esc] back to main menu")
	
	return content.String()
}

func (m *MenuModel) viewNewBookInput() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Render("📖 Start New Book")
	
	return fmt.Sprintf("%s\n\nEnter Fiction ID:\n%s\n\nPress [enter] to continue or [esc] to go back",
		title, m.fictionInput.View())
}

func (m *MenuModel) viewNewChapterInput() string {
	title := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Render("📖 Start New Book")
	
	return fmt.Sprintf("%s\n\nFiction ID: %s\n\nStarting chapter (optional):\n%s\n\nPress [enter] to start reading or [esc] to go back",
		title, m.fictionInput.Value(), m.chapterInput.View())
}