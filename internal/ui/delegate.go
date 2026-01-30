package ui

import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/lipgloss"
)

// NewFictionDelegate creates a styled list delegate for fiction items.
// Used by both browse and search views for consistent styling.
func NewFictionDelegate() list.DefaultDelegate {
	delegate := list.NewDefaultDelegate()
	delegate.SetHeight(3)
	delegate.SetSpacing(0)

	delegate.Styles.SelectedTitle = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(CurrentTheme.Primary).
		Foreground(CurrentTheme.Primary).
		Bold(true).
		Padding(0, 0, 0, 1)

	delegate.Styles.SelectedDesc = lipgloss.NewStyle().
		BorderLeft(true).
		BorderStyle(lipgloss.ThickBorder()).
		BorderForeground(CurrentTheme.Primary).
		Foreground(CurrentTheme.TextMuted).
		Padding(0, 0, 0, 1)

	delegate.Styles.NormalTitle = lipgloss.NewStyle().
		Foreground(CurrentTheme.Text).
		Padding(0, 0, 0, 2)

	delegate.Styles.NormalDesc = lipgloss.NewStyle().
		Foreground(CurrentTheme.TextMuted).
		Padding(0, 0, 0, 2)

	delegate.Styles.DimmedTitle = lipgloss.NewStyle().
		Foreground(CurrentTheme.TextMuted).
		Padding(0, 0, 0, 2)

	delegate.Styles.DimmedDesc = lipgloss.NewStyle().
		Foreground(CurrentTheme.Border).
		Padding(0, 0, 0, 2)

	return delegate
}
