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

type SearchModel struct {
	input       textinput.Model
	list        list.Model
	searching   bool
	err         error
	client      *api.Client
	fictions    []api.SearchFiction
	showResults bool
	width       int
	height      int
}

type searchResultsMsg []api.SearchFiction
type searchErrorMsg error

func NewSearchModel() SearchModel {
	termWidth, termHeight := getTerminalSize()

	input := textinput.New()
	input.Placeholder = "Enter search terms..."
	input.Prompt = ""
	input.Focus()
	input.Width = min(60, termWidth-8) - 6 // Account for border + padding

	delegate := NewFictionDelegate()

	l := list.New([]list.Item{}, delegate, termWidth, termHeight-4)
	l.Title = ""
	l.SetShowTitle(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()

	return SearchModel{
		input:  input,
		list:   l,
		client: api.NewClient(),
		width:  termWidth,
		height: termHeight,
	}
}

func (m SearchModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m SearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4)
		m.input.Width = min(60, msg.Width-8) - 6
		return m, nil

	case tea.KeyMsg:
		if m.showResults {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
				m.showResults = false
				m.input.Focus()
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
			case "ctrl+c", "q":
				return m, tea.Quit
			case "esc":
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
	}

	if !m.showResults {
		m.input, cmd = m.input.Update(msg)
	}
	return m, cmd
}

func (m SearchModel) View() string {
	s := NewStyles()

	// Determine context
	context := "Search"
	if m.showResults {
		context = fmt.Sprintf("Results for \"%s\"", m.input.Value())
	}

	// Header
	header := Header("Royal Road CLI", context, m.width)

	var content strings.Builder

	if m.showResults {
		// Show results list
		content.WriteString(m.list.View())
	} else {
		// Show search input
		content.WriteString("\n")
		content.WriteString(s.Title.Render("  Search Royal Road"))
		content.WriteString("\n\n")

		content.WriteString(s.Text.Render("  Enter your search query:"))
		content.WriteString("\n\n")

		// Styled input with fixed width to prevent border misalignment
		inputWidth := min(60, m.width-8)
		inputContent := lipgloss.Place(inputWidth-4, 1, lipgloss.Left, lipgloss.Center, m.input.View())
		inputStyle := lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(CurrentTheme.Primary).
			Padding(0, 1).
			Width(inputWidth).
			MarginLeft(2)

		content.WriteString(inputStyle.Render(inputContent))
		content.WriteString("\n\n")

		if m.searching {
			content.WriteString("  " + LoadingMessage("Searching...", 0))
		} else if m.err != nil {
			content.WriteString("  " + ErrorMessage(m.err))
		} else {
			content.WriteString(s.TextMuted.Render("  Press Enter to search"))
		}
	}

	// Footer
	var bindings []KeyBinding
	if m.showResults {
		bindings = []KeyBinding{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "read"},
			{Key: "esc", Desc: "new search"},
		}
	} else {
		bindings = []KeyBinding{
			{Key: "enter", Desc: "search"},
			{Key: "esc", Desc: "back to menu"},
		}
	}
	footer := Footer(bindings, m.width)

	// Calculate content height
	contentHeight := m.height - 4
	contentArea := lipgloss.NewStyle().
		Height(contentHeight).
		Render(content.String())

	return header + "\n" + contentArea + "\n" + footer
}

func (m SearchModel) search() tea.Cmd {
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
