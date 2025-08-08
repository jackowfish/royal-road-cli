package ui

import (
	"fmt"
	"royal-road-cli/internal/api"
	"strconv"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type searchModel struct {
	input       textinput.Model
	list        list.Model
	searching   bool
	err         error
	client      *api.Client
	fictions    []api.SearchFiction
	showResults bool
}

type searchResultsMsg []api.SearchFiction
type searchErrorMsg error

func NewSearchModel() searchModel {
	input := textinput.New()
	input.Placeholder = "Enter search terms..."
	input.Focus()

	items := []list.Item{}
	
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
	l.Title = "🔍 Search Results"
	l.StatusMessageLifetime = 0
	l.SetShowHelp(true)
	l.SetShowStatusBar(true)
	l.SetFilteringEnabled(false)

	return searchModel{
		input:  input,
		list:   l,
		client: api.NewClient(),
	}
}

func (m searchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m searchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.showResults {
			switch msg.String() {
			case "esc", "q":
				m.showResults = false
				return m, nil
			case "enter":
				if selected, ok := m.list.SelectedItem().(searchFictionItem); ok {
					readerModel := NewReaderModel(strconv.Itoa(selected.fiction.ID))
					return readerModel, readerModel.Init()
				}
			}
			m.list, cmd = m.list.Update(msg)
			return m, cmd
		} else {
			switch msg.String() {
			case "esc", "q":
				return NewMenuModel(), nil
			case "enter":
				if strings.TrimSpace(m.input.Value()) != "" {
					m.searching = true
					return m, m.search()
				}
			}
		}

	case searchResultsMsg:
		m.searching = false
		m.fictions = []api.SearchFiction(msg)
		items := make([]list.Item, len(m.fictions))
		for i, f := range m.fictions {
			items[i] = searchFictionItem{fiction: f}
		}
		m.list.SetItems(items)
		m.showResults = true
		return m, nil

	case searchErrorMsg:
		m.searching = false
		m.err = error(msg)
		return m, nil

	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2)
	}

	if !m.showResults {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m searchModel) View() string {
	if m.showResults {
		return m.list.View()
	}

	var s strings.Builder

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("205")).
		PaddingBottom(1)

	s.WriteString(titleStyle.Render("🔍 Search Royal Road Fictions"))
	s.WriteString("\n\n")
	s.WriteString(m.input.View())
	s.WriteString("\n\n")

	if m.searching {
		s.WriteString("Searching...")
	} else if m.err != nil {
		s.WriteString(fmt.Sprintf("Error: %v", m.err))
	} else {
		s.WriteString("Press Enter to search, Esc to go back")
	}

	return s.String()
}

func (m searchModel) search() tea.Cmd {
	query := strings.TrimSpace(m.input.Value())
	return func() tea.Msg {
		fictions, err := m.client.SearchFictions(query)
		if err != nil {
			return searchErrorMsg(err)
		}
		return searchResultsMsg(fictions)
	}
}

type searchFictionItem struct {
	fiction api.SearchFiction
}

func (i searchFictionItem) FilterValue() string {
	return i.fiction.Title
}

func (i searchFictionItem) Title() string {
	return i.fiction.Title
}

func (i searchFictionItem) Description() string {
	var parts []string

	if i.fiction.Author != "" {
		parts = append(parts, fmt.Sprintf("by %s", i.fiction.Author))
	}

	if i.fiction.Type != "" {
		parts = append(parts, i.fiction.Type)
	}

	if i.fiction.Status != "" {
		parts = append(parts, i.fiction.Status)
	}

	var statsStr strings.Builder
	if i.fiction.Stats.Rating > 0 {
		statsStr.WriteString(fmt.Sprintf("%.1f★", i.fiction.Stats.Rating))
	}
	if i.fiction.Stats.Pages > 0 {
		if statsStr.Len() > 0 {
			statsStr.WriteString(" • ")
		}
		if i.fiction.Stats.Pages >= 1000 {
			statsStr.WriteString(fmt.Sprintf("%.1fk pages", float64(i.fiction.Stats.Pages)/1000))
		} else {
			statsStr.WriteString(fmt.Sprintf("%d pages", i.fiction.Stats.Pages))
		}
	}
	if i.fiction.Stats.Followers > 0 {
		if statsStr.Len() > 0 {
			statsStr.WriteString(" • ")
		}
		if i.fiction.Stats.Followers >= 1000 {
			statsStr.WriteString(fmt.Sprintf("%.1fk followers", float64(i.fiction.Stats.Followers)/1000))
		} else {
			statsStr.WriteString(fmt.Sprintf("%d followers", i.fiction.Stats.Followers))
		}
	}

	if statsStr.Len() > 0 {
		parts = append(parts, statsStr.String())
	}

	if len(i.fiction.Tags) > 0 {
		maxTags := 2
		if len(i.fiction.Tags) < maxTags {
			maxTags = len(i.fiction.Tags)
		}
		tags := strings.Join(i.fiction.Tags[:maxTags], ", ")
		parts = append(parts, tags)
	}

	return strings.Join(parts, " • ")
}