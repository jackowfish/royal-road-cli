package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// KeyBinding represents a keyboard shortcut and its description
type KeyBinding struct {
	Key  string
	Desc string
}

// Header renders a consistent header bar across all screens
func Header(title, context string, width int) string {
	s := NewStyles()

	// App icon and title on the left
	left := s.HeaderTitle.Render("📚 Royal Road CLI")

	// Context/screen name on the right
	right := s.HeaderInfo.Render(context)

	// Calculate spacing
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := width - leftWidth - rightWidth - 4 // 4 for padding

	if spacing < 0 {
		spacing = 0
	}

	spacer := strings.Repeat(" ", spacing)

	content := left + spacer + right

	return lipgloss.NewStyle().
		Width(width).
		Background(CurrentTheme.Surface).
		Foreground(CurrentTheme.Text).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Bottom: "─"}).
		BorderForeground(CurrentTheme.Border).
		BorderBottom(true).
		Render(content)
}

// Footer renders a consistent footer bar with keybindings
func Footer(bindings []KeyBinding, width int) string {
	s := NewStyles()

	var parts []string
	for _, b := range bindings {
		key := s.FooterKey.Render(b.Key)
		desc := s.FooterText.Render(" " + b.Desc)
		parts = append(parts, key+desc)
	}

	content := strings.Join(parts, s.TextMuted.Render(" │ "))

	return lipgloss.NewStyle().
		Width(width).
		Background(CurrentTheme.Surface).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Top: "─"}).
		BorderForeground(CurrentTheme.Border).
		BorderTop(true).
		Render(content)
}

// FooterWithProgress renders a footer with keybindings and a progress indicator
func FooterWithProgress(bindings []KeyBinding, progress string, width int) string {
	s := NewStyles()

	var parts []string
	for _, b := range bindings {
		key := s.FooterKey.Render(b.Key)
		desc := s.FooterText.Render(" " + b.Desc)
		parts = append(parts, key+desc)
	}

	left := strings.Join(parts, s.TextMuted.Render(" │ "))
	right := s.HeaderInfo.Render(progress)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	spacing := width - leftWidth - rightWidth - 4

	if spacing < 0 {
		spacing = 0
		// Truncate if needed
		if len(parts) > 2 {
			parts = parts[:2]
			left = strings.Join(parts, s.TextMuted.Render(" │ "))
			leftWidth = lipgloss.Width(left)
			spacing = width - leftWidth - rightWidth - 4
			if spacing < 0 {
				spacing = 0
			}
		}
	}

	spacer := strings.Repeat(" ", spacing)
	content := left + spacer + right

	return lipgloss.NewStyle().
		Width(width).
		Background(CurrentTheme.Surface).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Top: "─"}).
		BorderForeground(CurrentTheme.Border).
		BorderTop(true).
		Render(content)
}

// Panel renders a bordered panel with optional title
func Panel(content, title string, width, height int) string {
	s := NewStyles()

	innerWidth := width - 4 // Account for border and padding

	// Wrap content to fit
	wrappedContent := lipgloss.NewStyle().
		Width(innerWidth).
		Render(content)

	panel := s.Panel.
		Width(width).
		Height(height).
		Render(wrappedContent)

	if title != "" {
		// Add title to the top border
		titleStr := s.PanelTitle.Render(" " + title + " ")
		lines := strings.Split(panel, "\n")
		if len(lines) > 0 {
			// Insert title into the top border
			topBorder := lines[0]
			titleWidth := lipgloss.Width(titleStr)
			if len(topBorder) > titleWidth+4 {
				// Replace part of the top border with the title
				lines[0] = topBorder[:2] + titleStr + topBorder[2+titleWidth:]
			}
			panel = strings.Join(lines, "\n")
		}
	}

	return panel
}

// ContentPanel renders a panel for main content with subtle borders
func ContentPanel(content string, width, height int) string {
	innerWidth := width - 4
	innerHeight := height - 2

	// Ensure content fits
	wrappedContent := lipgloss.NewStyle().
		Width(innerWidth).
		MaxHeight(innerHeight).
		Render(content)

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(SubtleBorder()).
		BorderForeground(CurrentTheme.Border).
		Padding(0, 1).
		Render(wrappedContent)
}

// MenuItem renders a single menu item with key and description
func MenuItem(key, title, description string, active bool) string {
	s := NewStyles()

	keyStyle := s.MenuKey.Copy()
	titleStyle := s.Text.Copy()
	descStyle := s.MenuDescription.Copy()

	if active {
		keyStyle = keyStyle.Background(CurrentTheme.Surface)
		titleStyle = titleStyle.Background(CurrentTheme.Surface).Bold(true).Foreground(CurrentTheme.Primary)
		descStyle = descStyle.Background(CurrentTheme.Surface)
	}

	keyPart := keyStyle.Render("[" + key + "]")
	titlePart := titleStyle.Render(" " + title)

	line := keyPart + titlePart
	if description != "" {
		descPart := descStyle.Render(" - " + description)
		line += descPart
	}

	return line
}

// LoadingSpinner returns a simple loading indicator
func LoadingSpinner(frame int) string {
	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	s := NewStyles()
	return s.Highlight.Render(frames[frame%len(frames)])
}

