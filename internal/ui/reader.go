package ui

import (
	"fmt"
	"html"
	"regexp"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"royal-road-cli/internal/api"
	"royal-road-cli/internal/config"
)

type ReaderModel struct {
	fictionID       string
	client          *api.Client
	fiction         *api.Fiction
	currentChapter  *api.Chapter
	chapterIndex    int
	startChapter    int
	loading         bool
	err             error
	showHelp        bool
	showTOC         bool
	ready           bool
	config          *config.Config
	
	// Page-based navigation
	content              []string  // All content lines
	currentPage          int       // Current page number (0-based)
	linesPerPage         int       // Lines per page
	totalPages           int       // Total number of pages
	termWidth            int       // Terminal width
	termHeight           int       // Terminal height
	goToLastPage         bool      // Flag to go to last page after loading
	savedChapterProgress float64   // Saved progress percentage to restore
}

type fictionLoadedMsg *api.Fiction
type chapterLoadedMsg struct {
	chapter *api.Chapter
	index   int
}

func NewReaderModel(fictionID string) *ReaderModel {
	// Get actual terminal dimensions
	termWidth, termHeight := getTerminalSize()
	
	// Calculate content area (minus header and footer)
	headerHeight := 4
	footerHeight := 1
	linesPerPage := max(termHeight-headerHeight-footerHeight, 10)

	cfg, _ := config.Load()

	return &ReaderModel{
		fictionID:     fictionID,
		client:        api.NewClient(),
		loading:       true,
		showHelp:      false,
		showTOC:       false,
		ready:         true,
		startChapter:  0, // Default to first chapter
		config:        cfg,
		termWidth:     termWidth,
		termHeight:    termHeight,
		linesPerPage:  linesPerPage,
		currentPage:   0,
		content:       []string{},
	}
}

func (m *ReaderModel) SetStartChapter(chapterIndex int) {
	m.startChapter = chapterIndex
}

func (m *ReaderModel) restoreReadingPosition() {
	if m.config == nil {
		return
	}

	// Find the saved progress for this fiction
	for _, entry := range m.config.ReadingHistory {
		if entry.FictionID == m.fictionID {
			// Only restore chapter if it wasn't explicitly set
			if m.startChapter == 0 {
				m.startChapter = entry.CurrentChapter
			}
			
			// Always restore page position if we're on the same chapter
			if m.startChapter == entry.CurrentChapter && entry.ChapterProgress > 0 {
				m.savedChapterProgress = entry.ChapterProgress
			}
			break
		}
	}
}

func (m *ReaderModel) Init() tea.Cmd {
	// Always try to restore reading position from history
	m.restoreReadingPosition()
	
	return tea.Batch(
		m.loadFiction(),
	)
}

