package api

import (
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

const baseURL = "https://www.royalroad.com"

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) get(path string) (*goquery.Document, error) {
	resp, err := c.httpClient.Get(baseURL + path)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	return doc, nil
}

func (c *Client) GetFiction(id int) (*Fiction, error) {
	path := fmt.Sprintf("/fiction/%d", id)
	doc, err := c.get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get fiction page: %w", err)
	}

	return c.parseFiction(doc, id)
}

func (c *Client) GetChapter(chapterID int) (*Chapter, error) {
	path := fmt.Sprintf("/fiction/0/_/chapter/%d/_", chapterID)
	doc, err := c.get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get chapter page: %w", err)
	}

	return c.parseChapter(doc)
}

func (c *Client) GetPopularFictions() ([]PopularFiction, error) {
	path := "/fictions/best-rated"
	doc, err := c.get(path)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular fictions: %w", err)
	}

	return c.parsePopularFictions(doc)
}

func (c *Client) parseFiction(doc *goquery.Document, id int) (*Fiction, error) {
	fiction := &Fiction{ID: id}

	fiction.Title = doc.Find("div.fic-title h1").Text()
	fiction.Image, _ = doc.Find("div.fic-header img").Attr("src")

	labels := doc.Find("span.bg-blue-hoki")
	if labels.Length() >= 2 {
		fiction.Type = labels.Eq(0).Text()
		fiction.Status = strings.TrimSpace(labels.Eq(1).Text())
	}

	doc.Find("span.tags a.label").Each(func(i int, s *goquery.Selection) {
		fiction.Tags = append(fiction.Tags, strings.TrimSpace(s.Text()))
	})

	doc.Find("ul.list-inline li").Each(func(i int, s *goquery.Selection) {
		warning := strings.TrimSpace(s.Text())
		if warning != "" {
			fiction.Warnings = append(fiction.Warnings, warning)
		}
	})

	fiction.Description = strings.TrimSpace(doc.Find("div.description > div.hidden-content").Text())

	author := doc.Find(".portlet-body").Eq(0)
	fiction.Author.Name = strings.TrimSpace(author.Find(".mt-card-content a").Text())
	fiction.Author.Title = author.Find(".mt-card-desc").Text()
	fiction.Author.Avatar, _ = author.Find("img[data-type=\"avatar\"]").Attr("src")
	
	if href, exists := author.Find(".mt-card-content a").Attr("href"); exists {
		parts := strings.Split(href, "/")
		if len(parts) > 2 {
			if authorID, err := strconv.Atoi(parts[2]); err == nil {
				fiction.Author.ID = authorID
			}
		}
	}

	c.parseStats(doc, &fiction.Stats)
	c.parseChapters(doc, &fiction.Chapters)

	return fiction, nil
}

func (c *Client) parseStats(doc *goquery.Document, stats *FictionStats) {
	parseNumber := func(raw string) int {
		cleaned := regexp.MustCompile(`[,\s]`).ReplaceAllString(raw, "")
		if num, err := strconv.Atoi(cleaned); err == nil {
			return num
		}
		return 0
	}

	parseRating := func(raw string) float64 {
		parts := strings.Split(raw, "/")
		if len(parts) > 0 {
			if rating, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64); err == nil {
				return rating
			}
		}
		return -1
	}

	statsEl := doc.Find("div.stats-content")
	statsList := statsEl.Find(".list-unstyled").Eq(1).Find("li")
	ratingList := statsEl.Find(".list-unstyled").Eq(0).Find("li")

	if statsList.Length() >= 12 {
		stats.Pages = parseNumber(statsList.Eq(11).Text())
		stats.Ratings = parseNumber(statsList.Eq(9).Text())
		stats.Followers = parseNumber(statsList.Eq(5).Text())
		stats.Favorites = parseNumber(statsList.Eq(7).Text())
		stats.Views.Total = parseNumber(statsList.Eq(1).Text())
		stats.Views.Average = parseNumber(statsList.Eq(3).Text())
	}

	if ratingList.Length() >= 10 {
		getContent := func(sel *goquery.Selection) string {
			if content, exists := sel.Find("span").Attr("data-content"); exists {
				return content
			}
			return ""
		}

		stats.Score.Overall = parseRating(getContent(ratingList.Eq(1)))
		stats.Score.Style = parseRating(getContent(ratingList.Eq(3)))
		stats.Score.Story = parseRating(getContent(ratingList.Eq(5)))
		stats.Score.Character = parseRating(getContent(ratingList.Eq(7)))
		stats.Score.Grammar = parseRating(getContent(ratingList.Eq(9)))
	}
}

