package ui

import "github.com/charmbracelet/lipgloss"

// Theme defines a color palette for the TUI
type Theme struct {
	Name       string
	Primary    lipgloss.Color // Main accent color
	Secondary  lipgloss.Color // Secondary accent
	Background lipgloss.Color // Dark background
	Surface    lipgloss.Color // Slightly lighter bg
	Text       lipgloss.Color // Main text
	TextMuted  lipgloss.Color // Dimmed text
	Border     lipgloss.Color // Border color
	Success    lipgloss.Color // Success/green
	Error      lipgloss.Color // Error/red
	Warning    lipgloss.Color // Warning/yellow
}

// Available themes
var (
	ThemePurple = Theme{
		Name:       "purple",
		Primary:    lipgloss.Color("170"),
		Secondary:  lipgloss.Color("213"),
		Background: lipgloss.Color("235"),
		Surface:    lipgloss.Color("237"),
		Text:       lipgloss.Color("252"),
		TextMuted:  lipgloss.Color("245"),
		Border:     lipgloss.Color("240"),
		Success:    lipgloss.Color("120"),
		Error:      lipgloss.Color("196"),
		Warning:    lipgloss.Color("214"),
	}

	ThemeBlue = Theme{
		Name:       "blue",
		Primary:    lipgloss.Color("39"),
		Secondary:  lipgloss.Color("51"),
		Background: lipgloss.Color("234"),
		Surface:    lipgloss.Color("236"),
		Text:       lipgloss.Color("252"),
		TextMuted:  lipgloss.Color("245"),
		Border:     lipgloss.Color("240"),
		Success:    lipgloss.Color("42"),
		Error:      lipgloss.Color("196"),
		Warning:    lipgloss.Color("214"),
	}

	ThemeGreen = Theme{
		Name:       "green",
		Primary:    lipgloss.Color("42"),
		Secondary:  lipgloss.Color("48"),
		Background: lipgloss.Color("233"),
		Surface:    lipgloss.Color("235"),
		Text:       lipgloss.Color("252"),
		TextMuted:  lipgloss.Color("245"),
		Border:     lipgloss.Color("238"),
		Success:    lipgloss.Color("42"),
		Error:      lipgloss.Color("196"),
		Warning:    lipgloss.Color("214"),
	}

	ThemeMono = Theme{
		Name:       "mono",
		Primary:    lipgloss.Color("255"),
		Secondary:  lipgloss.Color("250"),
		Background: lipgloss.Color("233"),
		Surface:    lipgloss.Color("235"),
		Text:       lipgloss.Color("252"),
		TextMuted:  lipgloss.Color("245"),
		Border:     lipgloss.Color("240"),
		Success:    lipgloss.Color("252"),
		Error:      lipgloss.Color("196"),
		Warning:    lipgloss.Color("214"),
	}

	// AllThemes is a list of all available themes
	AllThemes = []Theme{ThemePurple, ThemeBlue, ThemeGreen, ThemeMono}

	// CurrentTheme is the active theme (default to purple)
	CurrentTheme = ThemePurple
)

// GetThemeByName returns a theme by its name
func GetThemeByName(name string) Theme {
	for _, t := range AllThemes {
		if t.Name == name {
			return t
		}
	}
	return ThemePurple
}

// SetTheme sets the current theme
func SetTheme(t Theme) {
	CurrentTheme = t
}

// Styles provides pre-configured lipgloss styles using the current theme
type Styles struct {
	// Base styles
	App         lipgloss.Style
	Header      lipgloss.Style
	HeaderTitle lipgloss.Style
	HeaderInfo  lipgloss.Style
	Footer      lipgloss.Style
	FooterKey   lipgloss.Style
	FooterText  lipgloss.Style

	// Panel styles
	Panel       lipgloss.Style
	PanelTitle  lipgloss.Style
	PanelBorder lipgloss.Style

	// List styles
	ListItem         lipgloss.Style
	ListItemSelected lipgloss.Style
	ListItemTitle    lipgloss.Style
	ListItemDesc     lipgloss.Style

	// Text styles
	Title      lipgloss.Style
	Subtitle   lipgloss.Style
	Text       lipgloss.Style
	TextMuted  lipgloss.Style
	TextBold   lipgloss.Style
	Highlight  lipgloss.Style
	Success    lipgloss.Style
	Error      lipgloss.Style
	Warning    lipgloss.Style

	// Input styles
	Input       lipgloss.Style
	InputFocus  lipgloss.Style
	Placeholder lipgloss.Style

	// Menu styles
	MenuItem        lipgloss.Style
	MenuItemActive  lipgloss.Style
	MenuKey         lipgloss.Style
	MenuDescription lipgloss.Style
}

// NewStyles creates a new Styles instance using the current theme
func NewStyles() Styles {
	t := CurrentTheme

	return Styles{
		// Base styles
		App: lipgloss.NewStyle().
			Background(t.Background),

		Header: lipgloss.NewStyle().
			Background(t.Surface).
			Foreground(t.Text).
			Padding(0, 1).
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			BorderBottom(true),

		HeaderTitle: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		HeaderInfo: lipgloss.NewStyle().
			Foreground(t.TextMuted),

		Footer: lipgloss.NewStyle().
			Background(t.Surface).
			Foreground(t.TextMuted).
			Padding(0, 1),

		FooterKey: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		FooterText: lipgloss.NewStyle().
			Foreground(t.TextMuted),

		// Panel styles
		Panel: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(0, 1),

		PanelTitle: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true).
			Padding(0, 1),

		PanelBorder: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border),

		// List styles
		ListItem: lipgloss.NewStyle().
			Foreground(t.Text).
			Padding(0, 1),

		ListItemSelected: lipgloss.NewStyle().
			Foreground(t.Primary).
			Background(t.Surface).
			Bold(true).
			Padding(0, 1),

		ListItemTitle: lipgloss.NewStyle().
			Foreground(t.Text).
			Bold(true),

		ListItemDesc: lipgloss.NewStyle().
			Foreground(t.TextMuted),

		// Text styles
		Title: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		Subtitle: lipgloss.NewStyle().
			Foreground(t.Secondary).
			Italic(true),

		Text: lipgloss.NewStyle().
			Foreground(t.Text),

		TextMuted: lipgloss.NewStyle().
			Foreground(t.TextMuted),

		TextBold: lipgloss.NewStyle().
			Foreground(t.Text).
			Bold(true),

		Highlight: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		Success: lipgloss.NewStyle().
			Foreground(t.Success),

		Error: lipgloss.NewStyle().
			Foreground(t.Error),

		Warning: lipgloss.NewStyle().
			Foreground(t.Warning),

		// Input styles
		Input: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Border).
			Padding(0, 1),

		InputFocus: lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(t.Primary).
			Padding(0, 1),

		Placeholder: lipgloss.NewStyle().
			Foreground(t.TextMuted).
			Italic(true),

		// Menu styles
		MenuItem: lipgloss.NewStyle().
			Foreground(t.Text).
			Padding(0, 2),

		MenuItemActive: lipgloss.NewStyle().
			Foreground(t.Primary).
			Background(t.Surface).
			Bold(true).
			Padding(0, 2),

		MenuKey: lipgloss.NewStyle().
			Foreground(t.Primary).
			Bold(true),

		MenuDescription: lipgloss.NewStyle().
			Foreground(t.TextMuted),
	}
}

// Theme-aware border style
func ThemeBorder() lipgloss.Border {
	return lipgloss.RoundedBorder()
}

// Subtle border for reader content
func SubtleBorder() lipgloss.Border {
	return lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}
}
