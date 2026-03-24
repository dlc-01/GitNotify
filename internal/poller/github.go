package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sourceGitHub = "github"

type githubEvent struct {
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
	Repo      struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"repo"`
	Payload json.RawMessage `json:"payload"`
}

type GitHubPoller struct {
	client *http.Client
	token  string
	apiURL string
}

func NewGitHubPoller(token string) *GitHubPoller {
	return &GitHubPoller{
		client: &http.Client{Timeout: 10 * time.Second},
		token:  token,
		apiURL: "https://api.github.com",
	}
}

func (p *GitHubPoller) Source() string { return sourceGitHub }

func (p *GitHubPoller) Supports(url string) bool {
	return strings.HasPrefix(url, "https://github.com/")
}

func (p *GitHubPoller) Poll(ctx context.Context, url string, since time.Time) ([]Event, error) {
	owner, repo, err := extractGitHubOwnerRepo(url)
	if err != nil {
		return nil, wrap("Poll", sourceGitHub, url, err)
	}

	apiURL := fmt.Sprintf("%s/repos/%s/%s/events", p.apiURL, owner, repo)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, wrap("Poll", sourceGitHub, url, ErrFetch)
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	if p.token != "" {
		req.Header.Set("Authorization", "Bearer "+p.token)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, wrap("Poll", sourceGitHub, url, ErrFetch)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusUnauthorized {
		return nil, wrap("Poll", sourceGitHub, url, ErrFetch)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, wrap("Poll", sourceGitHub, url, ErrFetch)
	}

	var events []githubEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, wrap("Poll", sourceGitHub, url, ErrParse)
	}

	result := make([]Event, 0)
	for _, e := range events {
		if !e.CreatedAt.After(since) {
			continue
		}
		eventType := githubEventType(e.Type)
		if eventType == "" {
			continue
		}
		result = append(result, Event{
			URL:       url,
			Source:    sourceGitHub,
			EventType: eventType,
			Title:     fmt.Sprintf("%s on %s", e.Type, e.Repo.Name),
			Link:      url,
		})
	}
	return result, nil
}

func extractGitHubOwnerRepo(url string) (string, string, error) {
	trimmed := strings.TrimPrefix(url, "https://github.com/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidURL
	}
	return parts[0], parts[1], nil
}

func githubEventType(t string) string {
	switch t {
	case "PushEvent":
		return "push"
	case "PullRequestEvent":
		return "pr"
	case "IssuesEvent":
		return "issue"
	default:
		return ""
	}
}
