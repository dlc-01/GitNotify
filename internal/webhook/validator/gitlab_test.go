package validator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGitLabValidator_Source(t *testing.T) {
	v := NewGitLabValidator("secret")
	assert.Equal(t, "gitlab", v.Source())
}

func TestGitLabValidator_Validate_Success(t *testing.T) {
	v := NewGitLabValidator("test-secret")
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	headers := map[string]string{
		"x-gitlab-token": "test-secret",
	}

	err := v.Validate(payload, headers)
	require.NoError(t, err)
}

func TestGitLabValidator_Validate_InvalidToken(t *testing.T) {
	v := NewGitLabValidator("correct-secret")
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	headers := map[string]string{
		"x-gitlab-token": "wrong-secret",
	}

	err := v.Validate(payload, headers)
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrInvalidToken)
}

func TestGitLabValidator_Validate_MissingHeader(t *testing.T) {
	v := NewGitLabValidator("secret")
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	err := v.Validate(payload, map[string]string{})
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrMissingHeader)
}

func TestGitLabValidator_Validate_EmptySecret(t *testing.T) {
	v := NewGitLabValidator("")
	payload := []byte(`{"project":{"web_url":"https://gitlab.com/user/repo"}}`)

	headers := map[string]string{
		"x-gitlab-token": "some-token",
	}

	err := v.Validate(payload, headers)
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrEmptySecret)
}