func (m *ReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 4
		footerHeight := 1
		
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.linesPerPage = max(msg.Height-headerHeight-footerHeight, 10)
		m.ready = true
		
		// Recalculate pages when window size changes
		if m.currentChapter != nil {
			m.updateContent()
		}

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Save progress before quitting
			m.saveReadingProgress()
			return m, tea.Quit
		case "m":
			// Save progress before going back to menu
			m.saveReadingProgress()
			menuModel := NewMenuModel()
			return menuModel, menuModel.Init()
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "t":
			if m.fiction != nil {
				m.showTOC = !m.showTOC
			}
			return m, nil
		case "n", "b":
			// Next chapter
			if m.fiction != nil && m.chapterIndex < len(m.fiction.Chapters)-1 {
				m.chapterIndex++
				m.loading = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case "p":
			// Previous chapter - go to last page
			if m.fiction != nil && m.chapterIndex > 0 {
				m.chapterIndex--
				m.loading = true
				m.goToLastPage = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case " ", "f", "down", "j", "right", "l":
			// Next page
			if m.currentPage < m.totalPages-1 {
				m.currentPage++
			} else if m.fiction != nil && m.chapterIndex < len(m.fiction.Chapters)-1 {
				// Auto-navigate to next chapter at end of current chapter
				m.chapterIndex++
				m.loading = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case "up", "k", "left", "h":
			// Previous page
			if m.currentPage > 0 {
				m.currentPage--
			} else if m.fiction != nil && m.chapterIndex > 0 {
				// Go to previous chapter and show its last page
				m.chapterIndex--
				m.loading = true
				m.goToLastPage = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case "g", "home":
			// Go to first page
			m.currentPage = 0
			return m, nil
		case "G", "end":
			// Go to last page
			if m.totalPages > 0 {
				m.currentPage = m.totalPages - 1
			}
			return m, nil
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadFiction()
		case "0", "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if m.showTOC && m.fiction != nil {
				if chapterNum, err := strconv.Atoi(msg.String()); err == nil {
					if chapterNum >= 1 && chapterNum <= len(m.fiction.Chapters) {
						m.chapterIndex = chapterNum - 1
						m.loading = true
						m.showTOC = false
						return m, m.loadChapter(m.chapterIndex)
					}
				}
			}
		}

	case fictionLoadedMsg:
		m.loading = false
		m.fiction = msg
		if len(m.fiction.Chapters) > 0 {
			// Start from specified chapter or first chapter
			startIndex := m.startChapter
			if startIndex >= len(m.fiction.Chapters) {
				startIndex = len(m.fiction.Chapters) - 1
			}
			if startIndex < 0 {
				startIndex = 0
			}
			return m, m.loadChapter(startIndex)
		} else {
			m.err = fmt.Errorf("no chapters found")
		}
		return m, nil

	case chapterLoadedMsg:
		m.loading = false
		m.currentChapter = msg.chapter
		m.chapterIndex = msg.index
		m.updateContent()
		
		// Set page position
		if m.goToLastPage {
			// Go to last page
			if m.totalPages > 0 {
				m.currentPage = m.totalPages - 1
			}
			m.goToLastPage = false
		} else if m.savedChapterProgress > 0 {
			// Restore from saved progress percentage
			if m.totalPages > 0 {
				targetPage := int(float64(m.totalPages) * m.savedChapterProgress)
				if targetPage >= m.totalPages {
					targetPage = m.totalPages - 1
				}
				m.currentPage = targetPage
			}
			m.savedChapterProgress = 0 // Clear after using
		} else {
			// Go to first page
			m.currentPage = 0
		}
		
		// Save reading progress
		m.saveReadingProgress()
		
		return m, nil

	case errorMsg:
		m.loading = false
		m.err = msg
		return m, nil
		
	}

	return m, nil
}

func (m *ReaderModel) View() string {
	if !m.ready {
		return "\n  Initializing interface..."
	}

	if m.loading {
		return lipgloss.NewStyle().
			Padding(2).
			Render("🔄 Loading fiction data...")
	}

	if m.err != nil {
		return lipgloss.NewStyle().
			Padding(2).
			Foreground(lipgloss.Color("196")).
			Render(fmt.Sprintf("❌ Error: %v\n\nPress 'r' to retry, 'm' to go back to menu, or 'q' to quit.", m.err))
	}

	header := m.headerView()
	content := m.contentView()
	footer := m.footerView()

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
}

func (m *ReaderModel) headerView() string {
	if m.fiction == nil {
		return ""
	}

	title := m.fiction.Title
	author := m.fiction.Author.Name
	
	var chapterInfo string
	if m.currentChapter != nil && len(m.fiction.Chapters) > 0 {
		chapterInfo = fmt.Sprintf("Chapter %d/%d: %s", 
			m.chapterIndex+1, 
			len(m.fiction.Chapters),
			m.fiction.Chapters[m.chapterIndex].Title)
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("170"))
	
	authorStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(lipgloss.Color("240"))

	chapterStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color("150"))

	return fmt.Sprintf("%s\n%s\n%s", 
		titleStyle.Render(title),
		authorStyle.Render("by "+author),
		chapterStyle.Render(chapterInfo))
}

func (m *ReaderModel) contentView() string {
	if m.showHelp {
		return m.helpContent()
	}
	
	if m.showTOC && m.fiction != nil {
		return m.tocContent()
	}
	
	return m.getCurrentPageContent()
}

func (m *ReaderModel) getCurrentPageContent() string {
	if len(m.content) == 0 {
		if m.currentChapter == nil {
			return "Loading chapter content..."
		}
		return fmt.Sprintf("No content available (content length: 0, chapter loaded: yes, savedProgress: %.3f)", m.savedChapterProgress)
	}
	
	start := m.currentPage * m.linesPerPage
	end := start + m.linesPerPage
	
	if start >= len(m.content) {
		return fmt.Sprintf("End of chapter (page %d, total pages %d, content lines %d)", 
			m.currentPage+1, m.totalPages, len(m.content))
	}
	
	if end > len(m.content) {
		end = len(m.content)
	}
	
	pageContent := make([]string, m.linesPerPage)
	copy(pageContent, m.content[start:end])
	
	// Fill remaining lines with empty strings if needed
	for i := end - start; i < m.linesPerPage; i++ {
		pageContent[i] = ""
	}
	
	return strings.Join(pageContent, "\n")
}