func (c *Client) parseChapters(doc *goquery.Document, chapters *[]FictionChapter) {
	doc.Find("tbody tr").Each(func(i int, s *goquery.Selection) {
		chapter := FictionChapter{}
		
		titleLink := s.Find("td").Eq(0).Find("a")
		chapter.Title = strings.TrimSpace(titleLink.Text())
		
		if href, exists := titleLink.Attr("href"); exists {
			parts := strings.Split(href, "/")
			if len(parts) > 5 {
				if chapterID, err := strconv.Atoi(parts[5]); err == nil {
					chapter.ID = chapterID
				}
			}
		}

		timeText := s.Find("td").Eq(1).Find("time").Text()
		if releaseTime, err := parseRelativeTime(timeText); err == nil {
			chapter.Release = releaseTime
		}

		*chapters = append(*chapters, chapter)
	})
}

func (c *Client) parseChapter(doc *goquery.Document) (*Chapter, error) {
	chapter := &Chapter{}

	notes := doc.Find("div.author-note")
	if notes.Length() > 0 {
		chapter.PreNote = strings.TrimSpace(notes.Eq(0).Find("p").Text())
	}
	if notes.Length() > 1 {
		chapter.PostNote = strings.TrimSpace(notes.Eq(1).Find("p").Text())
	}

	content, err := doc.Find("div.chapter-inner.chapter-content").Html()
	if err == nil {
		chapter.Content = strings.TrimSpace(content)
	}

	if nextHref, exists := doc.Find("i.fa-chevron-double-right").Parent().Attr("href"); exists {
		chapter.Next = c.extractChapterID(nextHref)
	} else {
		chapter.Next = -1
	}

	if prevHref, exists := doc.Find("i.fa-chevron-double-left").Parent().Attr("href"); exists {
		chapter.Previous = c.extractChapterID(prevHref)
	} else {
		chapter.Previous = -1
	}

	return chapter, nil
}

func (c *Client) parsePopularFictions(doc *goquery.Document) ([]PopularFiction, error) {
	var fictions []PopularFiction

	doc.Find("div.fiction-list-item").Each(func(i int, s *goquery.Selection) {
		fiction := PopularFiction{}

		titleLink := s.Find("h2.fiction-title a")
		fiction.Title = strings.TrimSpace(titleLink.Text())
		
		if href, exists := titleLink.Attr("href"); exists {
			parts := strings.Split(href, "/")
			if len(parts) > 2 {
				if id, err := strconv.Atoi(parts[2]); err == nil {
					fiction.ID = id
				}
			}
		}

		if img, exists := s.Find("img").Attr("src"); exists {
			fiction.Image = img
		}

		fiction.Author = strings.TrimSpace(s.Find(".author").Text())

		s.Find(".tags .label").Each(func(j int, tag *goquery.Selection) {
			fiction.Tags = append(fiction.Tags, strings.TrimSpace(tag.Text()))
		})

		fictions = append(fictions, fiction)
	})

	return fictions, nil
}

func (c *Client) extractChapterID(url string) int {
	parts := strings.Split(url, "/")
	if len(parts) >= 6 {
		if id, err := strconv.Atoi(parts[5]); err == nil {
			return id
		}
	}
	return -1
}

func parseRelativeTime(timeText string) (time.Time, error) {
	now := time.Now()
	
	if strings.Contains(timeText, "ago") {
		if strings.Contains(timeText, "hour") {
			re := regexp.MustCompile(`(\d+)\s*hour`)
			if matches := re.FindStringSubmatch(timeText); len(matches) > 1 {
				if hours, err := strconv.Atoi(matches[1]); err == nil {
					return now.Add(-time.Duration(hours) * time.Hour), nil
				}
			}
		} else if strings.Contains(timeText, "day") {
			re := regexp.MustCompile(`(\d+)\s*day`)
			if matches := re.FindStringSubmatch(timeText); len(matches) > 1 {
				if days, err := strconv.Atoi(matches[1]); err == nil {
					return now.AddDate(0, 0, -days), nil
				}
			}
		} else if strings.Contains(timeText, "week") {
			re := regexp.MustCompile(`(\d+)\s*week`)
			if matches := re.FindStringSubmatch(timeText); len(matches) > 1 {
				if weeks, err := strconv.Atoi(matches[1]); err == nil {
					return now.AddDate(0, 0, -weeks*7), nil
				}
			}
		}
	}

	return now, nil
}