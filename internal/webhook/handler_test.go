package webhook

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/webhook/parser"
	"github.com/dlc-01/GitNotify/internal/webhook/validator"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func newTestLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))
}

type mockMultiProducer struct {
	mock.Mock
}

func (m *mockMultiProducer) ProduceTo(ctx context.Context, topic internalkafka.Topic, msg any) error {
	args := m.Called(ctx, topic, msg)
	return args.Error(0)
}

func (m *mockMultiProducer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func signGitHub(payload []byte, secret string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}

func githubPayload(repoURL string) []byte {
	p := map[string]any{
		"repository": map[string]any{
			"html_url": repoURL,
		},
	}
	data, _ := json.Marshal(p)
	return data
}

func gitlabPayload(repoURL string) []byte {
	p := map[string]any{
		"project": map[string]any{
			"web_url": repoURL,
		},
	}
	data, _ := json.Marshal(p)
	return data
}

func TestHandler_GitHub_Push_Success(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	prod.On("ProduceTo", mock.Anything, internalkafka.TopicEventPush, mock.MatchedBy(func(msg internalkafka.WebhookEventMessage) bool {
		return msg.RepoURL == "https://github.com/golang/go" &&
			msg.EventType == string(domain.EventPush) &&
			msg.Source == "github"
	})).Return(nil)

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitHubValidator("secret"))
	h.RegisterParser(parser.NewGitHubParser())

	payload := githubPayload("https://github.com/golang/go")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signGitHub(payload, "secret"))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	prod.AssertExpectations(t)
}

func TestHandler_GitHub_InvalidSignature(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitHubValidator("secret"))
	h.RegisterParser(parser.NewGitHubParser())

	payload := githubPayload("https://github.com/golang/go")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signGitHub(payload, "wrong-secret"))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestHandler_GitLab_Push_Success(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	prod.On("ProduceTo", mock.Anything, internalkafka.TopicEventPush, mock.MatchedBy(func(msg internalkafka.WebhookEventMessage) bool {
		return msg.RepoURL == "https://gitlab.com/user/repo" &&
			msg.EventType == string(domain.EventPush) &&
			msg.Source == "gitlab"
	})).Return(nil)

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitLabValidator("secret"))
	h.RegisterParser(parser.NewGitLabParser())

	payload := gitlabPayload("https://gitlab.com/user/repo")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", "secret")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	prod.AssertExpectations(t)
}

func TestHandler_GitLab_InvalidToken(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitLabValidator("secret"))
	h.RegisterParser(parser.NewGitLabParser())

	payload := gitlabPayload("https://gitlab.com/user/repo")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", "wrong-secret")

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestHandler_MethodNotAllowed(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewHandler(prod, log)

	req := httptest.NewRequest(http.MethodGet, "/webhook", nil)
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestHandler_UnknownSource(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewHandler(prod, log)

	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte(`{}`)))
	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestHandler_GitHub_UnknownEvent(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitHubValidator("secret"))
	h.RegisterParser(parser.NewGitHubParser())

	payload := githubPayload("https://github.com/golang/go")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "unknown_event")
	req.Header.Set("X-Hub-Signature-256", signGitHub(payload, "secret"))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	prod.AssertNotCalled(t, "ProduceTo")
}

func TestHandler_ProduceError(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	prod.On("ProduceTo", mock.Anything, mock.Anything, mock.Anything).
		Return(assert.AnError)

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitHubValidator("secret"))
	h.RegisterParser(parser.NewGitHubParser())

	payload := githubPayload("https://github.com/golang/go")
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signGitHub(payload, "secret"))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	prod.AssertExpectations(t)
}

func TestHandler_EmptyBody(t *testing.T) {
	prod := &mockMultiProducer{}
	log := newTestLogger()

	h := NewHandler(prod, log)
	h.RegisterValidator(validator.NewGitHubValidator("secret"))
	h.RegisterParser(parser.NewGitHubParser())

	payload := []byte{}
	req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(payload))
	req.Header.Set("X-GitHub-Event", "push")
	req.Header.Set("X-Hub-Signature-256", signGitHub(payload, "secret"))

	w := httptest.NewRecorder()
	h.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	prod.AssertNotCalled(t, "ProduceTo")
}
