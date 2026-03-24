package poller

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

const sourceGitLab = "gitlab"

type gitlabEvent struct {
	ActionName  string    `json:"action_name"`
	CreatedAt   time.Time `json:"created_at"`
	ProjectID   int       `json:"project_id"`
	TargetType  string    `json:"target_type"`
	TargetTitle string    `json:"target_title"`
}

type GitLabPoller struct {
	client *http.Client
	token  string
	apiURL string
}

func NewGitLabPoller(token string) *GitLabPoller {
	return &GitLabPoller{
		client: &http.Client{Timeout: 10 * time.Second},
		token:  token,
		apiURL: "https://gitlab.com/api/v4",
	}
}

func (p *GitLabPoller) Source() string { return sourceGitLab }

func (p *GitLabPoller) Supports(url string) bool {
	return strings.HasPrefix(url, "https://gitlab.com/")
}

func (p *GitLabPoller) Poll(ctx context.Context, url string, since time.Time) ([]Event, error) {
	namespace, project, err := extractGitLabNamespaceProject(url)
	if err != nil {
		return nil, wrap("Poll", sourceGitLab, url, err)
	}

	encodedPath := strings.ReplaceAll(
		fmt.Sprintf("%s/%s", namespace, project),
		"/", "%2F",
	)

	apiURL := fmt.Sprintf(
		"%s/projects/%s/events?after=%s",
		p.apiURL, encodedPath, since.Format(time.RFC3339),
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return nil, wrap("Poll", sourceGitLab, url, ErrFetch)
	}

	req.Header.Set("Accept", "application/json")
	if p.token != "" {
		req.Header.Set("PRIVATE-TOKEN", p.token)
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, wrap("Poll", sourceGitLab, url, ErrFetch)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, wrap("Poll", sourceGitLab, url, ErrFetch)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, wrap("Poll", sourceGitLab, url, ErrFetch)
	}

	var events []gitlabEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, wrap("Poll", sourceGitLab, url, ErrParse)
	}

	result := make([]Event, 0)
	for _, e := range events {
		if !e.CreatedAt.After(since) {
			continue
		}
		eventType := gitlabEventType(e.ActionName, e.TargetType)
		if eventType == "" {
			continue
		}
		result = append(result, Event{
			URL:       url,
			Source:    sourceGitLab,
			EventType: eventType,
			Title:     fmt.Sprintf("%s: %s", e.ActionName, e.TargetTitle),
			Link:      url,
		})
	}
	return result, nil
}

func extractGitLabNamespaceProject(url string) (string, string, error) {
	trimmed := strings.TrimPrefix(url, "https://gitlab.com/")
	trimmed = strings.TrimSuffix(trimmed, "/")
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidURL
	}
	return parts[0], parts[1], nil
}

func gitlabEventType(action, targetType string) string {
	switch {
	case action == "pushed to" || action == "pushed new":
		return "push"
	case targetType == "MergeRequest":
		return "pr"
	case targetType == "Issue":
		return "issue"
	default:
		return ""
	}
}
