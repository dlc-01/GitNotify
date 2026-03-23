package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlc-01/GitNotify/internal/domain"
)

func TestGitHubParser_Source(t *testing.T) {
	p := NewGitHubParser()
	assert.Equal(t, "github", p.Source())
}

func TestGitHubParser_Parse_Push(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	event, err := p.Parse("push", payload)
	require.NoError(t, err)
	assert.Equal(t, "https://github.com/golang/go", event.RepoURL)
	assert.Equal(t, domain.EventPush, event.EventType)
	assert.Equal(t, "github", event.Source)
}

func TestGitHubParser_Parse_PullRequest(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	event, err := p.Parse("pull_request", payload)
	require.NoError(t, err)
	assert.Equal(t, domain.EventPR, event.EventType)
}

func TestGitHubParser_Parse_Issues(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	event, err := p.Parse("issues", payload)
	require.NoError(t, err)
	assert.Equal(t, domain.EventIssue, event.EventType)
}

func TestGitHubParser_Parse_Pipeline(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	event, err := p.Parse("pipeline", payload)
	require.NoError(t, err)
	assert.Equal(t, domain.EventPipeline, event.EventType)
}

func TestGitHubParser_Parse_EmptyPayload(t *testing.T) {
	p := NewGitHubParser()

	_, err := p.Parse("push", []byte{})
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrEmptyPayload)
}

func TestGitHubParser_Parse_UnknownEvent(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	_, err := p.Parse("unknown_event", payload)
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrUnknownEvent)
}

func TestGitHubParser_Parse_MissingRepo(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{"repository":{"html_url":""}}`)

	_, err := p.Parse("push", payload)
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrMissingRepo)
}

func TestGitHubParser_Parse_InvalidJSON(t *testing.T) {
	p := NewGitHubParser()
	payload := []byte(`{invalid json}`)

	_, err := p.Parse("push", payload)
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrParsePayload)
}