func (m *ReaderModel) footerView() string {
	info := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	
	if m.showHelp {
		return info.Render("Keys: →/← turn pages • n/b next/prev chapter • m menu • q quit")
	}
	
	if m.showTOC {
		return info.Render("TOC mode: Press chapter number (1-9) to jump • t to exit TOC mode")
	}
	
	// Show page progress
	if m.totalPages > 0 {
		progress := fmt.Sprintf("Page %d/%d", m.currentPage+1, m.totalPages)
		
		// Add navigation hints based on position
		if m.currentPage == m.totalPages-1 {
			// On last page
			if m.chapterIndex < len(m.fiction.Chapters)-1 {
				progress += " • [→] next chapter"
			} else {
				progress += " • [end of book]"
			}
		} else {
			progress += " • [→] next page"
		}
		
		if m.currentPage == 0 {
			// On first page
			if m.chapterIndex > 0 {
				progress += " • [←] prev chapter"
			}
		} else {
			progress += " • [←] prev page"
		}
		
		return info.Render(progress)
	}
	
	return info.Render("Press ? for help")
}

func (m *ReaderModel) updateContent() {
	if m.currentChapter == nil {
		return
	}

	// Format and wrap chapter content
	formattedContent := m.formatChapterContent()
	
	// Split into lines for paging
	m.content = strings.Split(formattedContent, "\n")
	
	// Calculate total pages
	if len(m.content) == 0 {
		m.totalPages = 1
	} else {
		m.totalPages = (len(m.content) + m.linesPerPage - 1) / m.linesPerPage
	}
	
	// Ensure current page is valid
	if m.currentPage >= m.totalPages {
		m.currentPage = max(0, m.totalPages-1)
	}
}

func (m *ReaderModel) helpContent() string {
	help := `📖 Royal Road CLI Reader Help

PAGE NAVIGATION:
  → / l / space  Next page (auto-continues to next chapter)
  ← / h          Previous page (auto-goes to prev chapter's end)
  ↑ / k          Previous page
  ↓ / j          Next page
  
CHAPTER NAVIGATION:
  n / b          Next chapter
  p              Previous chapter
  g / home       First page of chapter
  G / end        Last page of chapter
  
FEATURES:
  t              Toggle table of contents
  ?              Toggle this help
  m              Back to main menu
  r              Refresh current content
  q              Quit
  
TABLE OF CONTENTS:
  When TOC is open, press a number (1-9) to jump to that chapter.
  
READING:
  Navigate like reading a book! Use left/right arrows or space to turn pages.
  When you reach the end of a chapter, it automatically continues to the next.
  
Press ? again to close this help.`

	return help
}

func (m *ReaderModel) tocContent() string {
	if m.fiction == nil || len(m.fiction.Chapters) == 0 {
		return "No chapters available"
	}

	var toc strings.Builder
	toc.WriteString("📑 Table of Contents\n\n")
	
	for i, chapter := range m.fiction.Chapters {
		prefix := "  "
		if i == m.chapterIndex {
			prefix = "▶ "
		}
		
		number := fmt.Sprintf("%2d", i+1)
		if i < 9 {
			number = fmt.Sprintf(" %d", i+1)
		}
		
		line := fmt.Sprintf("%s%s. %s", prefix, number, chapter.Title)
		if i == m.chapterIndex {
			line = lipgloss.NewStyle().
				Foreground(lipgloss.Color("170")).
				Bold(true).
				Render(line)
		}
		
		toc.WriteString(line + "\n")
	}
	
	toc.WriteString("\nPress a number (1-9) to jump to chapter, or 't' to exit TOC mode.")
	
	return toc.String()
}

func (m *ReaderModel) formatChapterContent() string {
	if m.currentChapter == nil {
		return "No chapter content available"
	}

	var content strings.Builder

	if m.currentChapter.PreNote != "" {
		authorNote := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("240")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 0, 0, 1)
		
		content.WriteString(authorNote.Render("Author's Note: "+m.currentChapter.PreNote))
		content.WriteString("\n\n")
	}

	chapterContent := m.cleanHTML(m.currentChapter.Content)
	// Use terminal width minus padding for text wrapping
	textWidth := max(m.termWidth-4, 40) // 4 = padding on both sides
	chapterContent = m.wrapText(chapterContent, textWidth)
	
	content.WriteString(chapterContent)

	if m.currentChapter.PostNote != "" {
		content.WriteString("\n\n")
		authorNote := lipgloss.NewStyle().
			Italic(true).
			Foreground(lipgloss.Color("240")).
			BorderStyle(lipgloss.NormalBorder()).
			BorderLeft(true).
			BorderForeground(lipgloss.Color("240")).
			Padding(0, 0, 0, 1)
		
		content.WriteString(authorNote.Render("Author's Note: "+m.currentChapter.PostNote))
	}

	return content.String()
}

