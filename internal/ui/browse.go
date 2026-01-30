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

	return fmt.Sprintf("by %s%s", author, tags)
}

func (f FictionListItem) FilterValue() string {
	return f.fiction.Title + " " + f.fiction.Author
}

type BrowseModel struct {
	list    list.Model
	client  *api.Client
	loading bool
	err     error
	width   int
	height  int
}

type fictionsLoadedMsg []api.PopularFiction
type errorMsg error

func NewBrowseModel() *BrowseModel {
	items := []list.Item{}

	termWidth, termHeight := getTerminalSize()

	delegate := NewFictionDelegate()

	l := list.New(items, delegate, termWidth, termHeight-4)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.DisableQuitKeybindings()

	// Style the filter input
	l.FilterInput.PromptStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Primary)
	l.FilterInput.TextStyle = lipgloss.NewStyle().Foreground(CurrentTheme.Text)
	l.FilterInput.Cursor.Style = lipgloss.NewStyle().Foreground(CurrentTheme.Primary)

	return &BrowseModel{
		list:    l,
		client:  api.NewClient(),
		loading: true,
		width:   termWidth,
		height:  termHeight,
	}
}

func (m *BrowseModel) Init() tea.Cmd {
	return m.loadFictions()
}

func (m *BrowseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		return m, nil

	case tea.KeyMsg:
		// Don't handle keys if filtering
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "esc":
			// Go back to menu
			menuModel := NewMenuModel()
			return menuModel, menuModel.Init()
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
	s := NewStyles()

	// Header
	header := Header("Royal Road CLI", "Browse Popular", m.width)

	// Content
	var content string

	if m.loading {
		content = lipgloss.NewStyle().
			Padding(2).
			Render(LoadingMessage("Loading popular fictions...", 0))
	} else if m.err != nil {
		content = lipgloss.NewStyle().
			Padding(2).
			Render(ErrorMessage(m.err) + "\n\n" + s.TextMuted.Render("Press 'r' to retry"))
	} else {
		content = m.list.View()
	}

	// Footer
	bindings := []KeyBinding{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "/", Desc: "filter"},
		{Key: "enter", Desc: "read"},
		{Key: "r", Desc: "refresh"},
		{Key: "esc", Desc: "back"},
	}
	footer := Footer(bindings, m.width)

	// Calculate content height
	contentHeight := m.height - 4
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Render(content)

	return header + "\n" + contentArea + "\n" + footer
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
