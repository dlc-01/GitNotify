package parser

import (
	"encoding/json"

	"github.com/dlc-01/GitNotify/internal/domain"
)

const sourceGitLab = "gitlab"

var gitlabEventMap = map[string]domain.EventType{
	"Push Hook":          domain.EventPush,
	"Merge Request Hook": domain.EventPR,
	"Issue Hook":         domain.EventIssue,
	"Pipeline Hook":      domain.EventPipeline,
}

type GitLabParser struct{}

func NewGitLabParser() *GitLabParser {
	return &GitLabParser{}
}

func (p *GitLabParser) Source() string { return sourceGitLab }

func (p *GitLabParser) Parse(eventType string, payload []byte) (*Event, error) {
	if len(payload) == 0 {
		return nil, wrap("Parse", sourceGitLab, ErrEmptyPayload)
	}

	event, ok := gitlabEventMap[eventType]
	if !ok {
		return nil, wrap("Parse", sourceGitLab, ErrUnknownEvent)
	}

	repoURL, err := extractGitLabRepoURL(payload)
	if err != nil {
		return nil, wrap("Parse", sourceGitLab, err)
	}

	return &Event{
		RepoURL:   repoURL,
		EventType: event,
		Source:    sourceGitLab,
	}, nil
}

type gitlabPayload struct {
	Project struct {
		WebURL string `json:"web_url"`
	} `json:"project"`
}

func extractGitLabRepoURL(payload []byte) (string, error) {
	var p gitlabPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return "", ErrParsePayload
	}
	if p.Project.WebURL == "" {
		return "", ErrMissingRepo
	}
	return p.Project.WebURL, nil
}
