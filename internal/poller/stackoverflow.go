package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sourceStackOverflow = "stackoverflow"

type stackOverflowResponse struct {
	Items []struct {
		QuestionID   int    `json:"question_id"`
		Title        string `json:"title"`
		Link         string `json:"link"`
		CreationDate int64  `json:"creation_date"`
	} `json:"items"`
}

type StackOverflowPoller struct {
	client *http.Client
	apiURL string
}

func NewStackOverflowPoller() *StackOverflowPoller {
	return &StackOverflowPoller{
		client: &http.Client{Timeout: 10 * time.Second},
		apiURL: "https://api.stackexchange.com/2.3/questions",
	}
}

func (p *StackOverflowPoller) Source() string { return sourceStackOverflow }

func (p *StackOverflowPoller) Supports(url string) bool {
	return strings.HasPrefix(url, "https://stackoverflow.com/")
}

func (p *StackOverflowPoller) Poll(ctx context.Context, url string, since time.Time) ([]Event, error) {
	tag, err := extractSOTag(url)
	if err != nil {
		return nil, wrap("Poll", sourceStackOverflow, url, err)
	}

	apiURL := fmt.Sprintf(
		"%s?order=desc&sort=creation&tagged=%s&site=stackoverflow&fromdate=%d",
		p.apiURL, tag, since.Unix(),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, wrap("Poll", sourceStackOverflow, url, ErrFetch)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, wrap("Poll", sourceStackOverflow, url, ErrFetch)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, wrap("Poll", sourceStackOverflow, url, ErrFetch)
	}

	var result stackOverflowResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, wrap("Poll", sourceStackOverflow, url, ErrParse)
	}

	events := make([]Event, 0, len(result.Items))
	for _, item := range result.Items {
		if time.Unix(item.CreationDate, 0).Before(since) {
			continue
		}
		events = append(events, Event{
			URL:       url,
			Source:    sourceStackOverflow,
			EventType: "answer",
			Title:     item.Title,
			Link:      item.Link,
		})
	}
	return events, nil
}

func extractSOTag(url string) (string, error) {
	tag := strings.TrimPrefix(url, "https://stackoverflow.com/questions/tagged/")
	if tag == "" || tag == url {
		return "", ErrInvalidURL
	}
	return tag, nil
}
