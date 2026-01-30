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
	fictionID      string
	client         *api.Client
	fiction        *api.Fiction
	currentChapter *api.Chapter
	chapterIndex   int
	startChapter   int
	loading        bool
	err            error
	showHelp       bool
	showTOC        bool
	ready          bool
	config         *config.Config
	tocModel       *TOCModel

	// Page-based navigation
	content              []string // All content lines
	currentPage          int      // Current page number (0-based)
	linesPerPage         int      // Lines per page
	totalPages           int      // Total number of pages
	termWidth            int      // Terminal width
	termHeight           int      // Terminal height
	goToLastPage         bool     // Flag to go to last page after loading
	savedChapterProgress float64  // Saved progress percentage to restore
}

type fictionLoadedMsg *api.Fiction
type chapterLoadedMsg struct {
	chapter *api.Chapter
	index   int
}

func NewReaderModel(fictionID string) *ReaderModel {
	termWidth, termHeight := getTerminalSize()

	// Calculate content area (minus header, footer, and content border)
	headerHeight := 4
	footerHeight := 2
	borderHeight := 2 // top + bottom border
	linesPerPage := max(termHeight-headerHeight-footerHeight-borderHeight, 10)

	cfg, _ := config.Load()

	return &ReaderModel{
		fictionID:    fictionID,
		client:       api.NewClient(),
		loading:      true,
		showHelp:     false,
		showTOC:      false,
		ready:        true,
		startChapter: 0,
		config:       cfg,
		termWidth:    termWidth,
		termHeight:   termHeight,
		linesPerPage: linesPerPage,
		currentPage:  0,
		content:      []string{},
		tocModel:     nil,
	}
}

func (m *ReaderModel) SetStartChapter(chapterIndex int) {
	m.startChapter = chapterIndex
}

func (m *ReaderModel) restoreReadingPosition() {
	if m.config == nil {
		return
	}

	for _, entry := range m.config.ReadingHistory {
		if entry.FictionID == m.fictionID {
			if m.startChapter == 0 {
				m.startChapter = entry.CurrentChapter
			}

			if m.startChapter == entry.CurrentChapter && entry.ChapterProgress > 0 {
				m.savedChapterProgress = entry.ChapterProgress
			}
			break
		}
	}
}

func (m *ReaderModel) Init() tea.Cmd {
	m.restoreReadingPosition()

	return tea.Batch(
		m.loadFiction(),
	)
}

