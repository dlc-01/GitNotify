package parser

import (
	"encoding/json"

	"github.com/dlc-01/GitNotify/internal/domain"
)

const sourceGitHub = "github"

var githubEventMap = map[string]domain.EventType{
	"push":         domain.EventPush,
	"pull_request": domain.EventPR,
	"issues":       domain.EventIssue,
	"pipeline":     domain.EventPipeline,
}

type GitHubParser struct{}

func NewGitHubParser() *GitHubParser {
	return &GitHubParser{}
}

func (p *GitHubParser) Source() string { return sourceGitHub }

func (p *GitHubParser) Parse(eventType string, payload []byte) (*Event, error) {
	if len(payload) == 0 {
		return nil, wrap("Parse", sourceGitHub, ErrEmptyPayload)
	}

	event, ok := githubEventMap[eventType]
	if !ok {
		return nil, wrap("Parse", sourceGitHub, ErrUnknownEvent)
	}

	repoURL, err := extractGitHubRepoURL(payload)
	if err != nil {
		return nil, wrap("Parse", sourceGitHub, err)
	}

	return &Event{
		RepoURL:   repoURL,
		EventType: event,
		Source:    sourceGitHub,
	}, nil
}

type githubPayload struct {
	Repository struct {
		HTMLURL string `json:"html_url"`
	} `json:"repository"`
}

func extractGitHubRepoURL(payload []byte) (string, error) {
	var p githubPayload
	if err := json.Unmarshal(payload, &p); err != nil {
		return "", ErrParsePayload
	}
	if p.Repository.HTMLURL == "" {
		return "", ErrMissingRepo
	}
	return p.Repository.HTMLURL, nil
}
