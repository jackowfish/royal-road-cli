package api

import "time"

type Fiction struct {
	ID          int              `json:"id"`
	Type        string           `json:"type"`
	Title       string           `json:"title"`
	Image       string           `json:"image"`
	Status      string           `json:"status"`
	Tags        []string         `json:"tags"`
	Warnings    []string         `json:"warnings"`
	Description string           `json:"description"`
	Stats       FictionStats     `json:"stats"`
	Author      FictionAuthor    `json:"author"`
	Chapters    []FictionChapter `json:"chapters"`
}

type FictionChapter struct {
	ID      int       `json:"id"`
	Title   string    `json:"title"`
	Release time.Time `json:"release"`
}

type FictionStats struct {
	Pages     int           `json:"pages"`
	Ratings   int           `json:"ratings"`
	Favorites int           `json:"favorites"`
	Followers int           `json:"followers"`
	Views     FictionViews  `json:"views"`
	Score     FictionScore  `json:"score"`
}

type FictionViews struct {
	Total   int `json:"total"`
	Average int `json:"average"`
}

type FictionScore struct {
	Style     float64 `json:"style"`
	Story     float64 `json:"story"`
	Grammar   float64 `json:"grammar"`
	Overall   float64 `json:"overall"`
	Character float64 `json:"character"`
}

type FictionAuthor struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Title  string `json:"title"`
	Avatar string `json:"avatar"`
}

type Chapter struct {
	Content  string `json:"content"`
	PreNote  string `json:"preNote"`
	PostNote string `json:"postNote"`
	Next     int    `json:"next"`
	Previous int    `json:"previous"`
}

type PopularFiction struct {
	ID     int    `json:"id"`
	Title  string `json:"title"`
	Image  string `json:"image"`
	Author string `json:"author"`
	Tags   []string `json:"tags"`
	Stats  struct {
		Pages     int `json:"pages"`
		Followers int `json:"followers"`
		Favorites int `json:"favorites"`
		Score     float64 `json:"score"`
	} `json:"stats"`
}

type SearchFiction struct {
	ID          int               `json:"id"`
	Title       string            `json:"title"`
	Image       string            `json:"image"`
	Author      string            `json:"author"`
	Tags        []string          `json:"tags"`
	Type        string            `json:"type"`
	Status      string            `json:"status"`
	Description string            `json:"description"`
	Stats       SearchFictionStats `json:"stats"`
}

type SearchFictionStats struct {
	Followers int     `json:"followers"`
	Rating    float64 `json:"rating"`
	Pages     int     `json:"pages"`
	Views     int     `json:"views"`
	Chapters  int     `json:"chapters"`
}