// LoadingMessage renders a loading state with spinner
func LoadingMessage(message string, frame int) string {
	s := NewStyles()
	spinner := LoadingSpinner(frame)
	return spinner + " " + s.TextMuted.Render(message)
}

// ErrorMessage renders an error message
func ErrorMessage(err error) string {
	s := NewStyles()
	return s.Error.Render("✗ Error: ") + s.Text.Render(err.Error())
}

// SuccessMessage renders a success message
func SuccessMessage(message string) string {
	s := NewStyles()
	return s.Success.Render("✓ ") + s.Text.Render(message)
}

// ProgressBar renders a simple progress bar
func ProgressBar(current, total, width int) string {
	if total == 0 {
		total = 1
	}

	percentage := float64(current) / float64(total)
	filled := int(float64(width-2) * percentage)
	empty := width - 2 - filled

	if filled < 0 {
		filled = 0
	}
	if empty < 0 {
		empty = 0
	}

	bar := "│" + strings.Repeat("█", filled) + strings.Repeat("░", empty) + "│"

	s := NewStyles()
	return s.Highlight.Render(bar) + s.TextMuted.Render(fmt.Sprintf(" %d/%d", current, total))
}

// Divider renders a horizontal divider
func Divider(width int) string {
	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Border).
		Render(strings.Repeat("─", width))
}

// TruncateWithEllipsis truncates a string to a max length with ellipsis
func TruncateWithEllipsis(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// CenterText centers text within a given width
func CenterText(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := (width - textWidth) / 2
	return strings.Repeat(" ", padding) + text
}

// RightAlign aligns text to the right within a given width
func RightAlign(text string, width int) string {
	textWidth := lipgloss.Width(text)
	if textWidth >= width {
		return text
	}
	padding := width - textWidth
	return strings.Repeat(" ", padding) + text
}

// Box creates a box around content with a title
func Box(content, title string, width int) string {
	innerWidth := width - 2

	// Create top border with title
	var topBorder string
	if title != "" {
		titlePart := "┤ " + title + " ├"
		remainingWidth := innerWidth - len(titlePart)
		leftWidth := 1
		rightWidth := remainingWidth - leftWidth
		if rightWidth < 0 {
			rightWidth = 0
		}
		topBorder = "╭" + strings.Repeat("─", leftWidth) + titlePart + strings.Repeat("─", rightWidth) + "╮"
	} else {
		topBorder = "╭" + strings.Repeat("─", innerWidth) + "╮"
	}

	// Bottom border
	bottomBorder := "╰" + strings.Repeat("─", innerWidth) + "╯"

	// Wrap content
	lines := strings.Split(content, "\n")
	var boxedLines []string
	boxedLines = append(boxedLines, topBorder)

	for _, line := range lines {
		lineWidth := lipgloss.Width(line)
		padding := innerWidth - lineWidth
		if padding < 0 {
			padding = 0
			line = line[:innerWidth]
		}
		boxedLines = append(boxedLines, "│"+line+strings.Repeat(" ", padding)+"│")
	}

	boxedLines = append(boxedLines, bottomBorder)

	return lipgloss.NewStyle().
		Foreground(CurrentTheme.Border).
		Render(strings.Join(boxedLines, "\n"))
}

// StatusBadge renders a colored status badge
func StatusBadge(status string) string {
	s := NewStyles()

	switch strings.ToLower(status) {
	case "ongoing", "active", "reading":
		return s.Success.Render("● " + status)
	case "completed", "finished":
		return s.Highlight.Render("✓ " + status)
	case "hiatus", "paused":
		return s.Warning.Render("◐ " + status)
	case "dropped", "cancelled":
		return s.Error.Render("✗ " + status)
	default:
		return s.TextMuted.Render("○ " + status)
	}
}

// HelpView renders a styled help overlay
func HelpView(sections map[string][]KeyBinding, width, height int) string {
	s := NewStyles()

	var content strings.Builder

	content.WriteString(s.Title.Render("Keyboard Shortcuts"))
	content.WriteString("\n\n")

	for section, bindings := range sections {
		content.WriteString(s.Subtitle.Render(section))
		content.WriteString("\n")

		for _, b := range bindings {
			key := s.MenuKey.Render(fmt.Sprintf("%-12s", b.Key))
			desc := s.Text.Render(b.Desc)
			content.WriteString("  " + key + " " + desc + "\n")
		}
		content.WriteString("\n")
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 2).
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(CurrentTheme.Primary).
		Render(content.String())
}

// TabBar renders a horizontal tab bar
func TabBar(tabs []string, activeIndex, width int) string {
	s := NewStyles()

	var parts []string
	for i, tab := range tabs {
		if i == activeIndex {
			parts = append(parts, s.Highlight.Render(" "+tab+" "))
		} else {
			parts = append(parts, s.TextMuted.Render(" "+tab+" "))
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		Background(CurrentTheme.Surface).
		BorderStyle(lipgloss.Border{Bottom: "─"}).
		BorderForeground(CurrentTheme.Border).
		BorderBottom(true).
		Render(strings.Join(parts, s.TextMuted.Render("│")))
}