func (m *ReaderModel) cleanHTML(htmlContent string) string {
	content := html.UnescapeString(htmlContent)
	
	tagRegex := regexp.MustCompile(`<[^>]*>`)
	content = tagRegex.ReplaceAllString(content, "")
	
	content = regexp.MustCompile(`\s+`).ReplaceAllString(content, " ")
	content = strings.TrimSpace(content)
	
	paragraphs := strings.Split(content, ". ")
	if len(paragraphs) > 1 {
		result := make([]string, 0, len(paragraphs))
		for i, p := range paragraphs {
			p = strings.TrimSpace(p)
			if p != "" {
				if i < len(paragraphs)-1 && !strings.HasSuffix(p, ".") {
					p += "."
				}
				result = append(result, p)
			}
		}
		content = strings.Join(result, "\n\n")
	}
	
	return content
}

func (m *ReaderModel) wrapText(text string, width int) string {
	if width <= 20 {
		width = 40 // Minimum readable width
	}
	
	paragraphs := strings.Split(text, "\n\n")
	var wrappedParagraphs []string
	
	for _, paragraph := range paragraphs {
		if strings.TrimSpace(paragraph) == "" {
			continue
		}
		
		words := strings.Fields(paragraph)
		if len(words) == 0 {
			continue
		}
		
		var lines []string
		currentLine := ""
		
		for _, word := range words {
			if len(currentLine)+len(word)+1 <= width {
				if currentLine == "" {
					currentLine = word
				} else {
					currentLine += " " + word
				}
			} else {
				if currentLine != "" {
					lines = append(lines, currentLine)
				}
				currentLine = word
			}
		}
		
		if currentLine != "" {
			lines = append(lines, currentLine)
		}
		
		wrappedParagraphs = append(wrappedParagraphs, strings.Join(lines, "\n"))
	}
	
	return strings.Join(wrappedParagraphs, "\n\n")
}

func (m *ReaderModel) loadFiction() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		fictionID, err := strconv.Atoi(m.fictionID)
		if err != nil {
			return errorMsg(fmt.Errorf("invalid fiction ID: %s", m.fictionID))
		}
		
		fiction, err := m.client.GetFiction(fictionID)
		if err != nil {
			return errorMsg(err)
		}
		
		return fictionLoadedMsg(fiction)
	})
}

func (m *ReaderModel) loadChapter(index int) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		if m.fiction == nil || index < 0 || index >= len(m.fiction.Chapters) {
			return errorMsg(fmt.Errorf("invalid chapter index"))
		}
		
		chapterID := m.fiction.Chapters[index].ID
		chapter, err := m.client.GetChapter(chapterID)
		if err != nil {
			return errorMsg(err)
		}
		
		return chapterLoadedMsg{chapter: chapter, index: index}
	})
}


func (m *ReaderModel) saveReadingProgress() {
	if m.fiction == nil || m.config == nil {
		return
	}

	chapterTitle := ""
	if m.chapterIndex < len(m.fiction.Chapters) {
		chapterTitle = m.fiction.Chapters[m.chapterIndex].Title
	}

	// Calculate progress through current chapter as a percentage
	var chapterProgress float64
	if m.totalPages > 0 {
		chapterProgress = float64(m.currentPage) / float64(m.totalPages)
		// Ensure we don't go over 1.0
		if chapterProgress > 1.0 {
			chapterProgress = 1.0
		}
	}

	entry := config.ReadingEntry{
		FictionID:       m.fictionID,
		FictionTitle:    m.fiction.Title,
		Author:          m.fiction.Author.Name,
		CurrentChapter:  m.chapterIndex,
		ChapterTitle:    chapterTitle,
		ChapterProgress: chapterProgress,
		LastRead:        time.Now().Format("2006-01-02 15:04"),
		TotalChapters:   len(m.fiction.Chapters),
	}

	m.config.UpdateReadingProgress(entry)
	m.config.Save() // Save to disk
}