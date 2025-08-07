package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"royal-road-cli/internal/api"
)

type FictionListItem struct {
	fiction api.PopularFiction
}

func (f FictionListItem) Title() string {
	return f.fiction.Title
}

func (f FictionListItem) Description() string {
	author := f.fiction.Author
	if author == "" {
		author = "Unknown Author"
	}
	
	tags := ""
	if len(f.fiction.Tags) > 0 {
		tags = " • " + strings.Join(f.fiction.Tags[:min(3, len(f.fiction.Tags))], ", ")
	}
	
	return fmt.Sprintf("%s%s", author, tags)
}

func (f FictionListItem) FilterValue() string {
	return f.fiction.Title + " " + f.fiction.Author
}

type BrowseModel struct {
	list      list.Model
	client    *api.Client
	loading   bool
	err       error
}

type fictionsLoadedMsg []api.PopularFiction
type errorMsg error

func NewBrowseModel() *BrowseModel {
	items := []list.Item{}
	
	// Get terminal size for proper initialization
	termWidth, termHeight := getTerminalSize()
	
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3)
	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("170")).
		Bold(true).
		Padding(0, 0, 0, 1)
	
	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("170")).
		Foreground(lipgloss.Color("240")).
		Padding(0, 0, 0, 1)

	l := list.New(items, delegate, termWidth, termHeight-2)
	l.Title = "📚 Popular Royal Road Fictions"
	l.StatusMessageLifetime = 0
	l.SetShowHelp(true)
	l.SetFilteringEnabled(true)

	return &BrowseModel{
		list:    l,
		client:  api.NewClient(),
		loading: true,
	}
}

func (m *BrowseModel) Init() tea.Cmd {
	return m.loadFictions()
}

func (m *BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2) // Leave space for status line
		return m, nil
	
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			if item, ok := m.list.SelectedItem().(FictionListItem); ok {
				readerModel := NewReaderModel(fmt.Sprintf("%d", item.fiction.ID))
				return readerModel, readerModel.Init()
			}
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadFictions()
		}
	
	case fictionsLoadedMsg:
		m.loading = false
		items := make([]list.Item, len(msg))
		for i, fiction := range msg {
			items[i] = FictionListItem{fiction: fiction}
		}
		m.list.SetItems(items)
		return m, nil
	
	case errorMsg:
		m.loading = false
		m.err = msg
		return m, nil
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m *BrowseModel) View() string {
	if m.loading {
		return lipgloss.NewStyle().
			Padding(2).
			Render("🔄 Loading popular fictions...")
	}
	
	if m.err != nil {
		return lipgloss.NewStyle().
			Padding(2).
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("❌ Error loading fictions: %v\n\nPress 'r' to retry or 'q' to quit.", m.err))
	}
	
	return m.list.View()
}

func (m *BrowseModel) loadFictions() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		fictions, err := m.client.GetPopularFictions()
		if err != nil {
			return errorMsg(err)
		}
		return fictionsLoadedMsg(fictions)
	})
}

