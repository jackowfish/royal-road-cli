package ui

import (
	"fmt"
	"strconv"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"royal-road-cli/internal/api"
)

type TOCModel struct {
	fiction       *api.Fiction
	currentIndex  int // Currently selected chapter in reader
	selectedIndex int // Selected chapter in TOC (for navigation)
	scrollOffset  int // Current scroll position
	viewHeight    int // Height of the TOC viewport
	visible       bool
	width         int
}

func NewTOCModel(fiction *api.Fiction, currentIndex int, viewHeight int) *TOCModel {
	width, _ := getTerminalSize()
	return &TOCModel{
		fiction:       fiction,
		currentIndex:  currentIndex,
		selectedIndex: currentIndex,
		scrollOffset:  0,
		viewHeight:    max(viewHeight-6, 10),
		visible:       false,
		width:         width,
	}
}

func (m *TOCModel) SetVisible(visible bool) {
	m.visible = visible
	if visible && m.fiction != nil {
		m.centerOnCurrentChapter()
	}
}

func (m *TOCModel) SetCurrentChapter(index int) {
	m.currentIndex = index
	m.selectedIndex = index
	if m.visible {
		m.centerOnCurrentChapter()
	}
}

func (m *TOCModel) centerOnCurrentChapter() {
	if m.fiction == nil || len(m.fiction.Chapters) == 0 {
		return
	}

	idealOffset := m.currentIndex - m.viewHeight/2
	m.scrollOffset = max(0, min(idealOffset, len(m.fiction.Chapters)-m.viewHeight))
}

func (m *TOCModel) Update(msg tea.Msg) (int, bool) {
	if !m.visible || m.fiction == nil {
		return -1, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if m.selectedIndex > 0 {
				m.selectedIndex--
				m.ensureVisible()
			}
			return -1, false
		case "down", "j":
			if m.selectedIndex < len(m.fiction.Chapters)-1 {
				m.selectedIndex++
				m.ensureVisible()
			}
			return -1, false
		case "g", "home":
			m.selectedIndex = 0
			m.scrollOffset = 0
			return -1, false
		case "G", "end":
			m.selectedIndex = len(m.fiction.Chapters) - 1
			m.scrollOffset = max(0, len(m.fiction.Chapters)-m.viewHeight)
			return -1, false
		case "enter":
			return m.selectedIndex, true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if chapterNum, err := strconv.Atoi(msg.String()); err == nil {
				if chapterNum >= 1 && chapterNum <= len(m.fiction.Chapters) {
					return chapterNum - 1, true
				}
			}
			return -1, false
		case "t", "escape":
			return -1, true
		}
	}

	return -1, false
}

func (m *TOCModel) ensureVisible() {
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	} else if m.selectedIndex >= m.scrollOffset+m.viewHeight {
		m.scrollOffset = m.selectedIndex - m.viewHeight + 1
	}

	m.scrollOffset = max(0, min(m.scrollOffset, len(m.fiction.Chapters)-m.viewHeight))
}

func (m *TOCModel) View() string {
	if !m.visible || m.fiction == nil || len(m.fiction.Chapters) == 0 {
		return ""
	}

	s := NewStyles()
	var content strings.Builder

	// Header
	content.WriteString(s.Title.Render("Table of Contents"))
	content.WriteString("\n")

	// Show scroll position
	if len(m.fiction.Chapters) > m.viewHeight {
		scrollInfo := fmt.Sprintf("Showing %d-%d of %d chapters",
			m.scrollOffset+1,
			min(m.scrollOffset+m.viewHeight, len(m.fiction.Chapters)),
			len(m.fiction.Chapters))
		content.WriteString(s.TextMuted.Render(scrollInfo))
		content.WriteString("\n")
	}

	content.WriteString("\n")

	// Calculate visible range
	start := m.scrollOffset
	end := min(start+m.viewHeight, len(m.fiction.Chapters))

	// Chapter list
	for i := start; i < end; i++ {
		chapter := m.fiction.Chapters[i]

		// Determine styling
		var prefix string
		var style lipgloss.Style

		isCurrent := i == m.currentIndex
		isSelected := i == m.selectedIndex

		if isCurrent && isSelected {
			// Current and selected
			prefix = "▶ "
			style = lipgloss.NewStyle().
				Foreground(CurrentTheme.Primary).
				Background(CurrentTheme.Surface).
				Bold(true)
		} else if isCurrent {
			// Current chapter (reading)
			prefix = "▶ "
			style = lipgloss.NewStyle().
				Foreground(CurrentTheme.Primary).
				Bold(true)
		} else if isSelected {
			// Selected for navigation
			prefix = "● "
			style = lipgloss.NewStyle().
				Foreground(CurrentTheme.Text).
				Background(CurrentTheme.Surface)
		} else {
			prefix = "  "
			style = s.Text.Copy()
		}

		// Format chapter number
		number := fmt.Sprintf("%3d", i+1)

		// Truncate title if needed
		maxTitleLen := m.width - 10
		title := chapter.Title
		if len(title) > maxTitleLen {
			title = title[:maxTitleLen-3] + "..."
		}

		line := fmt.Sprintf("%s%s. %s", prefix, number, title)

		// Apply full-width background for selected items
		if isSelected || isCurrent {
			lineWidth := lipgloss.Width(line)
			padding := m.width - lineWidth - 4
			if padding > 0 {
				line = line + strings.Repeat(" ", padding)
			}
		}

		content.WriteString(style.Render(line))
		content.WriteString("\n")
	}

	// Scroll indicators
	if len(m.fiction.Chapters) > m.viewHeight {
		content.WriteString("\n")
		var hints []string
		if m.scrollOffset > 0 {
			hints = append(hints, "↑ more above")
		}
		if end < len(m.fiction.Chapters) {
			hints = append(hints, "↓ more below")
		}
		if len(hints) > 0 {
			content.WriteString(s.TextMuted.Render(strings.Join(hints, " • ")))
		}
	}

	return lipgloss.NewStyle().
		Width(m.width - 4).
		Padding(1, 2).
		Render(content.String())
}

func (m *TOCModel) FooterView() string {
	if !m.visible {
		return ""
	}

	return Footer([]KeyBinding{
		{Key: "↑↓", Desc: "navigate"},
		{Key: "enter", Desc: "jump to chapter"},
		{Key: "1-9", Desc: "quick jump"},
		{Key: "t", Desc: "close"},
	}, m.width)
}
