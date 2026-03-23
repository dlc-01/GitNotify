package validator

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strings"
)

const sourceGitHub = "github"

type GitHubValidator struct {
	secret string
}

func NewGitHubValidator(secret string) *GitHubValidator {
	return &GitHubValidator{secret: secret}
}

func (v *GitHubValidator) Source() string { return sourceGitHub }

func (v *GitHubValidator) Validate(payload []byte, headers map[string]string) error {
	if v.secret == "" {
		return wrap("Validate", sourceGitHub, ErrEmptySecret)
	}

	signature, ok := headers["x-hub-signature-256"]
	if !ok {
		return wrap("Validate", sourceGitHub, ErrMissingHeader)
	}

	if !strings.HasPrefix(signature, "sha256=") {
		return wrap("Validate", sourceGitHub, ErrInvalidSignature)
	}

	expected := computeHMAC(payload, v.secret)
	actual := strings.TrimPrefix(signature, "sha256=")

	if !hmac.Equal([]byte(expected), []byte(actual)) {
		return wrap("Validate", sourceGitHub, ErrInvalidSignature)
	}

	return nil
}

func computeHMAC(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return hex.EncodeToString(mac.Sum(nil))
}
