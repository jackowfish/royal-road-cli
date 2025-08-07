package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	Theme           Theme           `json:"theme"`
	Reading         Reading         `json:"reading"`
	LastFiction     string          `json:"lastFiction"`
	Bookmarks       []Bookmark      `json:"bookmarks"`
	ReadingHistory  []ReadingEntry  `json:"readingHistory"`
}

type Theme struct {
	AccentColor   string `json:"accentColor"`
	BackgroundColor string `json:"backgroundColor"`
	TextColor     string `json:"textColor"`
}

type Reading struct {
	TextWidth     int  `json:"textWidth"`
	ShowProgress  bool `json:"showProgress"`
	WrapText      bool `json:"wrapText"`
}

type Bookmark struct {
	FictionID    string `json:"fictionId"`
	FictionTitle string `json:"fictionTitle"`
	ChapterIndex int    `json:"chapterIndex"`
	ChapterTitle string `json:"chapterTitle"`
	Position     int    `json:"position"`
	CreatedAt    string `json:"createdAt"`
}

type ReadingEntry struct {
	FictionID      string  `json:"fictionId"`
	FictionTitle   string  `json:"fictionTitle"`
	Author         string  `json:"author"`
	CurrentChapter int     `json:"currentChapter"`
	ChapterTitle   string  `json:"chapterTitle"`
	ChapterProgress float64 `json:"chapterProgress"`  // Percentage through chapter (0.0-1.0)
	LastRead       string  `json:"lastRead"`
	TotalChapters  int     `json:"totalChapters"`
}

func DefaultConfig() *Config {
	return &Config{
		Theme: Theme{
			AccentColor:     "170",
			BackgroundColor: "0",
			TextColor:       "15",
		},
		Reading: Reading{
			TextWidth:    78,
			ShowProgress: true,
			WrapText:     true,
		},
		LastFiction:    "",
		Bookmarks:      []Bookmark{},
		ReadingHistory: []ReadingEntry{},
	}
}

func Load() (*Config, error) {
	configPath, err := getConfigPath()
	if err != nil {
		return DefaultConfig(), nil
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		config := DefaultConfig()
		_ = config.Save()
		return config, nil
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return DefaultConfig(), err
	}

	config := DefaultConfig()
	if err := json.Unmarshal(data, config); err != nil {
		return DefaultConfig(), err
	}

	return config, nil
}

func (c *Config) Save() error {
	configPath, err := getConfigPath()
	if err != nil {
		return err
	}

	configDir := filepath.Dir(configPath)
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, data, 0644)
}

func (c *Config) AddBookmark(bookmark Bookmark) {
	for i, existing := range c.Bookmarks {
		if existing.FictionID == bookmark.FictionID && existing.ChapterIndex == bookmark.ChapterIndex {
			c.Bookmarks[i] = bookmark
			return
		}
	}
	c.Bookmarks = append(c.Bookmarks, bookmark)
}

func (c *Config) RemoveBookmark(fictionID string, chapterIndex int) {
	for i, bookmark := range c.Bookmarks {
		if bookmark.FictionID == fictionID && bookmark.ChapterIndex == chapterIndex {
			c.Bookmarks = append(c.Bookmarks[:i], c.Bookmarks[i+1:]...)
			return
		}
	}
}

func (c *Config) UpdateReadingProgress(entry ReadingEntry) {
	// Update existing entry or add new one
	for i, existing := range c.ReadingHistory {
		if existing.FictionID == entry.FictionID {
			// Update existing entry and move to front (most recent)
			c.ReadingHistory[i] = entry
			if i != 0 {
				// Move to front
				c.ReadingHistory = append([]ReadingEntry{entry}, append(c.ReadingHistory[:i], c.ReadingHistory[i+1:]...)...)
			}
			c.LastFiction = entry.FictionID
			return
		}
	}
	
	// Add new entry at the beginning (most recent first)
	c.ReadingHistory = append([]ReadingEntry{entry}, c.ReadingHistory...)
	c.LastFiction = entry.FictionID
}

func (c *Config) GetReadingHistoryPage(page, pageSize int) ([]ReadingEntry, int, bool, bool) {
	total := len(c.ReadingHistory)
	if total == 0 {
		return []ReadingEntry{}, 0, false, false
	}
	
	totalPages := (total + pageSize - 1) / pageSize
	if page < 1 {
		page = 1
	}
	if page > totalPages {
		page = totalPages
	}
	
	start := (page - 1) * pageSize
	end := start + pageSize
	if end > total {
		end = total
	}
	
	hasNext := page < totalPages
	hasPrev := page > 1
	
	return c.ReadingHistory[start:end], totalPages, hasNext, hasPrev
}

func (c *Config) GetLastReadEntry() *ReadingEntry {
	if len(c.ReadingHistory) > 0 {
		return &c.ReadingHistory[0]
	}
	return nil
}

func getConfigPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	
	return filepath.Join(homeDir, ".config", "royal-road-cli", "config.json"), nil
}