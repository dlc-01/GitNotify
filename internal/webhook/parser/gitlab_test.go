package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dlc-01/GitNotify/internal/domain"
)

func TestGitLabParser_Source(t *testing.T) {
	p := NewGitLabParser()
	assert.Equal(t, "gitlab", p.Source())
}

func TestGitLabParser_Parse_Push(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	event, err := p.Parse("Push Hook", payload)
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com/user/repo", event.RepoURL)
	assert.Equal(t, domain.EventPush, event.EventType)
	assert.Equal(t, "gitlab", event.Source)
}

func TestGitLabParser_Parse_MergeRequest(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	event, err := p.Parse("Merge Request Hook", payload)
	require.NoError(t, err)
	assert.Equal(t, domain.EventPR, event.EventType)
}

func TestGitLabParser_Parse_Issue(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	event, err := p.Parse("Issue Hook", payload)
	require.NoError(t, err)
	assert.Equal(t, domain.EventIssue, event.EventType)
}

func TestGitLabParser_Parse_Pipeline(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	event, err := p.Parse("Pipeline Hook", payload)
	require.NoError(t, err)
	assert.Equal(t, domain.EventPipeline, event.EventType)
}

func TestGitLabParser_Parse_EmptyPayload(t *testing.T) {
	p := NewGitLabParser()

	_, err := p.Parse("Push Hook", []byte{})
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrEmptyPayload)
}

func TestGitLabParser_Parse_UnknownEvent(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	_, err := p.Parse("Unknown Hook", payload)
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrUnknownEvent)
}

func TestGitLabParser_Parse_MissingRepo(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{"project":{"web_url":""}}`)

	_, err := p.Parse("Push Hook", payload)
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrMissingRepo)
}

func TestGitLabParser_Parse_InvalidJSON(t *testing.T) {
	p := NewGitLabParser()
	payload := []byte(`{invalid json}`)

	_, err := p.Parse("Push Hook", payload)
	require.Error(t, err)

	var parserErr *Error
	require.ErrorAs(t, err, &parserErr)
	assert.ErrorIs(t, parserErr, ErrParsePayload)
}
