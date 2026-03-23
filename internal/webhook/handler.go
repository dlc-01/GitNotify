package webhook

import (
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/dlc-01/GitNotify/internal/domain"
	internalkafka "github.com/dlc-01/GitNotify/internal/kafka"
	"github.com/dlc-01/GitNotify/internal/kafka/producer"
	"github.com/dlc-01/GitNotify/internal/webhook/parser"
	"github.com/dlc-01/GitNotify/internal/webhook/validator"
)

type Handler struct {
	validators map[string]validator.Validator
	parsers    map[string]parser.Parser
	producer   producer.MultiProducer
	log        *slog.Logger
}

func NewHandler(p producer.MultiProducer, log *slog.Logger) *Handler {
	return &Handler{
		validators: make(map[string]validator.Validator),
		parsers:    make(map[string]parser.Parser),
		producer:   p,
		log:        log,
	}
}

func (h *Handler) RegisterValidator(v validator.Validator) {
	h.validators[v.Source()] = v
}

func (h *Handler) RegisterParser(p parser.Parser) {
	h.parsers[p.Source()] = p
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		h.log.Error("read body", slog.String("err", err.Error()))
		http.Error(w, "failed to read body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	headers := extractHeaders(r)
	source := detectSource(headers)

	if source == "" {
		h.log.Warn("unknown webhook source",
			slog.String("remote_addr", r.RemoteAddr),
		)
		http.Error(w, "unknown source", http.StatusBadRequest)
		return
	}

	v, ok := h.validators[source]
	if !ok {
		h.log.Warn("no validator for source",
			slog.String("source", source),
		)
		http.Error(w, "unknown source", http.StatusBadRequest)
		return
	}

	if err := v.Validate(payload, headers); err != nil {
		h.log.Warn("invalid webhook signature",
			slog.String("source", source),
			slog.String("err", err.Error()),
		)
		http.Error(w, "invalid signature", http.StatusUnauthorized)
		return
	}

	p, ok := h.parsers[source]
	if !ok {
		h.log.Warn("no parser for source", slog.String("source", source))
		http.Error(w, "unknown source", http.StatusBadRequest)
		return
	}

	eventType := detectEventType(source, headers)
	event, err := p.Parse(eventType, payload)
	if err != nil {
		h.log.Warn("failed to parse payload",
			slog.String("source", source),
			slog.String("event_type", eventType),
			slog.String("err", err.Error()),
		)
		http.Error(w, "failed to parse payload", http.StatusBadRequest)
		return
	}

	topic := eventToTopic(event.EventType)

	if err := h.producer.ProduceTo(r.Context(), topic, internalkafka.WebhookEventMessage{
		RepoURL:   event.RepoURL,
		EventType: string(event.EventType),
		Source:    event.Source,
	}); err != nil {
		h.log.Error("produce event",
			slog.String("source", source),
			slog.String("topic", string(topic)),
			slog.String("repo", event.RepoURL),
			slog.String("err", err.Error()),
		)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	h.log.Info("webhook event processed",
		slog.String("source", source),
		slog.String("event_type", eventType),
		slog.String("repo", event.RepoURL),
		slog.String("topic", string(topic)),
	)

	w.WriteHeader(http.StatusOK)
}

func extractHeaders(r *http.Request) map[string]string {
	headers := make(map[string]string)
	for k, v := range r.Header {
		if len(v) > 0 {
			headers[strings.ToLower(k)] = v[0]
		}
	}
	return headers
}

func detectSource(headers map[string]string) string {
	if _, ok := headers["x-github-event"]; ok {
		return "github"
	}
	if _, ok := headers["x-gitlab-event"]; ok {
		return "gitlab"
	}
	return ""
}

func detectEventType(source string, headers map[string]string) string {
	switch source {
	case "github":
		return headers["x-github-event"]
	case "gitlab":
		return headers["x-gitlab-event"]
	}
	return ""
}

func eventToTopic(event domain.EventType) internalkafka.Topic {
	switch event {
	case domain.EventPush:
		return internalkafka.TopicEventPush
	case domain.EventPR:
		return internalkafka.TopicEventPR
	case domain.EventIssue:
		return internalkafka.TopicEventIssue
	case domain.EventPipeline:
		return internalkafka.TopicEventPipeline
	case domain.EventAnswer:
		return internalkafka.TopicEventAnswer
	case domain.EventPost:
		return internalkafka.TopicEventPost
	case domain.EventVideo:
		return internalkafka.TopicEventVideo
	default:
		return internalkafka.TopicEventPush
	}
}
