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
	currentIndex  int           // Currently selected chapter in reader
	selectedIndex int           // Selected chapter in TOC (for navigation)
	scrollOffset  int           // Current scroll position
	viewHeight    int           // Height of the TOC viewport
	visible       bool          // Whether TOC is currently visible
}

func NewTOCModel(fiction *api.Fiction, currentIndex int, viewHeight int) *TOCModel {
	return &TOCModel{
		fiction:       fiction,
		currentIndex:  currentIndex,
		selectedIndex: currentIndex,
		scrollOffset:  0,
		viewHeight:    max(viewHeight-4, 10), // Account for header and footer
		visible:       false,
	}
}

func (m *TOCModel) SetVisible(visible bool) {
	m.visible = visible
	if visible && m.fiction != nil {
		// Center the current chapter when TOC becomes visible
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
	
	// Center the current chapter in the viewport
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
			// Jump to selected chapter
			return m.selectedIndex, true
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if chapterNum, err := strconv.Atoi(msg.String()); err == nil {
				if chapterNum >= 1 && chapterNum <= len(m.fiction.Chapters) {
					return chapterNum - 1, true
				}
			}
			return -1, false
		case "t", "escape":
			// Close TOC
			return -1, true
		}
	}
	
	return -1, false
}

func (m *TOCModel) ensureVisible() {
	// Ensure selected item is visible in viewport
	if m.selectedIndex < m.scrollOffset {
		m.scrollOffset = m.selectedIndex
	} else if m.selectedIndex >= m.scrollOffset+m.viewHeight {
		m.scrollOffset = m.selectedIndex - m.viewHeight + 1
	}
	
	// Ensure scroll offset is within bounds
	m.scrollOffset = max(0, min(m.scrollOffset, len(m.fiction.Chapters)-m.viewHeight))
}

func (m *TOCModel) View() string {
	if !m.visible || m.fiction == nil || len(m.fiction.Chapters) == 0 {
		return ""
	}

	var content strings.Builder
	
	// Header
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170")).
		Padding(0, 1)
	content.WriteString(headerStyle.Render("📑 Table of Contents"))
	content.WriteString("\n\n")
	
	// Calculate visible range
	start := m.scrollOffset
	end := min(start+m.viewHeight, len(m.fiction.Chapters))
	
	// Show scroll indicator if needed
	if len(m.fiction.Chapters) > m.viewHeight {
		scrollInfo := fmt.Sprintf("(%d-%d of %d chapters)", 
			start+1, end, len(m.fiction.Chapters))
		infoStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color("240")).
			Italic(true)
		content.WriteString(infoStyle.Render(scrollInfo))
		content.WriteString("\n")
	}
	
	// Chapter list
	for i := start; i < end; i++ {
		chapter := m.fiction.Chapters[i]
		
		// Determine prefix and styling
		var prefix string
		var style lipgloss.Style
		
		if i == m.currentIndex && i == m.selectedIndex {
			// Current and selected
			prefix = "▶ "
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Background(lipgloss.Color("235")).
				Bold(true)
		} else if i == m.currentIndex {
			// Current chapter
			prefix = "▶ "
			style = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true)
		} else if i == m.selectedIndex {
			// Selected for navigation
			prefix = "● "
			style = lipgloss.NewStyle().
				Background(lipgloss.Color("235"))
		} else {
			prefix = "  "
			style = lipgloss.NewStyle()
		}
		
		// Format chapter number
		number := fmt.Sprintf("%2d", i+1)
		if i < 9 {
			number = fmt.Sprintf(" %d", i+1)
		}
		
		line := fmt.Sprintf("%s%s. %s", prefix, number, chapter.Title)
		content.WriteString(style.Render(line))
		content.WriteString("\n")
	}
	
	// Show scroll indicators
	if len(m.fiction.Chapters) > m.viewHeight {
		content.WriteString("\n")
		hints := []string{}
		if m.scrollOffset > 0 {
			hints = append(hints, "↑ more above")
		}
		if end < len(m.fiction.Chapters) {
			hints = append(hints, "↓ more below")
		}
		if len(hints) > 0 {
			hintStyle := lipgloss.NewStyle().
				Foreground(lipgloss.Color("240")).
				Italic(true)
			content.WriteString(hintStyle.Render(strings.Join(hints, " • ")))
		}
	}
	
	return content.String()
}

func (m *TOCModel) FooterView() string {
	if !m.visible {
		return ""
	}
	
	infoStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	return infoStyle.Render("TOC: ↑↓/jk navigate • Enter jump to chapter • 1-9 quick jump • t/Esc close")
}

