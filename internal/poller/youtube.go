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

type youtubeChannelResponse struct {
	Items []struct {
		ContentDetails struct {
			RelatedPlaylists struct {
				Uploads string `json:"uploads"`
			} `json:"relatedPlaylists"`
		} `json:"contentDetails"`
	} `json:"items"`
}

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
		apiURL: "https://www.googleapis.com/youtube/v3",
	}
}

func (p *YouTubePoller) Source() string { return sourceYouTube }

func (p *YouTubePoller) Supports(url string) bool {
	return strings.HasPrefix(url, "https://youtube.com/") ||
		strings.HasPrefix(url, "https://www.youtube.com/")
}

func (p *YouTubePoller) Poll(ctx context.Context, url string, since time.Time) ([]Event, error) {
	handle, err := extractYouTubeHandle(url)
	if err != nil {
		return nil, wrap("Poll", sourceYouTube, url, err)
	}

	playlistID, err := p.getUploadsPlaylistID(ctx, handle)
	if err != nil {
		return nil, wrap("Poll", sourceYouTube, url, err)
	}

	return p.getVideos(ctx, url, playlistID, since)
}

func (p *YouTubePoller) getUploadsPlaylistID(ctx context.Context, handle string) (string, error) {
	apiURL := fmt.Sprintf(
		"%s/channels?part=contentDetails&forHandle=%s&key=%s",
		p.apiURL, handle, p.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", ErrFetch
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return "", ErrFetch
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", ErrFetch
	}

	var result youtubeChannelResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", ErrParse
	}

	if len(result.Items) == 0 {
		return "", ErrInvalidURL
	}

	playlistID := result.Items[0].ContentDetails.RelatedPlaylists.Uploads
	if playlistID == "" {
		return "", ErrInvalidURL
	}

	return playlistID, nil
}

func (p *YouTubePoller) getVideos(ctx context.Context, url string, playlistID string, since time.Time) ([]Event, error) {
	apiURL := fmt.Sprintf(
		"%s/playlistItems?part=snippet&maxResults=10&playlistId=%s&key=%s",
		p.apiURL, playlistID, p.apiKey,
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, ErrFetch
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, ErrFetch
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, ErrFetch
	}

	var result youtubeResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, ErrParse
	}

	events := make([]Event, 0, len(result.Items))
	for _, item := range result.Items {
		published, err := time.Parse(time.RFC3339, item.Snippet.PublishedAt)
		if err != nil || !published.After(since) {
			continue
		}
		events = append(events, Event{
			URL:       url,
			Source:    sourceYouTube,
			EventType: "video",
			Title:     item.Snippet.Title,
			Link:      fmt.Sprintf("https://youtube.com/watch?v=%s", item.Snippet.ResourceID.VideoID),
		})
	}
	return events, nil
}

func extractYouTubeHandle(url string) (string, error) {
	cleaned := strings.TrimPrefix(url, "https://www.youtube.com/")
	cleaned = strings.TrimPrefix(cleaned, "https://youtube.com/")
	cleaned = strings.TrimPrefix(cleaned, "@")
	cleaned = strings.TrimSuffix(cleaned, "/")

	if cleaned == "" {
		return "", ErrInvalidURL
	}
	return cleaned, nil
}

func extractYouTubeChannelID(url string) (string, error) {
	return extractYouTubeHandle(url)
}
