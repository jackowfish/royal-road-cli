package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"royal-road-cli/internal/config"
)

// ThemeSelectorModel handles theme selection at first launch
type ThemeSelectorModel struct {
	themes       []Theme
	selectedIdx  int
	config       *config.Config
	width        int
	height       int
	confirmed    bool
}

// ThemeSelectedMsg is sent when a theme is confirmed
type ThemeSelectedMsg struct {
	Theme Theme
}

// NewThemeSelectorModel creates a new theme selector
func NewThemeSelectorModel() *ThemeSelectorModel {
	cfg, _ := config.Load()
	width, height := getTerminalSize()

	return &ThemeSelectorModel{
		themes:      AllThemes,
		selectedIdx: 0,
		config:      cfg,
		width:       width,
		height:      height,
	}
}

func (m *ThemeSelectorModel) Init() tea.Cmd {
	return nil
}

func (m *ThemeSelectorModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit

		case "up", "k":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, nil

		case "down", "j":
			if m.selectedIdx < len(m.themes)-1 {
				m.selectedIdx++
			}
			return m, nil

		case "left", "h":
			if m.selectedIdx > 0 {
				m.selectedIdx--
			}
			return m, nil

		case "right", "l":
			if m.selectedIdx < len(m.themes)-1 {
				m.selectedIdx++
			}
			return m, nil

		case "enter", " ":
			// Confirm selection
			selectedTheme := m.themes[m.selectedIdx]
			SetTheme(selectedTheme)
			m.config.SetTheme(selectedTheme.Name)
			if err := m.config.Save(); err != nil {
				// Log error but continue; theme selection is still applied in memory
				_ = err
			}
			m.confirmed = true

			// Return to main menu
			menuModel := NewMenuModel()
			return menuModel, menuModel.Init()
		}
	}

	return m, nil
}

func (m *ThemeSelectorModel) View() string {
	// Temporarily set theme for preview
	previewTheme := m.themes[m.selectedIdx]
	oldTheme := CurrentTheme
	SetTheme(previewTheme)
	defer SetTheme(oldTheme)

	var content strings.Builder

	// Welcome header
	headerStyle := lipgloss.NewStyle().
		Foreground(previewTheme.Primary).
		Bold(true).
		Padding(1, 0)

	content.WriteString(headerStyle.Render("Welcome to Royal Road CLI"))
	content.WriteString("\n\n")

	// Subtitle
	subtitleStyle := lipgloss.NewStyle().
		Foreground(previewTheme.TextMuted)

	content.WriteString(subtitleStyle.Render("Choose a theme to get started. Use arrow keys to preview, Enter to confirm."))
	content.WriteString("\n\n")

	// Theme selector
	content.WriteString(m.renderThemeSelector(previewTheme))
	content.WriteString("\n\n")

	// Preview panel
	content.WriteString(m.renderPreview(previewTheme))

	// Footer
	content.WriteString("\n\n")
	footerStyle := lipgloss.NewStyle().
		Foreground(previewTheme.TextMuted)
	content.WriteString(footerStyle.Render("Press Enter to confirm selection • q to quit"))

	// Center everything
	return lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Padding(2, 4).
		Render(content.String())
}

