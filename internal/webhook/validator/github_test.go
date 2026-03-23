package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func signPayload(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func TestGitHubValidator_Source(t *testing.T) {
	v := NewGitHubValidator("secret")
	assert.Equal(t, "github", v.Source())
}

func TestGitHubValidator_Validate_Success(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)
	v := NewGitHubValidator(secret)

	headers := map[string]string{
		"x-hub-signature-256": signPayload(payload, secret),
	}

	err := v.Validate(payload, headers)
	require.NoError(t, err)
}

func TestGitHubValidator_Validate_InvalidSignature(t *testing.T) {
	v := NewGitHubValidator("correct-secret")
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	headers := map[string]string{
		"x-hub-signature-256": signPayload(payload, "wrong-secret"),
	}

	err := v.Validate(payload, headers)
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrInvalidSignature)
}

func TestGitHubValidator_Validate_MissingHeader(t *testing.T) {
	v := NewGitHubValidator("secret")
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	err := v.Validate(payload, map[string]string{})
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrMissingHeader)
}

func TestGitHubValidator_Validate_EmptySecret(t *testing.T) {
	v := NewGitHubValidator("")
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)

	headers := map[string]string{
		"x-hub-signature-256": "sha256=abc",
	}

	err := v.Validate(payload, headers)
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrEmptySecret)
}

func TestGitHubValidator_Validate_MissingPrefix(t *testing.T) {
	secret := "test-secret"
	payload := []byte(`{"repository":{"html_url":"https://github.com/golang/go"}}`)
	v := NewGitHubValidator(secret)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	sig := hex.EncodeToString(mac.Sum(nil))

	headers := map[string]string{
		"x-hub-signature-256": sig,
	}

	err := v.Validate(payload, headers)
	require.Error(t, err)

	var validatorErr *Error
	require.ErrorAs(t, err, &validatorErr)
	assert.ErrorIs(t, validatorErr, ErrInvalidSignature)
}