func (m *ReaderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		headerHeight := 4
		footerHeight := 2
		borderHeight := 2 // top + bottom border of content panel

		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.linesPerPage = max(msg.Height-headerHeight-footerHeight-borderHeight, 10)
		m.ready = true

		if m.currentChapter != nil {
			m.updateContent()
		}

	case tea.KeyMsg:
		// Handle TOC navigation first if TOC is visible
		if m.showTOC && m.tocModel != nil {
			if selectedChapter, shouldClose := m.tocModel.Update(msg); shouldClose {
				m.showTOC = false
				m.tocModel.SetVisible(false)
				if selectedChapter >= 0 {
					m.chapterIndex = selectedChapter
					m.loading = true
					return m, m.loadChapter(selectedChapter)
				}
				return m, nil
			}
			return m, nil
		}

		switch msg.String() {
		case "ctrl+c", "q":
			m.saveReadingProgress()
			return m, tea.Quit
		case "m", "esc":
			m.saveReadingProgress()
			menuModel := NewMenuModel()
			return menuModel, menuModel.Init()
		case "?":
			m.showHelp = !m.showHelp
			return m, nil
		case "t":
			if m.fiction != nil && m.tocModel != nil {
				m.showTOC = !m.showTOC
				m.tocModel.SetVisible(m.showTOC)
			}
			return m, nil
		case "n", "b":
			if m.fiction != nil && m.chapterIndex < len(m.fiction.Chapters)-1 {
				m.chapterIndex++
				m.loading = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case "p":
			if m.fiction != nil && m.chapterIndex > 0 {
				m.chapterIndex--
				m.loading = true
				m.goToLastPage = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case " ", "f", "down", "j", "right", "l":
			if m.currentPage < m.totalPages-1 {
				m.currentPage++
				return m, tea.ClearScreen
			} else if m.fiction != nil && m.chapterIndex < len(m.fiction.Chapters)-1 {
				m.chapterIndex++
				m.loading = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case "up", "k", "left", "h":
			if m.currentPage > 0 {
				m.currentPage--
				return m, tea.ClearScreen
			} else if m.fiction != nil && m.chapterIndex > 0 {
				m.chapterIndex--
				m.loading = true
				m.goToLastPage = true
				return m, m.loadChapter(m.chapterIndex)
			}
			return m, nil
		case "g", "home":
			m.currentPage = 0
			return m, tea.ClearScreen
		case "G", "end":
			if m.totalPages > 0 {
				m.currentPage = m.totalPages - 1
			}
			return m, tea.ClearScreen
		case "r":
			m.loading = true
			m.err = nil
			return m, m.loadFiction()
		}

	case fictionLoadedMsg:
		m.loading = false
		m.fiction = msg

		m.tocModel = NewTOCModel(m.fiction, 0, m.termHeight)

		if len(m.fiction.Chapters) > 0 {
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

		if m.tocModel != nil {
			m.tocModel.SetCurrentChapter(msg.index)
		}

		m.updateContent()

		if m.goToLastPage {
			if m.totalPages > 0 {
				m.currentPage = m.totalPages - 1
			}
			m.goToLastPage = false
		} else if m.savedChapterProgress > 0 {
			if m.totalPages > 0 {
				targetPage := int(float64(m.totalPages) * m.savedChapterProgress)
				if targetPage >= m.totalPages {
					targetPage = m.totalPages - 1
				}
				m.currentPage = targetPage
			}
			m.savedChapterProgress = 0
		} else {
			m.currentPage = 0
		}

		m.saveReadingProgress()

		return m, tea.ClearScreen

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

	s := NewStyles()

	contentHeight := m.termHeight - 6 // Account for header and footer

	if m.loading {
		// Loading state with header/footer
		header := m.headerView()
		loadingContent := "\n\n" + CenterText(LoadingMessage("Loading chapter...", 0), m.termWidth-6)
		content := m.renderContentPanel(loadingContent, m.termWidth, contentHeight)
		footer := m.footerView()
		return header + "\n" + content + "\n" + footer
	}

	if m.err != nil {
		header := m.headerView()
		errContent := "\n\n" + CenterText(ErrorMessage(m.err), m.termWidth-6) + "\n\n" + CenterText(s.TextMuted.Render("Press 'r' to retry, 'm' to go back to menu"), m.termWidth-6)
		content := m.renderContentPanel(errContent, m.termWidth, contentHeight)
		footer := Footer([]KeyBinding{
			{Key: "r", Desc: "retry"},
			{Key: "m", Desc: "menu"},
			{Key: "q", Desc: "quit"},
		}, m.termWidth)
		return header + "\n" + content + "\n" + footer
	}

	header := m.headerView()
	content := m.contentView()
	footer := m.footerView()

	// Wrap content in a bordered panel for a polished look
	borderedContent := m.renderContentPanel(content, m.termWidth, contentHeight)

	return header + "\n" + borderedContent + "\n" + footer
}

func (m *ReaderModel) headerView() string {
	s := NewStyles()

	if m.fiction == nil {
		return Header("Royal Road CLI", "Reader", m.termWidth)
	}

	// Build a rich header
	title := m.fiction.Title
	author := "by " + m.fiction.Author.Name

	var chapterInfo string
	var progress string

	if m.currentChapter != nil && len(m.fiction.Chapters) > 0 {
		chapterInfo = fmt.Sprintf("Ch %d/%d: %s",
			m.chapterIndex+1,
			len(m.fiction.Chapters),
			TruncateWithEllipsis(m.fiction.Chapters[m.chapterIndex].Title, 40))
	}

	if m.totalPages > 0 {
		progress = fmt.Sprintf("Page %d/%d", m.currentPage+1, m.totalPages)
	}

	// First line: Title and progress
	titleStyle := s.Title.Copy()
	progressStyle := s.TextMuted.Copy()

	titleStr := titleStyle.Render(TruncateWithEllipsis(title, m.termWidth-len(progress)-10))
	progressStr := progressStyle.Render(progress)

	titleWidth := lipgloss.Width(titleStr)
	progressWidth := lipgloss.Width(progressStr)
	spacing := m.termWidth - titleWidth - progressWidth - 4
	if spacing < 0 {
		spacing = 0
	}

	line1 := titleStr + strings.Repeat(" ", spacing) + progressStr

	// Second line: Author
	authorStr := s.Subtitle.Render(author)

	// Third line: Chapter info
	chapterStr := s.Text.Render(chapterInfo)

	// Build header with border
	headerContent := line1 + "\n" + authorStr + "\n" + chapterStr

	return lipgloss.NewStyle().
		Width(m.termWidth).
		Padding(0, 1).
		BorderStyle(lipgloss.Border{Bottom: "─"}).
		BorderForeground(CurrentTheme.Border).
		BorderBottom(true).
		Render(headerContent)
}

func (m *ReaderModel) contentView() string {
	if m.showHelp {
		return m.helpContent()
	}

	if m.showTOC && m.tocModel != nil {
		return m.tocModel.View()
	}

	return m.getCurrentPageContent()
}

func (m *ReaderModel) getCurrentPageContent() string {
	// Calculate the width for clearing lines (account for border and padding)
	clearWidth := m.termWidth - 6
	if clearWidth < 1 {
		clearWidth = 1
	}
	clearLine := strings.Repeat(" ", clearWidth)

	if len(m.content) == 0 {
		if m.currentChapter == nil {
			return "Loading chapter content..."
		}
		return "No content available"
	}

	start := m.currentPage * m.linesPerPage
	end := start + m.linesPerPage

	if start >= len(m.content) {
		return "End of chapter"
	}

	if end > len(m.content) {
		end = len(m.content)
	}

	pageContent := make([]string, m.linesPerPage)
	copy(pageContent, m.content[start:end])

	// Fill remaining lines with spaces to clear any previous content
	for i := end - start; i < m.linesPerPage; i++ {
		pageContent[i] = clearLine
	}

	return strings.Join(pageContent, "\n")
}

func (m *ReaderModel) renderContentPanel(content string, width, height int) string {
	innerWidth := width - 4   // Account for border and padding
	innerHeight := height - 2 // Account for top and bottom border

	// Use lipgloss.Place to ensure the entire area is filled (clears previous content)
	placedContent := lipgloss.Place(
		innerWidth,
		innerHeight,
		lipgloss.Left,
		lipgloss.Top,
		content,
		lipgloss.WithWhitespaceChars(" "),
	)

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		BorderStyle(SubtleBorder()).
		BorderForeground(CurrentTheme.Border).
		Padding(0, 1).
		Render(placedContent)
}

func (m *ReaderModel) footerView() string {
	if m.showHelp {
		return Footer([]KeyBinding{
			{Key: "?", Desc: "close help"},
		}, m.termWidth)
	}

	if m.showTOC && m.tocModel != nil {
		return Footer([]KeyBinding{
			{Key: "↑↓", Desc: "navigate"},
			{Key: "enter", Desc: "jump"},
			{Key: "t", Desc: "close"},
		}, m.termWidth)
	}

	// Build footer with progress and keybindings
	var bindings []KeyBinding

	// Navigation based on position
	if m.currentPage < m.totalPages-1 {
		bindings = append(bindings, KeyBinding{Key: "→", Desc: "next page"})
	} else if m.fiction != nil && m.chapterIndex < len(m.fiction.Chapters)-1 {
		bindings = append(bindings, KeyBinding{Key: "→", Desc: "next chapter"})
	}

	if m.currentPage > 0 {
		bindings = append(bindings, KeyBinding{Key: "←", Desc: "prev page"})
	} else if m.fiction != nil && m.chapterIndex > 0 {
		bindings = append(bindings, KeyBinding{Key: "←", Desc: "prev chapter"})
	}

	bindings = append(bindings,
		KeyBinding{Key: "t", Desc: "TOC"},
		KeyBinding{Key: "?", Desc: "help"},
		KeyBinding{Key: "m", Desc: "menu"},
	)

	// Progress indicator
	progress := ""
	if m.totalPages > 0 {
		pct := float64(m.currentPage+1) / float64(m.totalPages) * 100
		progress = fmt.Sprintf("%.0f%%", pct)
	}

	return FooterWithProgress(bindings, progress, m.termWidth)
}

func (m *ReaderModel) updateContent() {
	if m.currentChapter == nil {
		return
	}

	formattedContent := m.formatChapterContent()
	m.content = strings.Split(formattedContent, "\n")

	if len(m.content) == 0 {
		m.totalPages = 1
	} else {
		m.totalPages = (len(m.content) + m.linesPerPage - 1) / m.linesPerPage
	}

	if m.currentPage >= m.totalPages {
		m.currentPage = max(0, m.totalPages-1)
	}
}

func (m *ReaderModel) helpContent() string {
	s := NewStyles()

	help := s.Title.Render("Keyboard Shortcuts") + "\n\n"

	sections := []struct {
		title    string
		bindings []KeyBinding
	}{
		{
			title: "Page Navigation",
			bindings: []KeyBinding{
				{Key: "→ l space", Desc: "Next page (auto-advances to next chapter)"},
				{Key: "← h", Desc: "Previous page (auto-goes to prev chapter)"},
				{Key: "↑ k", Desc: "Previous page"},
				{Key: "↓ j", Desc: "Next page"},
				{Key: "g home", Desc: "First page of chapter"},
				{Key: "G end", Desc: "Last page of chapter"},
			},
		},
		{
			title: "Chapter Navigation",
			bindings: []KeyBinding{
				{Key: "n b", Desc: "Next chapter"},
				{Key: "p", Desc: "Previous chapter"},
				{Key: "t", Desc: "Toggle table of contents"},
			},
		},
		{
			title: "Other",
			bindings: []KeyBinding{
				{Key: "?", Desc: "Toggle this help"},
				{Key: "m esc", Desc: "Back to main menu"},
				{Key: "r", Desc: "Refresh current content"},
				{Key: "q", Desc: "Quit"},
			},
		},
	}

	for _, section := range sections {
		help += s.Subtitle.Render(section.title) + "\n"
		for _, b := range section.bindings {
			key := s.MenuKey.Render(fmt.Sprintf("  %-12s", b.Key))
			desc := s.Text.Render(b.Desc)
			help += key + " " + desc + "\n"
		}
		help += "\n"
	}

	help += s.TextMuted.Render("Press ? to close this help")

	return help
}

func (m *ReaderModel) formatChapterContent() string {
	if m.currentChapter == nil {
		return "No chapter content available"
	}

	var content strings.Builder
	s := NewStyles()

	if m.currentChapter.PreNote != "" {
		noteStyle := s.TextMuted.Copy().
			Italic(true).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(CurrentTheme.Border).
			PaddingLeft(1)

		content.WriteString(noteStyle.Render("Author's Note: "+m.currentChapter.PreNote))
		content.WriteString("\n\n")
	}

	chapterContent := m.cleanHTML(m.currentChapter.Content)
	textWidth := max(m.termWidth-8, 40) // Account for border (2) + padding (4) + margin (2)
	chapterContent = m.wrapText(chapterContent, textWidth)

	content.WriteString(chapterContent)

	if m.currentChapter.PostNote != "" {
		content.WriteString("\n\n")
		noteStyle := s.TextMuted.Copy().
			Italic(true).
			BorderStyle(lipgloss.Border{Left: "│"}).
			BorderForeground(CurrentTheme.Border).
			PaddingLeft(1)

		content.WriteString(noteStyle.Render("Author's Note: "+m.currentChapter.PostNote))
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
		width = 40
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

	var chapterProgress float64
	if m.totalPages > 0 {
		chapterProgress = float64(m.currentPage) / float64(m.totalPages)
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
	if err := m.config.Save(); err != nil {
		// Log error but don't interrupt reading experience
		_ = err // error logged silently; config save is best-effort
	}
}