func (m *ThemeSelectorModel) renderThemeSelector(theme Theme) string {
	var parts []string

	themeNames := map[string]string{
		"purple": "Purple",
		"blue":   "Blue",
		"green":  "Green",
		"mono":   "Mono",
	}

	for i, t := range m.themes {
		name := themeNames[t.Name]
		if name == "" {
			name = t.Name
		}

		var style lipgloss.Style
		if i == m.selectedIdx {
			// Selected theme - show with highlight
			style = lipgloss.NewStyle().
				Foreground(t.Primary).
				Background(theme.Surface).
				Bold(true).
				Padding(0, 2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(t.Primary)
		} else {
			// Non-selected theme
			style = lipgloss.NewStyle().
				Foreground(theme.TextMuted).
				Padding(0, 2).
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(theme.Border)
		}

		// Color swatch
		swatch := lipgloss.NewStyle().
			Foreground(t.Primary).
			Render("●")

		parts = append(parts, style.Render(swatch+" "+name))
	}

	return lipgloss.JoinHorizontal(lipgloss.Center, parts...)
}

func (m *ThemeSelectorModel) renderPreview(theme Theme) string {
	s := NewStyles()

	previewWidth := min(m.width-8, 70)
	previewHeight := min(m.height-15, 20)

	var preview strings.Builder

	// Simulated header
	headerContent := s.HeaderTitle.Render("Royal Road CLI") +
		strings.Repeat(" ", max(0, previewWidth-35)) +
		s.HeaderInfo.Render("Main Menu")

	headerBar := lipgloss.NewStyle().
		Width(previewWidth).
		Background(theme.Surface).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Bottom: "─"}).
		BorderForeground(theme.Border).
		BorderBottom(true).
		Render(headerContent)

	preview.WriteString(headerBar)
	preview.WriteString("\n\n")

	// Sample menu items
	menuItems := []struct {
		key   string
		label string
		desc  string
	}{
		{"c", "Continue Reading", "The Primal Hunter - Ch 234/1500"},
		{"h", "Reading History", "View your reading history"},
		{"n", "New Book", "Start reading a new fiction"},
		{"b", "Browse", "Explore popular fictions"},
	}

	for i, item := range menuItems {
		keyStyle := s.MenuKey.Copy()
		labelStyle := s.Text.Copy()
		descStyle := s.MenuDescription.Copy()

		if i == 0 {
			// Simulate active state
			keyStyle = keyStyle.Background(theme.Surface)
			labelStyle = labelStyle.Background(theme.Surface).Foreground(theme.Primary).Bold(true)
			descStyle = descStyle.Background(theme.Surface)
		}

		line := keyStyle.Render("["+item.key+"]") + " " +
			labelStyle.Render(item.label) + " " +
			descStyle.Render("- "+item.desc)

		preview.WriteString("  " + line + "\n")
	}

	preview.WriteString("\n")

	// Sample content box
	contentBox := lipgloss.NewStyle().
		Width(previewWidth - 4).
		Padding(0, 1).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Border).
		Render(s.TextMuted.Render("Sample text content would appear here.\nThis is how your reading experience will look."))

	preview.WriteString(contentBox)
	preview.WriteString("\n\n")

	// Simulated footer
	footerKeys := []struct {
		key  string
		desc string
	}{
		{"↑↓", "navigate"},
		{"enter", "select"},
		{"q", "quit"},
	}

	var footerParts []string
	for _, k := range footerKeys {
		footerParts = append(footerParts,
			s.FooterKey.Render(k.key)+" "+s.FooterText.Render(k.desc))
	}

	footerContent := strings.Join(footerParts, s.TextMuted.Render(" │ "))
	footerBar := lipgloss.NewStyle().
		Width(previewWidth).
		Background(theme.Surface).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Top: "─"}).
		BorderForeground(theme.Border).
		BorderTop(true).
		Render(footerContent)

	preview.WriteString(footerBar)

	// Wrap in a preview box
	previewBox := lipgloss.NewStyle().
		Width(previewWidth + 4).
		Height(previewHeight).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(theme.Primary).
		Padding(1, 1).
		Render(preview.String())

	// Add theme name label
	themeName := map[string]string{
		"purple": "Purple/Magenta Theme",
		"blue":   "Blue/Cyan Theme",
		"green":  "Green/Teal Theme",
		"mono":   "Minimal/Monochrome Theme",
	}[theme.Name]

	if themeName == "" {
		themeName = theme.Name + " Theme"
	}

	label := lipgloss.NewStyle().
		Foreground(theme.Primary).
		Bold(true).
		Render("Preview: " + themeName)

	return label + "\n" + previewBox
}

// GetThemeDisplayName returns a human-readable theme name
func GetThemeDisplayName(name string) string {
	names := map[string]string{
		"purple": "Purple/Magenta",
		"blue":   "Blue/Cyan",
		"green":  "Green/Teal",
		"mono":   "Minimal/Mono",
	}
	if display, ok := names[name]; ok {
		return display
	}
	return name
}

// ThemeColorSwatch returns a colored dot representing the theme
func ThemeColorSwatch(t Theme) string {
	return lipgloss.NewStyle().
		Foreground(t.Primary).
		Render("●")
}

// RenderAllThemeSwatches shows all themes inline
func RenderAllThemeSwatches(selectedName string) string {
	var parts []string
	for _, t := range AllThemes {
		swatch := ThemeColorSwatch(t)
		name := GetThemeDisplayName(t.Name)

		if t.Name == selectedName {
			parts = append(parts, fmt.Sprintf("[%s %s]", swatch, name))
		} else {
			parts = append(parts, fmt.Sprintf(" %s %s ", swatch, name))
		}
	}
	return strings.Join(parts, " ")
}
