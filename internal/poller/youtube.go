package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sourceYouTube = "youtube"

type youtubeResponse struct {
	Items []struct {
		Snippet struct {
			Title       string `json:"title"`
			PublishedAt string `json:"publishedAt"`
			ResourceID  struct {
				VideoID string `json:"videoId"`
			} `json:"resourceId"`
		} `json:"snippet"`
	} `json:"items"`
}

type YouTubePoller struct {
	client *http.Client
	apiKey string
	apiURL string
}

func NewYouTubePoller(apiKey string) *YouTubePoller {
	return &YouTubePoller{
		client: &http.Client{Timeout: 10 * time.Second},
		apiKey: apiKey,
		apiURL: "https://www.googleapis.com/youtube/v3/playlistItems",
	}
}

func (p *YouTubePoller) Source() string { return sourceYouTube }

func (p *YouTubePoller) Supports(url string) bool {
	return strings.HasPrefix(url, "https://youtube.com/") ||
		strings.HasPrefix(url, "https://www.youtube.com/")
}

func (p *YouTubePoller) Poll(ctx context.Context, url string, since time.Time) ([]Event, error) {
	channelID, err := extractYouTubeChannelID(url)
	if err != nil {
		return nil, wrap("Poll", sourceYouTube, url, err)
	}

	apiURL := fmt.Sprintf(
		"%s?part=snippet&maxResults=10&playlistId=%s&key=%s",
		p.apiURL, channelID, p.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, wrap("Poll", sourceYouTube, url, ErrFetch)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, wrap("Poll", sourceYouTube, url, ErrFetch)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, wrap("Poll", sourceYouTube, url, ErrFetch)
	}

	var result youtubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, wrap("Poll", sourceYouTube, url, ErrParse)
	}

	events := make([]Event, 0, len(result.Items))
	for _, item := range result.Items {
		published, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil || !published.After(since) {
			continue
		}
		events = append(events, Event{
			URL:    url,
			Source: sourceYouTube,
			Title:  item.Snippet.Title,
			Link:   fmt.Sprintf("https://youtube.com/watch?v=%s", item.Snippet.ResourceID.VideoID),
		})
	}
	return events, nil
}

func extractYouTubeChannelID(url string) (string, error) {
	cleaned := strings.TrimPrefix(url, "https://www.youtube.com/")
	cleaned = strings.TrimPrefix(cleaned, "https://youtube.com/")
	cleaned = strings.TrimPrefix(cleaned, "@")
	cleaned = strings.TrimSuffix(cleaned, "/")

	if cleaned == "" {
		return "", ErrInvalidURL
	}
	return cleaned, nil
}
