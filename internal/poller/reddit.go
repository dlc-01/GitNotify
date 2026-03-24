package poller

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sourceReddit = "reddit"

type redditFeed struct {
	Entries []struct {
		Title string `xml:"title"`
		Link  struct {
			Href string `xml:"href,attr"`
		} `xml:"link"`
		Updated string `xml:"updated"`
	} `xml:"entry"`
}

type RedditPoller struct {
	client *http.Client
}

func NewRedditPoller() *RedditPoller {
	return &RedditPoller{
		client: &http.Client{Timeout: 10 * time.Second},
	}
}

func (p *RedditPoller) Source() string { return sourceReddit }

func (p *RedditPoller) Supports(url string) bool {
	return strings.HasPrefix(url, "https://reddit.com/") ||
		strings.HasPrefix(url, "https://www.reddit.com/")
}

func (p *RedditPoller) Poll(ctx context.Context, url string, since time.Time) ([]Event, error) {
	feedURL := buildRedditFeedURL(url)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, wrap("Poll", sourceReddit, url, ErrFetch)
	}
	req.Header.Set("User-Agent", "GitNotify/1.0")

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, wrap("Poll", sourceReddit, url, ErrFetch)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, wrap("Poll", sourceReddit, url, ErrFetch)
	}

	var feed redditFeed
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, wrap("Poll", sourceReddit, url, ErrParse)
	}

	events := make([]Event, 0, len(feed.Entries))
	for _, entry := range feed.Entries {
		updated, err := time.Parse(time.RFC3339, entry.Updated)
		if err != nil || !updated.After(since) {
			continue
		}
		events = append(events, Event{
			URL:       url,
			Source:    sourceReddit,
			EventType: "post",
			Title:     entry.Title,
			Link:      entry.Link.Href,
		})
	}
	return events, nil
}

func buildRedditFeedURL(url string) string {
	url = strings.TrimSuffix(url, "/")
	if !strings.HasSuffix(url, ".rss") {
		url = fmt.Sprintf("%s/.rss", url)
	}
	return url
}
